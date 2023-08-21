package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run pkgen.go [filename]")
		os.Exit(1)
	}

	filename := os.Args[1]
	err := GeneratePrivateKey(filename)
	if err != nil {
		fmt.Println("Error generating private key:", err)
		os.Exit(1)
	}
}

func GeneratePrivateKey(filename string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	pemFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer pemFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	err = pem.Encode(pemFile, privateKeyBlock)
	if err != nil {
		return err
	}

	return nil
}
