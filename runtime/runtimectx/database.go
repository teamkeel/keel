package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/db"
)

type dbContextKey string

var dbKey dbContextKey = "database"

func GetDatabase(ctx context.Context) (db.Db, error) {
	v := ctx.Value(dbKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", dbKey)
	}

	database, ok := v.(db.Db)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return database, nil
}

func WithDatabase(ctx context.Context, database db.Db) context.Context {
	return context.WithValue(ctx, dbKey, database)
}
