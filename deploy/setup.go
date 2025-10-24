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
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
)

type AwsIdentityResult struct {
	AccountID string
	UserID    string
}

func getAwsIdentity(ctx context.Context, cfg aws.Config) (*AwsIdentityResult, error) {
	res, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		log(ctx, "%s error checking AWS auth: %s", IconCross, err.Error())
		return nil, err
	}

	return &AwsIdentityResult{
		AccountID: *res.Account,
		UserID:    *res.UserId,
	}, nil
}

type ResolveKeelConfigArgs struct {
	ProjectRoot string
	Env         string
}

func ResolveKeelConfig(ctx context.Context, args *ResolveKeelConfigArgs) (*config.ConfigFile, error) {
	t := NewTiming()

	configFiles, err := config.LoadAll(args.ProjectRoot)
	if err != nil {
		log(ctx, "%s Error loading Keel config files: %s", IconCross, err.Error())
		return nil, err
	}

	var baseConfig *config.ConfigFile
	var envConfig *config.ConfigFile

	for _, c := range configFiles {
		if c.Env == "" {
			baseConfig = c
		}
		if c.Env == args.Env {
			envConfig = c
		}
	}

	// Resolution logic is:
	// - If a keelconfig.<env>.yaml file exists, use that
	// - Else if a keelconfig.yaml file exists, use that
	// - Else use an empty config object
	c := envConfig
	if c == nil {
		c = baseConfig
	}
	if c == nil {
		c = &config.ConfigFile{
			Config: &config.ProjectConfig{},
		}
	}

	if c.Errors != nil {
		log(ctx, "%s Errors found in %s\n\n%s", IconCross, c.Filename, c.Errors.Error())
		return nil, c.Errors
	}

	log(ctx, "%s Using %s %s", IconTick, orange("%s", c.Filename), t.Since())

	return c, nil
}

type PulumiConfig struct {
	StackName        string
	WorkspaceOptions []auto.LocalWorkspaceOption
}

type SetupPulumiArgs struct {
	AwsConfig aws.Config
	Config    *config.ProjectConfig
	Env       string
}

func setupPulumi(ctx context.Context, args *SetupPulumiArgs) (*PulumiConfig, error) {
	// This is the bucket where Pulumi will store it's state for the stack
	bucketNameKey := fmt.Sprintf("/keel/%s/pulumi-bucket-name", args.Config.Deploy.ProjectName)

	// Pulumi requires a passphrase to encrypt secret values in the state file. We don't currently use this feature but
	// the passphrase is still required
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
		log(ctx, "%s error getting deploy config from SSM: %s", IconCross, err.Error())
		return nil, err
	}

	for _, p := range getParamsOut.Parameters {
		if p.Name == nil {
			continue
		}
		if *p.Name == bucketNameKey {
			bucketName = *p.Value
		}
		if *p.Name == passphraseKey {
			passphrase = *p.Value
		}
	}

	if bucketName == "" {
		t := NewTiming()
		v, err := randomString(hex.EncodeToString)
		if err != nil {
			log(ctx, "%s error generating Pulumi state bucket name: %s", IconCross, gray("%s", err.Error()))
			return nil, err
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
			log(ctx, "%s error creating Pulumi state bucket: %s", IconCross, gray("%s", err.Error()))
			return nil, err
		}

		_, err = ssm.NewFromConfig(args.AwsConfig).PutParameter(ctx, &ssm.PutParameterInput{
			Name:  &bucketNameKey,
			Value: aws.String(bucketName),
			Type:  ssmtypes.ParameterTypeString,
		})
		if err != nil {
			log(ctx, "%s error setting Pulumi state bucket name in SSM: %s", IconCross, gray("%s", err.Error()))
			return nil, err
		}

		log(ctx, "%s Pulumi state bucket created %s", IconTick, t.Since())
	}

	if passphrase == "" {
		t := NewTiming()
		value, err := randomString(base64.StdEncoding.EncodeToString)
		if err != nil {
			log(ctx, "%s error generating Pulumi passphrase: %s", IconCross, gray("%s", err.Error()))
			return nil, err
		}

		passphrase = value
		_, err = ssm.NewFromConfig(args.AwsConfig).PutParameter(ctx, &ssm.PutParameterInput{
			Name:  &passphraseKey,
			Value: aws.String(value),
			Type:  ssmtypes.ParameterTypeSecureString,
		})
		if err != nil {
			log(ctx, "%s error setting Pulumi passphrase in SSM: %s", IconCross, gray("%s", err.Error()))
			return nil, err
		}

		log(ctx, "%s Pulumi passphrase created %s", IconTick, t.Since())
	}

	t := NewTiming()
	pulumiCmd, err := auto.InstallPulumiCommand(ctx, &auto.PulumiCommandOptions{
		SkipVersionCheck: true,
	})
	if err != nil {
		log(ctx, "%s error installing Pulumi: %s", IconCross, gray(err.Error()))
		return nil, err
	}

	log(ctx, "%s Pulumi version %s installed %s", IconTick, orange("%s", pulumiCmd.Version().String()), t.Since())

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
				// Explicitly setting region here to make sure we use the one from config and not from the users environment
				"AWS_REGION":               args.Config.Deploy.Region,
				"PULUMI_CONFIG_PASSPHRASE": passphrase,
			},
		),
	}

	return result, nil
}

type CreatePrivateKeySecretArgs struct {
	AwsConfig   aws.Config
	ProjectName string
	Env         string
}

func createPrivateKeySecret(ctx context.Context, args *CreatePrivateKeySecretArgs) error {
	ssmClient := ssm.NewFromConfig(args.AwsConfig)

	privateKeyParamName := runtime.SsmParameterName(args.ProjectName, args.Env, "KEEL_PRIVATE_KEY")

	getParamResult, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(privateKeyParamName),
	})
	if err != nil {
		var ae smithy.APIError
		if !errors.As(err, &ae) || ae.ErrorCode() != "ParameterNotFound" {
			log(ctx, "%s error fetching private key secret from SSM: %s", IconCross, gray(err.Error()))
			return err
		}
	}

	if getParamResult != nil && getParamResult.Parameter != nil {
		log(ctx, "%s Using existing private key", IconTick)
		return nil
	}

	t := NewTiming()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log(ctx, "%s error generating private key: %s", IconCross, gray(err.Error()))
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
		log(ctx, "%s error setting private key secret in SSM: %s", IconCross, gray(err.Error()))
		return err
	}

	log(ctx, "%s New private key created %s", IconTick, t.Since())
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
