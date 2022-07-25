package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func GetDB(ctx context.Context) (*gorm.DB, error) {
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

func NewContext(db *gorm.DB) context.Context {
	return context.WithValue(context.Background(), dbKey, db)
}

const dbKey string = "database"
