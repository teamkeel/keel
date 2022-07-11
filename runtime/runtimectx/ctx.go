package runtimectx

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func GetDB(ctx context.Context) *gorm.DB {
	v := ctx.Value(dbKey)
	if v == nil {
		panic(fmt.Sprintf("Context does not have key: %s", dbKey))
	}
	return v.(*gorm.DB)
}

func ContextWithDB(parent context.Context, db *gorm.DB) context.Context {
	return context.WithValue(parent, dbKey, db)
}

const dbKey string = "database"
