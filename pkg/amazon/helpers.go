package amazon

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/cobra"
)

// Options helper options for creating an AWS config
type Options struct {
	AWSProfile string `env:"AWS_PROFILE"`
	AWSRegion  string `env:"AWS_REGION"`
	Context    context.Context
	Config     *aws.Config
}

func (o *Options) GetConfig() (*aws.Config, error) {
	if o.Config != nil {
		return o.Config, nil
	}
	if o.Context == nil {
		o.Context = context.TODO()
	}

	var ops []func(*config.LoadOptions) error
	if o.AWSRegion != "" {
		ops = append(ops, config.WithRegion(o.AWSRegion))
	}
	log.Logger().Infof("loading config with AWS region: '%s'", o.AWSRegion)
	cfg, err := config.LoadDefaultConfig(o.Context, ops...)
	o.Config = &cfg
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create AWS config")
	}
	return o.Config, nil
}

// AddFlags adds the flags
func (o *Options) AddFlags(cmd *cobra.Command) {
	if o.Context == nil {
		o.Context = cmd.Context()
	}
	o.GetContext()
	cmd.Flags().StringVarP(&o.AWSProfile, "aws-profile", "", o.AWSProfile, "The AWS profile to use. Defaults to $AWS_PROFILE")
	cmd.Flags().StringVarP(&o.AWSRegion, "aws-region", "", o.AWSRegion, "The AWS region. Defaults to $AWS_REGION or its read from the 'jx-requirements.yml' for the development environment")
}

// EnvProcess processes the environment variable defaults
func (o *Options) EnvProcess() {
	err := envconfig.Process(o.GetContext(), o)
	if err != nil {
		log.Logger().Warnf("failed to default env vars: %s", err.Error())
	}
}

// GetContext returns the context, lazily creating one if required
func (o *Options) GetContext() context.Context {
	if o.Context == nil {
		o.Context = context.TODO()
	}
	return o.Context
}
