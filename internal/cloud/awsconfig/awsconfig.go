package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AssumeRoleOptions configures optional assume-role behavior.
type AssumeRoleOptions struct {
	RoleARN     string
	ExternalID  string
	SessionName string
}

// Options controls AWS SDK configuration loading.
type Options struct {
	Region               string
	AssumeRole           *AssumeRoleOptions
	CredentialsProvider  aws.CredentialsProvider
	SharedConfigProfile  string
	SharedConfigFilePath string
}

// Load builds an aws.Config using the default credential chain and optional assume-role.
func Load(ctx context.Context, opts Options) (aws.Config, error) {
	loadOptions := []func(*config.LoadOptions) error{}

	if opts.Region != "" {
		loadOptions = append(loadOptions, config.WithRegion(opts.Region))
	}
	if opts.CredentialsProvider != nil {
		loadOptions = append(loadOptions, config.WithCredentialsProvider(opts.CredentialsProvider))
	}
	if opts.SharedConfigProfile != "" {
		loadOptions = append(loadOptions, config.WithSharedConfigProfile(opts.SharedConfigProfile))
	}
	if opts.SharedConfigFilePath != "" {
		loadOptions = append(loadOptions, config.WithSharedConfigFiles([]string{opts.SharedConfigFilePath}))
	}

	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return aws.Config{}, err
	}

	if opts.AssumeRole != nil && opts.AssumeRole.RoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, opts.AssumeRole.RoleARN, func(o *stscreds.AssumeRoleOptions) {
			if opts.AssumeRole.ExternalID != "" {
				o.ExternalID = aws.String(opts.AssumeRole.ExternalID)
			}
			if opts.AssumeRole.SessionName != "" {
				o.RoleSessionName = opts.AssumeRole.SessionName
			}
		})
		cfg.Credentials = aws.NewCredentialsCache(provider)
	}

	return cfg, nil
}
