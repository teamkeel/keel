package runtimectx

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
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

// ExternalIssuersFromEnv is responsible for parsing the JSON contents of the EXTERNAL_ISSUERS environment variable
// into a map[string]*rsa.PublicKey structure. This map will contain any custom issuer definitions added by the backend and by the user
func ExternalIssuersFromEnv() (map[string]*rsa.PublicKey, error) {
	envVar := os.Getenv(ExternalIssuersEnvKey)

	if envVar == "" {
		return nil, errors.New("no external issuers in env")
	}

	issuersMap := make(map[string]string)

	// example JSON string stored in the EXTERNAL_ISSUERS environment variable:
	// { "customissuer": "{base64 public key}", "customissuer2": "{base64 public key}" }
	err := json.Unmarshal([]byte(envVar), &issuersMap)

	if err != nil {
		return nil, err
	}

	issuers := map[string]*rsa.PublicKey{}

	for key, value := range issuersMap {
		// the public key is stored in base64 in the json in order to preserve line breaks etc
		decodedBase64, err := base64.StdEncoding.DecodeString(value)

		if err != nil {
			continue
		}

		// The decoded string is expected to be in PKIX, ASN.1 DER format and will look like this:
		// -----BEGIN RSA PUBLIC KEY-----
		// MIIBigKCAYEAq3DnhgYgLVJknvDA3clATozPtjI7yauqD4/ZuqgZn4KzzzkQ4BzJ
		// ar4jRygpzbghlFn0Luk1mdVKzPUgYj0VkbRlHyYfcahbgOHixOOnXkKXrtZW7yWG
		// jXPqy/ZJ/+...
		// -----END RSA PUBLIC KEY-----

		pemBlock, _ := pem.Decode(decodedBase64)

		if pemBlock == nil {
			continue
		}

		publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)

		if err != nil {
			continue
		}

		// if we have been able to marshal the public key, then add it to the issuers registry
		if publicKey != nil {
			issuers[key] = publicKey.(*rsa.PublicKey)
		}
	}

	return issuers, nil
}
