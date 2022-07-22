package runtimectx

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

func GetDB(ctx context.Context) *gorm.DB {
	v := ctx.Value(dbKey)
	if v == nil {
		// todo - if this remains connected after some more work on Create action, then
		// convert it to returning an error
		panic(fmt.Sprintf("Context does not have key: %s", dbKey))
	}
	return v.(*gorm.DB)
}

func GetSchema(ctx context.Context) *proto.Schema {
	v := ctx.Value(schemaKey)
	if v == nil {
		// todo - if this remains connected after some more work on Create action, then
		// convert it to returning an error
		panic(fmt.Sprintf("Context does not have key: %s", schemaKey))
	}
	return v.(*proto.Schema)
}

func WithDB(parent context.Context, db *gorm.DB) context.Context {
	return context.WithValue(parent, dbKey, db)
}

func WithSchema(parent context.Context, schema *proto.Schema) context.Context {
	return context.WithValue(parent, schemaKey, schema)
}

const dbKey string = "database"
const schemaKey = "schema"
