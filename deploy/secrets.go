package deploy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
)

type SetSecretArgs struct {
	ProjectRoot string
	Env         string
	Key         string
	Value       string
}

func SetSecret(ctx context.Context, args *SetSecretArgs) error {
	result, err := secretSetup(ctx, &SecretSetupArgs{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return err
	}

	c := result.config
	client := result.client
	name := result.name

	// We only allow setting secrets that are defined in the config file. This is because if the config file is valid
	// then the secrets listed are valid too. This avoids us doing any validation here on the secret name, if it's in the
	// config, then it's fine.
	if !lo.Contains(c.Config.AllSecrets(), args.Key) {
		log(ctx, "%s Secret %s not defined in %s", IconCross, orange(args.Key), orange(c.Filename))
		return fmt.Errorf("secret %s not defined in %s", args.Key, "")
	}

	_, err = client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(args.Value),
		Overwrite: aws.Bool(true),
		Type:      types.ParameterTypeSecureString,
	})
	if err != nil {
		log(ctx, "%s Error setting secret in AWS: %s", IconCross, gray(err.Error()))
		return err
	}

	return nil
}

type GetSecretArgs struct {
	ProjectRoot string
	Env         string
	Key         string
}

func GetSecret(ctx context.Context, args *GetSecretArgs) (*types.Parameter, error) {
	result, err := secretSetup(WithLogLevel(ctx, LogLevelSilent), &SecretSetupArgs{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return nil, err
	}

	client := result.client
	name := result.name

	p, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if !isSmithyAPIError(err, "ParameterNotFound") {
			log(ctx, "%s Error fetching secret %s from SSM: %s", IconCross, orange(args.Key), gray(err.Error()))
			return nil, err
		}
		log(ctx, "%s Secret %s not set", IconCross, orange(args.Key))
		return nil, err
	}

	return p.Parameter, nil
}

type ListSecretsArgs struct {
	ProjectRoot string
	Env         string
	Silent      bool
}

func ListSecrets(ctx context.Context, args *ListSecretsArgs) ([]types.Parameter, error) {
	result, err := secretSetup(ctx, &SecretSetupArgs{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         "",
	})
	if err != nil {
		return nil, err
	}

	// don't need final slash
	name := result.name
	path := strings.TrimSuffix(name, "/")

	client := result.client

	var token *string
	params := []types.Parameter{}
	for {
		p, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			WithDecryption: aws.Bool(true),
			NextToken:      token,
		})
		if err != nil {
			log(ctx, "%s Error listing secrets from AWS: %s", IconCross, gray(err.Error()))
			return nil, err
		}

		params = append(params, p.Parameters...)
		if p.NextToken != nil {
			token = p.NextToken
		}
		if token == nil {
			break
		}
	}

	return params, nil
}

type DeleteSecretArgs struct {
	ProjectRoot string
	Env         string
	Key         string
}

func DeleteSecret(ctx context.Context, args *DeleteSecretArgs) error {
	result, err := secretSetup(ctx, &SecretSetupArgs{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return err
	}

	client := result.client
	name := result.name

	_, err = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	if err != nil {
		if !isSmithyAPIError(err, "ParameterNotFound") {
			log(ctx, "%s Error deleting secret from SSM: %s", IconCross, gray(err.Error()))
			return err
		}
		log(ctx, "%s Secret %s not set", IconPipe, orange(args.Key))
		return nil
	}

	return nil
}

type SecretSetupArgs struct {
	projectRoot string
	env         string
	key         string
}

type SecretSetupResult struct {
	client *ssm.Client
	config *config.ConfigFile
	name   string
}

func secretSetup(ctx context.Context, args *SecretSetupArgs) (*SecretSetupResult, error) {
	conf, err := ResolveKeelConfig(ctx, &ResolveKeelConfigArgs{
		ProjectRoot: args.projectRoot,
		Env:         args.env,
	})
	if err != nil {
		return nil, err
	}

	if conf.Config.Deploy == nil {
		log(ctx, "%s Missing 'deploy' section in Keel config file %s", IconCross, orange(conf.Filename))
		return nil, errors.New("missing 'deploy' section in config file")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(conf.Config.Deploy.Region))
	if err != nil {
		log(ctx, "%s error loading AWS config: %s", IconCross, err.Error())
		return nil, err
	}

	paramName := runtime.SsmParameterName(conf.Config.Deploy.ProjectName, args.env, args.key)

	return &SecretSetupResult{
		client: ssm.NewFromConfig(cfg),
		config: conf,
		name:   paramName,
	}, nil
}
