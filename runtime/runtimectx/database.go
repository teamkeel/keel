package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type dbContextKey string

var dbKey dbContextKey = "database"

func GetDatabase(ctx context.Context) (*gorm.DB, error) {
	v := ctx.Value(dbKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", dbKey)
	}
	db, ok := v.(*gorm.DB)
	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return db, nil
}

func WithDatabase(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}
