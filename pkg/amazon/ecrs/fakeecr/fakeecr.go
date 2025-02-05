package fakeecr

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/smithy-go/middleware"
)

// FakeECR a fake ECR implementation for testing
type FakeECR struct {
	Region       string
	Repositories map[string]*types.Repository
}

func (f *FakeECR) GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error) {
	return nil, &types.LifecyclePolicyNotFoundException{}
}

func (f *FakeECR) PutLifecyclePolicy(ctx context.Context, params *ecr.PutLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.PutLifecyclePolicyOutput, error) {
	repo := f.createRepo(*params.RepositoryName)
	text := "default"
	return &ecr.PutLifecyclePolicyOutput{
		LifecyclePolicyText: &text,
		RegistryId:          repo.RegistryId,
		RepositoryName:      repo.RepositoryName,
		ResultMetadata:      middleware.Metadata{},
	}, nil
}

func (f *FakeECR) GetRepositoryPolicy(ctx context.Context, params *ecr.GetRepositoryPolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error) {
	return nil, &types.RepositoryPolicyNotFoundException{}
}

func (f *FakeECR) SetRepositoryPolicy(ctx context.Context, params *ecr.SetRepositoryPolicyInput, optFns ...func(*ecr.Options)) (*ecr.SetRepositoryPolicyOutput, error) {
	repo := f.createRepo(*params.RepositoryName)
	text := "default"
	return &ecr.SetRepositoryPolicyOutput{
		PolicyText:     &text,
		RegistryId:     repo.RegistryId,
		RepositoryName: repo.RepositoryName,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (f *FakeECR) DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, opts ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	var repos []types.Repository
	{
	}
	if input != nil && f.Repositories != nil {
		for _, name := range input.RepositoryNames {
			r := f.Repositories[name]
			if r != nil {
				repos = append(repos, *r)
			}
		}
	}
	return &ecr.DescribeRepositoriesOutput{
		Repositories:   repos,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (f *FakeECR) CreateRepository(ctx context.Context, params *ecr.CreateRepositoryInput, opts ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	if params.RepositoryName == nil {
		return nil, fmt.Errorf("missing params.RepositoryName")
	}
	name := *params.RepositoryName
	if f.Repositories[name] != nil {
		return nil, fmt.Errorf("name %s already exists", name)
	}

	repo := f.createRepo(name)
	f.Repositories[name] = repo

	return &ecr.CreateRepositoryOutput{
		Repository:     repo,
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (f *FakeECR) createRepo(name string) *types.Repository {
	now := time.Now()
	if f.Region == "" {
		f.Region = "us-east-1"
	}
	id := "123456789012"
	arn := "arn:aws:ecr:" + f.Region + ":" + id + ":repository/" + name
	uri := id + ".dkr.ecr." + f.Region + ".amazonaws.com/" + name
	repo := &types.Repository{
		CreatedAt:      &now,
		RegistryId:     &id,
		RepositoryArn:  &arn,
		RepositoryName: &name,
		RepositoryUri:  &uri,
	}
	return repo
}

// NewFakeECR creates a new fake ECR
func NewFakeECR() *FakeECR {
	return &FakeECR{
		Repositories: map[string]*types.Repository{},
	}
}
