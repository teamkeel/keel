package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/dchest/uniuri"
	"github.com/teamkeel/keel/db"
)

const (
	// Character length of crypo-generated auth code.
	authCodeLength = 32
	authCodeExpiry = time.Duration(60) * time.Second
)

// NewAuthCode generates a new auth code for the identity using the
// configured or default expiry time.
func NewAuthCode(ctx context.Context, identityId string) (string, error) {
	ctx, span := tracer.Start(ctx, "New Auth Code")
	defer span.End()

	if identityId == "" {
		return "", errors.New("identity ID cannot be empty when generating new auth code")
	}

	code := uniuri.NewLen(authCodeLength)
	hash, err := hashToken(code)
	if err != nil {
		return "", err
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(authCodeExpiry)

	sql := `
		INSERT INTO 
			keel_auth_code (code, identity_id, expires_at, created_at) 
		VALUES 
			(?, ?, ?, ?)`

	db := database.GetDB().Exec(sql, hash, identityId, expiresAt, now)
	if db.Error != nil {
		return "", db.Error
	}

	if db.RowsAffected != 1 {
		return "", errors.New("failed to insert auth code token into database")
	}

	return code, nil
}

// ConsumeAuthCode checks that the provided auth code has not expired,
// consumes it (making it unusable again), and returning the identity it is associated with.
func ConsumeAuthCode(ctx context.Context, code string) (isValid bool, identityId string, err error) {
	ctx, span := tracer.Start(ctx, "Consume Auth Code")
	defer span.End()

	codeHash, err := hashToken(code)
	if err != nil {
		return false, "", err
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		return false, "", err
	}

	sql := `
		DELETE FROM 
			keel_auth_code
		WHERE 
			code = ? AND
			expires_at >= now()
		RETURNING 
			code, identity_id, expires_at, now()`

	rows := []map[string]any{}
	err = database.GetDB().Raw(sql, codeHash).Scan(&rows).Error
	if err != nil {
		return false, "", err
	}

	// There was no auth code found, and thus it is not valid
	if len(rows) != 1 {
		return false, "", nil
	}

	identityId, ok := rows[0]["identity_id"].(string)
	if !ok {
		return false, "", errors.New("could not parse identity_id from database result")
	}

	return true, identityId, nil
}
