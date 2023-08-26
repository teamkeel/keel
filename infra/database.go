package infra

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/sst-go"
)

var (
	databaseMutex sync.Mutex
	database      db.Database
)

func GetDatabase(ctx context.Context) (db.Database, error) {
	if database != nil {
		return database, nil
	}

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	// External database
	postgresUrl, err := sst.Secret(ctx, "POSTGRES_URL")
	if err == nil {
		database, err = db.New(ctx, postgresUrl)
		return database, err
	}

	// AWS database
	rds, err := sst.RDS(ctx, "Database")
	if err != nil {
		return nil, errors.New("missing both RDS and POSTGRES_URL secret - one of these must be set")
	}

	fmt.Println("Fetching RDS secret", rds.SecretArn)
	values, err := GetSecretJSON[DatabaseSecret](ctx, rds.SecretArn)
	if err != nil {
		return nil, err
	}

	conn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		values.Username, values.Password, values.DBName, values.Host, values.Port,
	)

	database, err = db.New(ctx, conn)
	return database, err
}

type DatabaseSecret struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}
