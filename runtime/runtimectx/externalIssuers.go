package runtimectx

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

type externalIssuersKey string

var externalIssuersKeyContext externalIssuersKey = "externalIssuers"

func WithExternalIssuers(ctx context.Context, issuers map[string]*rsa.PublicKey) context.Context {
	return context.WithValue(ctx, externalIssuersKeyContext, issuers)
}

const (
	ExternalIssuersEnvKey = "KEEL_EXTERNAL_ISSUERS"
)

type Jwks struct {
	Set jwk.Set
}

func NewJwks(uri string) (*Jwks, error) {
	jwksUri, err := url.Parse(fmt.Sprintf("%s/.well-known/jwks.json", strings.TrimSuffix(uri, "/")))

	if err != nil {
		return nil, err
	}

	keyset, err := jwk.Fetch(context.Background(), jwksUri.String())

	if err != nil {
		return nil, err
	}

	return &Jwks{
		Set: keyset,
	}, nil
}

func (j *Jwks) PublicKey(tokenKid string) (*rsa.PublicKey, error) {
	allKeys := j.Set.Keys(context.Background())

	found := false

	var publicKey rsa.PublicKey

	for allKeys.Next(context.Background()) {
		curr := allKeys.Pair()

		switch v := curr.Value.(type) {
		case jwk.RSAPublicKey:
			kid := v.KeyID()

			if tokenKid == kid {
				found = true

				err := v.Raw(&publicKey)

				if err != nil {
					found = false

				}

				if found {
					break
				}
			}
		}

	}

	if !found {
		return nil, errors.New("No RSA public key found")
	}

	return &publicKey, nil
}

// ExternalIssuersFromEnv is responsible for parsing the contents of the KEEL_EXTERNAL_ISSUERS environment variable. This environment variable is a comma separated list of authorization server uris. For every value in the csv, it is assumed that the target host will expose an endpoint at /.well-known/jwks.json that contains a list of public keys within it. Any value that is not a valid URI will be ignored.
func ExternalIssuersFromEnv() (providers []string) {
	envVar := os.Getenv(ExternalIssuersEnvKey)

	// KEEL_EXTERNAL_ISSUERS=https://auth.keel.xyz

	if envVar == "" {
		return []string{"https://auth.staging.keel.xyz"}
	}

	for _, uri := range strings.Split(envVar, ",") {
		if _, err := url.Parse(uri); err == nil {
			providers = append(providers, uri)
		}
	}

	return providers
}

func PublicKeyForIssuer(issuerUri string, tokenKid string) (*rsa.PublicKey, error) {
	jwks, err := NewJwks(issuerUri)

	if err != nil {
		return nil, err
	}

	publicKey, err := jwks.PublicKey(tokenKid)

	if err != nil {
		return nil, err
	}

	return publicKey, nil
}
