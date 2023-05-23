package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/db"
)

type dbContextKey string

var dbKey dbContextKey = "database"

func GetDatabase(ctx context.Context) (db.Database, error) {
	v := ctx.Value(dbKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", dbKey)
	}

	database, ok := v.(db.Database)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return database, nil
}

func WithDatabase(ctx context.Context, database db.Database) context.Context {
	return context.WithValue(ctx, dbKey, database)
}
