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

func GetExternalIssuers(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	v := ctx.Value(externalIssuersKeyContext)

	if v == nil {
		return nil, nil
	}

	issuers, ok := v.(map[string]*rsa.PublicKey)

	if !ok {
		return nil, errors.New("external issuers not in context")
	}

	return issuers, nil
}

func WithExternalIssuers(ctx context.Context, issuers map[string]*rsa.PublicKey) context.Context {
	return context.WithValue(ctx, externalIssuersKeyContext, issuers)
}

const (
	ExternalIssuersEnvKey = "EXTERNAL_ISSUERS"
)

type Jwks struct {
	Set jwk.Set
}

func NewJwks(uri string) (*Jwks, error) {
	jwksUri, err := url.Parse(fmt.Sprintf("%s/.well-known/jwks.json", uri))

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

func (j *Jwks) PublicKey() (*rsa.PublicKey, error) {
	allKeys := j.Set.Keys(context.Background())

	found := false

	var publicKey rsa.PublicKey

	for allKeys.Next(context.Background()) {
		curr := allKeys.Pair()

		switch v := curr.Value.(type) {
		case jwk.RSAPublicKey:

			found = true

			err := v.Raw(&publicKey)

			if err != nil {
				found = false
			}
		}

	}

	if !found {
		return nil, errors.New("No RSA public key found")
	}

	return &publicKey, nil
}

// ExternalIssuersFromEnv is responsible for parsing the contents of the EXTERNAL_ISSUERS environment variable
// into a []string. This environment variable is a comma separated list of authorization server uris.
func ExternalIssuersFromEnv() map[string]*rsa.PublicKey {
	issuers := make(map[string]*rsa.PublicKey)
	envVar := os.Getenv(ExternalIssuersEnvKey)

	if envVar == "" {
		return make(map[string]*rsa.PublicKey)
	}

	for _, uri := range strings.Split(envVar, ",") {
		jwks, err := NewJwks(uri)

		if err != nil {
			continue
		}

		publicKey, err := jwks.PublicKey()

		if err != nil {
			continue
		}

		if publicKey != nil {
			issuers[uri] = publicKey
		}
	}

	return issuers
}
