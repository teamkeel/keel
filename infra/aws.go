package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var cfg aws.Config

func init() {
	var err error
	cfg, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
}

func GetSecret(ctx context.Context, arn string) (string, error) {
	client := secretsmanager.NewFromConfig(cfg)

	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
	if err != nil {
		return "", err
	}

	if out.SecretString == nil {
		return "", errors.New("secret has no value")
	}

	return *out.SecretString, nil
}

func GetSecretJSON[T any](ctx context.Context, arn string) (*T, error) {
	value, err := GetSecret(ctx, arn)
	if err != nil {
		return nil, err
	}

	var dest T
	err = json.Unmarshal([]byte(value), &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

func GetS3File(ctx context.Context, bucket, key string) ([]byte, error) {
	client := s3.NewFromConfig(cfg)

	fmt.Println("reading S3 file", bucket, key)
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
