package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	keelconfig "github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/deploy"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var log *logrus.Logger
var dbConn db.Database
var dbConnString string
var schema *proto.Schema
var keelConfig *keelconfig.ProjectConfig
var privateKey *rsa.PrivateKey

func initSchema() {
	b, err := os.ReadFile("schema.json")
	if err != nil {
		panic(err)
	}

	var s proto.Schema
	err = protojson.Unmarshal(b, &s)
	if err != nil {
		panic(err)
	}

	schema = &s
}

func initConfig() {
	b, err := os.ReadFile("keelconfig.yaml")
	if err != nil {
		panic(err)
	}

	var c keelconfig.ProjectConfig
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}

	keelConfig = &c
}

func initPrivateKey() {
	privateKeyPem, ok := secrets["KEEL_PRIVATE_KEY"]
	if !ok {
		panic("missing KEEL_PRIVATE_KEY secret")
	}

	privateKeyBlock, _ := pem.Decode([]byte(privateKeyPem))
	if privateKeyBlock == nil {
		panic("error decoding private key PEM")
	}

	var err error
	privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	initLogger()
	initSchema()
	initConfig()
	initTracing()
	initSecrets()
	initPrivateKey()
	initDB()
	initEvents()

	switch os.Getenv("KEEL_RUNTIME_MODE") {
	case deploy.RuntimeModeApi:
		lambda.Start(apiHandler)
	case deploy.RuntimeModeSubscriber:
		lambda.Start(eventHandler)
	}
}
