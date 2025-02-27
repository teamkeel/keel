package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/teamkeel/keel/db"
)

func initDB(secrets map[string]string, dbEndpoint, dbName, secretName string) (db.Database, error) {
	ctx := context.Background()

	dbConnString := ""

	defer func() {
		// The functions-runtime expects a secret called KEEL_DB_CONN so we set that here
		// once we know what the value is. This is means we don't expose the connection
		// string as an env var to function and also don't need to re-implement secret fetching in JS.
		secrets["KEEL_DB_CONN"] = dbConnString
	}()

	// Try and get url from secrets (external database)
	var ok bool
	dbConnString, ok = secrets["DATABASE_URL"]
	if ok {
		return db.New(ctx, dbConnString)
	}

	// Otherwise use to secrets manager (RDS)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	dbConnString, err = GetRDSConnection(ctx, &GetRDSConnectionArgs{
		Cfg:       cfg,
		Endpoint:  dbEndpoint,
		DbName:    dbName,
		SecretArn: secretName,
	})
	if err != nil {
		panic(err)
	}

	return db.New(ctx, dbConnString)
}

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

	// Password may contain special characters to needs to be URL encoded
	encodedPassword := url.QueryEscape(secret.Password)

	return fmt.Sprintf("postgres://%s:%s@%s/%s", secret.Username, encodedPassword, args.Endpoint, args.DbName), nil
}
