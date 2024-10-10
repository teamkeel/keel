package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/deploy"
)

func initDB() {
	ctx := context.Background()

	defer func() {
		// TODO: this is needed for functions for now
		secrets["KEEL_DB_CONN"] = dbConnString
	}()

	var ok bool
	dbConnString, ok = secrets["DATABASE_URL"]
	if ok {
		var err error
		dbConn, err = db.New(ctx, dbConnString)
		if err != nil {
			panic(err)
		}
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	dbConnString, err = deploy.GetRDSConnection(ctx, &deploy.GetRDSConnectionArgs{
		Cfg:       cfg,
		Endpoint:  os.Getenv("KEEL_DATABASE_ENDPOINT"),
		DbName:    os.Getenv("KEEL_DATABASE_DB_NAME"),
		SecretArn: os.Getenv("KEEL_DATABASE_SECRET_ARN"),
	})
	if err != nil {
		panic(err)
	}

	dbConn, err = db.New(ctx, dbConnString)
	if err != nil {
		panic(err)
	}
}
