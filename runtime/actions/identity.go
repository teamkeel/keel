package actions

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"

	"github.com/iancoleman/strcase"

	"golang.org/x/crypto/bcrypt"
)

type Identity struct {
	Id       string `gorm:"column:id"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
}

type AuthenticateArgs struct {
	CreateIfNotExists bool
	Email             string
	Password          string
}

const (
	IdColumnName       string = "id"
	EmailColumnName    string = "email"
	PasswordColumnName string = "password"
)

// Authenticate will return the identity ID if it is successfully authenticated or when a new identity is created.
func Authenticate(ctx context.Context, schema *proto.Schema, args *AuthenticateArgs) (*ksuid.KSUID, bool, error) {
	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, false, err
	}

	identity, err := find(ctx, args.Email)

	if err != nil {
		return nil, false, err
	}

	if identity != nil {
		authenticated := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(args.Password)) == nil

		if authenticated {
			id, err := ksuid.Parse(identity.Id)

			if err != nil {
				return nil, false, err
			}

			return &id, false, nil
		} else {
			return nil, false, nil
		}
	} else if args.CreateIfNotExists {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

		if err != nil {
			return nil, false, err
		}

		identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

		modelMap, err := initialValueForModel(identityModel, schema)
		if err != nil {
			return nil, false, err
		}

		modelMap[strcase.ToSnake(EmailColumnName)] = args.Email
		modelMap[strcase.ToSnake(PasswordColumnName)] = string(hashedBytes)

		if err := db.Table(strcase.ToSnake(identityModel.Name)).Create(modelMap).Error; err != nil {
			return nil, false, err
		}

		id := modelMap[IdColumnName].(ksuid.KSUID)

		return &id, true, nil
	}

	return nil, false, nil
}

func find(ctx context.Context, email string) (*Identity, error) {
	db, _ := runtimectx.GetDB(ctx)

	var identity Identity

	result := db.Table(strcase.ToSnake(parser.ImplicitIdentityModelName)).Limit(1).Where(EmailColumnName, email).Find(&identity)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &identity, nil
}
