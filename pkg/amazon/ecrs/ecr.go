package ecrs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/jenkins-x-plugins/jx-registry/pkg/amazon"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/cobra"
)

var (
	defaultECRLifecyclePolicy = `
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Expire images older than 14 days",
            "selection": {
                "tagStatus": "tagged",
                "countType": "sinceImagePushed",
                "tagPrefixList": ["0.0.0-"],
                "countUnit": "days",
                "countNumber": 14
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
`
)

var (
	defaultECRRepositoryPolicy = `
	{
		"Version": "2008-10-17",
		"Statement": []
	  }
`
)

type ECRClient interface {
	DescribeRepositories(context.Context, *ecr.DescribeRepositoriesInput, ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	CreateRepository(ctx context.Context, params *ecr.CreateRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)
	GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error)
	PutLifecyclePolicy(ctx context.Context, params *ecr.PutLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.PutLifecyclePolicyOutput, error)
	GetRepositoryPolicy(ctx context.Context, params *ecr.GetRepositoryPolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error)
	SetRepositoryPolicy(ctx context.Context, params *ecr.SetRepositoryPolicyInput, optFns ...func(*ecr.Options)) (*ecr.SetRepositoryPolicyOutput, error)
}

type Options struct {
	amazon.Options
	RegistryID               string `env:"REGISTRY_ID"`
	Registry                 string `env:"DOCKER_REGISTRY"`
	RegistryOrganisation     string `env:"DOCKER_REGISTRY_ORG"`
	AppName                  string `env:"APP_NAME"`
	ECRLifecyclePolicy       string `env:"ECR_LIFECYCLE_POLICY"`
	ECRRepositoryPolicy      string `env:"ECR_REPOSITORY_POLICY"`
	CreateECRLifeCyclePolicy bool   `env:"CREATE_ECR_LIFECYCLE_POLICY,default=true"`
	CreateECRRepositoryPolicy bool   `env:"CREATE_ECR_REPOSITORY_POLICY,default=false"`
	ECRClient                ECRClient
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	o.Options.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.RegistryID, "registry-id", "", o.RegistryID, "The registry ID to use. If not specified finds the first path of the registry. $REGISTRY_ID")
	cmd.Flags().StringVarP(&o.Registry, "registry", "r", o.Registry, "The registry to use. Defaults to $DOCKER_REGISTRY")
	cmd.Flags().StringVarP(&o.RegistryOrganisation, "organisation", "o", o.RegistryOrganisation, "The registry organisation to use. Defaults to $DOCKER_REGISTRY_ORG")
	cmd.Flags().StringVarP(&o.AppName, "app", "a", o.AppName, "The app name to use. Defaults to $APP_NAME")
	cmd.Flags().StringVarP(&o.ECRLifecyclePolicy, "ecr-lifecycle-policy", "", o.ECRLifecyclePolicy, "ECR lifecycle policies to apply to the repository. Can be specified in $ECR_LIFECYCLE_POLICY.")
	cmd.Flags().StringVarP(&o.ECRRepositoryPolicy, "ecr-repository-policy", "", o.ECRRepositoryPolicy, "ECR repository policies to apply to the repository. Can be specified in $ECR_REPOSITORY_POLICY.")	
	cmd.Flags().BoolVarP(&o.CreateECRLifeCyclePolicy, "create-ecr-lifecycle-policy", "", o.CreateECRLifeCyclePolicy, "Should ECR Lifecycle Policy be created. Can be specified in $CREATE_ECR_LIFECYCLE_POLICY.")
	cmd.Flags().BoolVarP(&o.CreateECRRepositoryPolicy, "create-ecr-repository-policy", "", o.CreateECRRepositoryPolicy, "Should ECR Repository Policy be created. Can be specified in $CREATE_ECR_REPOSITORY_POLICY.")
}

func (o *Options) Validate() error {
	cfg, err := o.GetConfig()
	if err != nil {
		return errors.Wrapf(err, "failed to create AWS config")
	}
	if cfg == nil {
		return errors.Errorf("no AWS config")
	}
	return nil
}

// EnvProcess processes the environment variable defaults
func (o *Options) EnvProcess() {
	nilCfg := o.Config == nil
	err := envconfig.Process(o.GetContext(), o)
	if err != nil {
		log.Logger().Warnf("failed to default env vars: %s", err.Error())
	}
	// lets avoid an empty config being created by the envconfig
	if nilCfg {
		o.Config = nil
	}
}

// LazyCreateRegistry lazily creates the ECR registry if it does not already exist
func (o *Options) LazyCreateRegistry(appName string) error {
	ctx := o.GetContext()
	cfg, err := o.GetConfig()
	if err != nil {
		return errors.Wrapf(err, "failed to create the AWS configuration")
	}
	if cfg == nil {
		return errors.Errorf("no AWS configuration could be found")
	}
	if len(appName) <= 2 {
		return errors.Errorf("missing valid app name: '%s'", appName)
	}

	region := o.AWSRegion
	if region == "" {
		return options.MissingOption("aws-region")
	}

	// strip any tag/version from the app name
	idx := strings.Index(appName, ":")
	if idx > 0 {
		appName = appName[0:idx]
	}
	repoName := appName
	if o.RegistryOrganisation != "" {
		repoName = o.RegistryOrganisation + "/" + appName
	}
	repoName = strings.ToLower(repoName)
	log.Logger().Infof("Let's ensure that we have an ECR repository for the image %s", termcolor.ColorInfo(repoName))

	if o.ECRClient == nil {
		o.ECRClient = ecr.NewFromConfig(*cfg)
	}
	svc := o.ECRClient

	repoInput := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	}
	if o.RegistryID != "" {
		repoInput.RegistryId = &o.RegistryID
	}
	result, err := svc.DescribeRepositories(ctx, repoInput)
	if err != nil {
		var notFoundErr *types.RepositoryNotFoundException
		if !errors.As(err, &notFoundErr) {
			return errors.Wrapf(err, "failed to check for repository with registry ID %s", o.RegistryID)
		}
	}
	if result != nil {
		for _, repo := range result.Repositories {
			if repo.RepositoryName == nil {
				continue
			}
			name := *repo.RepositoryName
			log.Logger().Infof("Found repository: %s", name)
			if name == repoName {
				return o.EnsureLifecyclePolicy(repoName)
			}
		}
	}
	createRepoInput := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	}
	createResult, err := svc.CreateRepository(ctx, createRepoInput)
	if err != nil {
		return fmt.Errorf("Failed to create the ECR repository for %s due to: %s", repoName, err)
	}
	repo := createResult.Repository
	if repo != nil {
		u := repo.RepositoryUri
		if u != nil {
			log.Logger().Infof("Created ECR repository: %s", termcolor.ColorInfo(*u))
		}
	}
	return o.EnsureLifecyclePolicy(repoName)
}

func (o *Options) EnsureLifecyclePolicy(repoName string) error {
	if o.CreateECRLifeCyclePolicy {
		client := o.ECRClient
		ctx := o.GetContext()

		getLifecyclePolicyInput := &ecr.GetLifecyclePolicyInput{
			RepositoryName: aws.String(repoName),
		}
		if o.RegistryID != "" {
			getLifecyclePolicyInput.RegistryId = &o.RegistryID
		}
		getLifecyclePolicyOutput, err := client.GetLifecyclePolicy(ctx, getLifecyclePolicyInput)
		if err == nil && o.ECRLifecyclePolicy == "" {
			// Won't overwrite existing lifecycle policy if no policy has been specified
			return nil
		}
		if err != nil {
			var notFoundErr *types.LifecyclePolicyNotFoundException
			if !errors.As(err, &notFoundErr) {
				// LifecyclePolicyNotFoundException is OK since we then create it below
				return fmt.Errorf("Failed to fetch lifecycle policy for the ECR repository %s due to: %s",
					repoName, err)
			}
		}
		if o.ECRLifecyclePolicy == "" {
			o.ECRLifecyclePolicy = defaultECRLifecyclePolicy
		}
		if err == nil && o.ECRLifecyclePolicy == *getLifecyclePolicyOutput.LifecyclePolicyText {
			// No need to put policy if it already set. I'm not sure
			return nil
		}
		putLifecyclePolicyInput := &ecr.PutLifecyclePolicyInput{
			LifecyclePolicyText: aws.String(o.ECRLifecyclePolicy),
			RepositoryName:      aws.String(repoName),
		}
		if o.RegistryID != "" {
			putLifecyclePolicyInput.RegistryId = &o.RegistryID
		}
		putLifecyclePolicyOutput, err := client.PutLifecyclePolicy(ctx, putLifecyclePolicyInput)
		if err != nil {
			return fmt.Errorf("Failed to put lifecycle policy '%s' for the ECR repository %s due to: %s",
				o.ECRLifecyclePolicy, repoName, err)
		}
		log.Logger().Infof("Put ECR repository lifecycle policy: %s", termcolor.ColorInfo(*(*putLifecyclePolicyOutput).LifecyclePolicyText))
	}
	return nil
}

func (o *Options) EnsureRepositoryPolicy(repoName string) error {
	if o.CreateECRRepositoryPolicy {
		client := o.ECRClient
		ctx := o.GetContext()

		getRepositoryPolicyInput := &ecr.GetRepositoryPolicyInput{
			RepositoryName: aws.String(repoName),
		}
		if o.RegistryID != "" {
			getRepositoryPolicyInput.RegistryId = &o.RegistryID
		}
		getRepositoryPolicyOutput, err := client.GetRepositoryPolicy(ctx, getRepositoryPolicyInput)
		if err == nil && o.ECRRepositoryPolicy == "" {
			// Won't overwrite existing lifecycle policy if no policy has been specified
			return nil
		}
		if err != nil {
			var notFoundErr *types.LifecyclePolicyNotFoundException
			if !errors.As(err, &notFoundErr) {
				// LifecyclePolicyNotFoundException is OK since we then create it below
				return fmt.Errorf("Failed to fetch lifecycle policy for the ECR repository %s due to: %s",
					repoName, err)
			}
		}
		if o.ECRRepositoryPolicy == "" {
			o.ECRRepositoryPolicy = defaultECRRepositoryPolicy
		}
		if err == nil && o.ECRRepositoryPolicy == *getRepositoryPolicyOutput.PolicyText {			
			// No need to put policy if it already set. I'm not sure
			return nil
		}
		setRepositoryPolicyInput := &ecr.SetRepositoryPolicyInput{			
			PolicyText: aws.String(o.ECRRepositoryPolicy),
			RepositoryName:      aws.String(repoName),
		}
		if o.RegistryID != "" {
			setRepositoryPolicyInput.RegistryId = &o.RegistryID
		}
		setRegistryPolicyOutput, err := client.SetRepositoryPolicy(ctx, setRepositoryPolicyInput)
		if err != nil {
			return fmt.Errorf("Failed to set repository policy '%s' for the ECR repository %s due to: %s",
				o.ECRRepositoryPolicy, repoName, err)
		}
		log.Logger().Infof("Put ECR repository repository policy: %s", termcolor.ColorInfo(*(*setRegistryPolicyOutput).PolicyText))
	}
	return nil
}
