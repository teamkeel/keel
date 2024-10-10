package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/deploy"
)

var secrets = map[string]string{}

func initSecrets() {
	secretNames := strings.Split(os.Getenv("KEEL_SECRETS"), ":")

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	res, err := ssm.NewFromConfig(cfg).GetParameters(context.Background(), &ssm.GetParametersInput{
		Names: lo.Map(secretNames, func(s string, _ int) string {
			return deploy.SsmParameterName(os.Getenv("KEEL_PROJECT_NAME"), os.Getenv("KEEL_ENV"), s)
		}),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	for _, p := range res.Parameters {
		parts := strings.Split(*p.Name, "/")
		name := parts[len(parts)-1]
		secrets[name] = *p.Value
	}
}
