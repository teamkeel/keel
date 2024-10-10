package deploy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/teamkeel/keel/config"
)

type Output struct {
	Heading string
	Icon    string
	Message string
	Error   error
}

const (
	OutputIconTick  = "tick"
	OutputIconCross = "cross"
	OutputIconPipe  = "pipe"
)

type AwsIdentityResult struct {
	AccountID string
	UserID    string
}

func getAwsIdentity(ctx context.Context, cfg aws.Config, events chan Output) (*AwsIdentityResult, error) {
	res, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		events <- Output{
			Icon:    OutputIconCross,
			Message: "Error checking AWS auth",
			Error:   err,
		}
		return nil, err
	}

	events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Account ID: %s", *res.Account),
	}
	events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("UserID ID: %s", *res.UserId),
	}

	return &AwsIdentityResult{
		AccountID: *res.Account,
		UserID:    *res.UserId,
	}, nil
}

func loadKeelConfig(args *RunArgs) (*config.ProjectConfig, error) {
	configFiles, err := config.LoadAll(args.ProjectRoot)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error loading Keel config",
			Error:   err,
		}
		return nil, err
	}

	var configFile *config.ConfigFile

	for _, c := range configFiles {
		if c.Env == args.Env {
			configFile = c
		}
	}

	if configFile == nil {
		err = fmt.Errorf("no keelconfig.%s.yaml found", args.Env)
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: fmt.Sprintf("no keelconfig.%s.yaml found", args.Env),
		}
		return nil, err
	}

	if configFile.Errors != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: fmt.Sprintf("Errors found in keelconfig.%s.yaml", args.Env),
			Error:   configFile.Errors,
		}
		return nil, configFile.Errors
	}

	deploy := configFile.Config.Deploy
	if deploy == nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: fmt.Sprintf("No deploy config found in keelconfig.%s.yaml", args.Env),
		}
		return nil, fmt.Errorf("missing deploy config in keelconfig.%s.yaml", args.Env)
	}

	config := configFile.Config

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Config file: %s", configFile.Filename),
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Project: %s", config.Deploy.ProjectName),
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Region: %s", config.Deploy.Region),
	}

	return config, nil
}

type PulumiConfig struct {
	StackName        string
	WorkspaceOptions []auto.LocalWorkspaceOption
}

type SetupPulumiArgs struct {
	AwsConfig aws.Config
	Config    *config.ProjectConfig
	Env       string
	Events    chan Output
}

func setupPulumi(ctx context.Context, args *SetupPulumiArgs) *PulumiConfig {
	bucketNameKey := fmt.Sprintf("/keel/%s/pulumi-bucket-name", args.Config.Deploy.ProjectName)
	passphraseKey := fmt.Sprintf("/keel/%s/pulumi-passphrase", args.Config.Deploy.ProjectName)

	bucketName := ""
	passphrase := ""

	result := &PulumiConfig{
		StackName: fmt.Sprintf("keel-%s-%s", args.Config.Deploy.ProjectName, args.Env),
	}

	getParamsOut, err := ssm.NewFromConfig(args.AwsConfig).GetParameters(ctx, &ssm.GetParametersInput{
		Names: []string{
			bucketNameKey,
			passphraseKey,
		},
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error getting deploy config from SSM",
			Error:   err,
		}
		return nil
	}

	for _, p := range getParamsOut.Parameters {
		if p.Name == nil {
			continue
		}
		if *p.Name == bucketNameKey {
			args.Events <- Output{
				Icon:    OutputIconTick,
				Message: fmt.Sprintf("S3 state bucket: %s", *p.Value),
			}
			bucketName = *p.Value
		}
		if *p.Name == passphraseKey {
			args.Events <- Output{
				Icon:    OutputIconTick,
				Message: "Retreived stored passphrase",
			}
			passphrase = *p.Value
		}
	}

	if bucketName == "" {
		v, err := randomString(hex.EncodeToString)
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error generating state bucket name",
				Error:   err,
			}
			return nil
		}

		// Very annoying bit of the AWS SDK. If the region is us-east-1, then CreateBucketConfiguration must be set to nil
		var bucketConfig *types.CreateBucketConfiguration
		if args.Config.Deploy.Region != "us-east-1" {
			bucketConfig = &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraint(args.Config.Deploy.Region)}
		}

		bucketName = fmt.Sprintf("keel-%s-%s", args.Config.Deploy.ProjectName, strings.ToLower(v[:7]))
		_, err = s3.NewFromConfig(args.AwsConfig).CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket:                    aws.String(bucketName),
			CreateBucketConfiguration: bucketConfig,
		})
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error creating Pulumi state bucket",
				Error:   err,
			}
			return nil
		}

		_, err = ssm.NewFromConfig(args.AwsConfig).PutParameter(ctx, &ssm.PutParameterInput{
			Name:  &bucketNameKey,
			Value: aws.String(bucketName),
			Type:  ssmtypes.ParameterTypeString,
		})
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error storing bucket name in SSM",
				Error:   err,
			}
			return nil
		}

		args.Events <- Output{
			Icon:    OutputIconTick,
			Message: fmt.Sprintf("Created Pulumi state bucket: %s", bucketName),
		}
	}

	if passphrase == "" {
		value, err := randomString(base64.StdEncoding.EncodeToString)
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error generating passphrase",
				Error:   err,
			}
			return nil
		}

		passphrase = value
		_, err = ssm.NewFromConfig(args.AwsConfig).PutParameter(ctx, &ssm.PutParameterInput{
			Name:  &passphraseKey,
			Value: aws.String(value),
			Type:  ssmtypes.ParameterTypeSecureString,
		})
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error storing passphrase in SSM",
				Error:   err,
			}
			return nil
		}

		args.Events <- Output{
			Icon:    OutputIconTick,
			Message: "Passphrase created",
		}
	}

	pulumiCmd, err := auto.InstallPulumiCommand(ctx, &auto.PulumiCommandOptions{
		SkipVersionCheck: true,
	})
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error installing the Pulumi CLI",
			Error:   err,
		}
		return nil
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Using Pulumi CLI v%s", pulumiCmd.Version().String()),
	}

	result.WorkspaceOptions = []auto.LocalWorkspaceOption{
		auto.Pulumi(pulumiCmd),
		auto.Project(workspace.Project{
			Name:    tokens.PackageName(args.Config.Deploy.ProjectName),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{
				URL: fmt.Sprintf("s3://%s", bucketName),
			},
		}),
		auto.EnvVars(
			map[string]string{
				"PULUMI_CONFIG_PASSPHRASE": passphrase,
			},
		),
	}

	return result
}

type CreatePrivateKeySecretArgs struct {
	AwsConfig   aws.Config
	Events      chan Output
	ProjectName string
	Env         string
}

func createPrivateKeySecret(ctx context.Context, args *CreatePrivateKeySecretArgs) error {
	ssmClient := ssm.NewFromConfig(args.AwsConfig)

	privateKeyParamName := SsmParameterName(args.ProjectName, args.Env, "KEEL_PRIVATE_KEY")

	getParamResult, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(privateKeyParamName),
	})
	if err != nil {
		var ae smithy.APIError
		if !errors.As(err, &ae) || ae.ErrorCode() != "ParameterNotFound" {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error checking for private key secret",
				Error:   err,
			}
			return err
		}
	}

	if getParamResult != nil && getParamResult.Parameter != nil {
		args.Events <- Output{
			Icon:    OutputIconTick,
			Message: fmt.Sprintf("Existing private key found for env %s", args.Env),
		}
		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error generating private key",
			Error:   err,
		}
		return err
	}

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	_, err = ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(privateKeyParamName),
		Value: aws.String(string(privateKeyPem)),
		Type:  ssmtypes.ParameterTypeSecureString,
	})
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error storing private key in SSM",
			Error:   err,
		}
		return err
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Generated and saved private key for env %s", args.Env),
	}
	return nil
}

func randomString(encode func([]byte) string) (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return encode(bytes), nil
}
