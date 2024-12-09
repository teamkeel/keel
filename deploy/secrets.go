package deploy

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
)

type SetSecretArgs struct {
	ProjectRoot string
	Env         string
	Key         string
	Value       string
}

func SetSecret(ctx context.Context, args *SetSecretArgs) error {
	client, name, err := getSsmClient(ctx, &getSssmClient{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return err
	}

	_, err = client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(args.Value),
		Overwrite: aws.Bool(true),
		Type:      types.ParameterTypeSecureString,
	})
	if err != nil {
		log("%s errors setting secret in AWS: %s", IconCross, err.Error())
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
	client, name, err := getSsmClient(ctx, &getSssmClient{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return nil, err
	}

	p, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if !isSmithyAPIError(err, "ParameterNotFound") {
			log("%s error fetching secret from SSM: %s", IconCross, err.Error())
			return nil, err
		}
		log("%s secret '%s' not set", IconCross, args.Key)
		return nil, err
	}

	log(*p.Parameter.Value)
	return p.Parameter, nil
}

type ListSecretsArgs struct {
	ProjectRoot string
	Env         string
}

func ListSecrets(ctx context.Context, args *ListSecretsArgs) ([]types.Parameter, error) {
	client, name, err := getSsmClient(ctx, &getSssmClient{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         "",
	})
	if err != nil {
		return nil, err
	}

	// don't need final slash
	path := strings.TrimSuffix(name, "/")

	var token *string
	params := []types.Parameter{}
	for {
		p, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			WithDecryption: aws.Bool(true),
			NextToken:      token,
		})
		if err != nil {
			log("%s error listing secrets from AWS: %s", IconCross, err.Error())
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
	client, name, err := getSsmClient(ctx, &getSssmClient{
		projectRoot: args.ProjectRoot,
		env:         args.Env,
		key:         args.Key,
	})
	if err != nil {
		return err
	}

	_, err = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	if err != nil {
		if !isSmithyAPIError(err, "ParameterNotFound") {
			log("%s error deleting secret from SSM: %s", IconCross, err.Error())
			return err
		}
		log("%s secret '%s' not set", IconPipe, args.Key)
		return nil
	}

	return nil
}

type getSssmClient struct {
	projectRoot string
	env         string
	key         string
}

func getSsmClient(ctx context.Context, args *getSssmClient) (client *ssm.Client, name string, err error) {
	projectConfig, err := loadKeelConfig(&LoadKeelConfigArgs{
		ProjectRoot: args.projectRoot,
		Env:         args.env,
	})
	if err != nil {
		return nil, "", err
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(projectConfig.Deploy.Region))
	if err != nil {
		log("%s error loading AWS config: %s", IconCross, err.Error())
		return nil, "", err
	}

	paramName := runtime.SsmParameterName(projectConfig.Deploy.ProjectName, args.env, args.key)

	return ssm.NewFromConfig(cfg), paramName, nil
}
