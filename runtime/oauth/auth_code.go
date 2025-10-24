package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	return NewAuthCodeWithPKCE(ctx, identityId, "", "", "")
}

// NewAuthCodeWithPKCE generates a new auth code with PKCE support
func NewAuthCodeWithPKCE(ctx context.Context, identityId string, codeChallenge string, codeChallengeMethod string, resource string) (string, error) {
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
			keel_auth_code (code, identity_id, expires_at, created_at, code_challenge, code_challenge_method, resource)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)`

	db := database.GetDB().Exec(sql, hash, identityId, expiresAt, now, codeChallenge, codeChallengeMethod, resource)
	if db.Error != nil {
		return "", db.Error
	}

	if db.RowsAffected != 1 {
		return "", errors.New("failed to insert auth code token into database")
	}

	return code, nil
}

// AuthCodeData contains all the data associated with an auth code
type AuthCodeData struct {
	IdentityID          string
	CodeChallenge       *string
	CodeChallengeMethod *string
	Resource            *string
}

// ConsumeAuthCode checks that the provided auth code has not expired,
// consumes it (making it unusable again), and returning the identity it is associated with.
func ConsumeAuthCode(ctx context.Context, code string) (isValid bool, identityId string, err error) {
	data, valid, err := ConsumeAuthCodeWithPKCE(ctx, code)
	if err != nil || !valid {
		return valid, "", err
	}
	return true, data.IdentityID, nil
}

// ConsumeAuthCodeWithPKCE checks that the provided auth code has not expired,
// consumes it, and returns all associated data including PKCE parameters
func ConsumeAuthCodeWithPKCE(ctx context.Context, code string) (*AuthCodeData, bool, error) {
	ctx, span := tracer.Start(ctx, "Consume Auth Code")
	defer span.End()

	codeHash, err := hashToken(code)
	if err != nil {
		return nil, false, err
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, false, err
	}

	sql := `
		DELETE FROM
			keel_auth_code
		WHERE
			code = ? AND
			expires_at >= now()
		RETURNING
			code, identity_id, code_challenge, code_challenge_method, resource, expires_at, now()`

	rows := []map[string]any{}
	err = database.GetDB().Raw(sql, codeHash).Scan(&rows).Error
	if err != nil {
		return nil, false, err
	}

	// There was no auth code found, and thus it is not valid
	if len(rows) != 1 {
		return nil, false, nil
	}

	identityId, ok := rows[0]["identity_id"].(string)
	if !ok {
		return nil, false, errors.New("could not parse identity_id from database result")
	}

	data := &AuthCodeData{
		IdentityID: identityId,
	}

	if codeChallenge, ok := rows[0]["code_challenge"].(string); ok && codeChallenge != "" {
		data.CodeChallenge = &codeChallenge
	}

	if codeChallengeMethod, ok := rows[0]["code_challenge_method"].(string); ok && codeChallengeMethod != "" {
		data.CodeChallengeMethod = &codeChallengeMethod
	}

	if resource, ok := rows[0]["resource"].(string); ok && resource != "" {
		data.Resource = &resource
	}

	return data, true, nil
}

// ValidatePKCE validates a code_verifier against a code_challenge using the S256 method
func ValidatePKCE(codeVerifier string, codeChallenge string, method string) bool {
	if method != "S256" {
		// Only S256 is supported per MCP spec
		return false
	}

	// Calculate S256 challenge: BASE64URL(SHA256(ASCII(code_verifier)))
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return expectedChallenge == codeChallenge
}
