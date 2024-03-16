package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	privateKeyFile, err := os.Create("./private.pem")
	if err != nil {
		panic(err)
	}
	publicKeyFile, err := os.Create("./public.pem")
	if err != nil {
		panic(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 16384)
	if err != nil {
		panic(err)
	}

	publicKey := &privateKey.PublicKey

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	err = pem.Encode(privateKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err != nil {
		panic(err)
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)

	err = pem.Encode(publicKeyFile, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err != nil {
		panic(err)
	}

}
