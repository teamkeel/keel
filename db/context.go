package db

import (
	"context"
	"errors"
	"fmt"
)

type dbContextKey string

var dbKey dbContextKey = "database"

func GetDatabase(ctx context.Context) (Database, error) {
	v := ctx.Value(dbKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", dbKey)
	}

	database, ok := v.(Database)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return database, nil
}

func WithDatabase(ctx context.Context, database Database) context.Context {
	database = database.WithAuditing()
	return context.WithValue(ctx, dbKey, database)
}
