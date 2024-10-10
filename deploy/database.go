package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type GetRDSConnectionArgs struct {
	Cfg       aws.Config
	Endpoint  string
	DbName    string
	SecretArn string
}

func GetRDSConnection(ctx context.Context, args *GetRDSConnectionArgs) (string, error) {
	r, err := secretsmanager.NewFromConfig(args.Cfg).GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(args.SecretArn),
	})
	if err != nil {
		return "", err
	}

	type Secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var secret Secret
	err = json.Unmarshal([]byte(*r.SecretString), &secret)
	if err != nil {
		return "", err
	}

	encodedPassword := url.QueryEscape(secret.Password)
	return fmt.Sprintf("postgres://%s:%s@%s/%s", secret.Username, encodedPassword, args.Endpoint, args.DbName), nil
}
