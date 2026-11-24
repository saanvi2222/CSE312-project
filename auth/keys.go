package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func loadOrCreateRSAKeys(dir string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privPath := filepath.Join(dir, "private.pem")
	pubPath := filepath.Join(dir, "public.pem")

	//checks if the file path to private is on disk

	if _, err := os.Stat(privPath); err == nil {
		privBytes, err := os.ReadFile(privPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read private key: %w", err)
		}
		block, _ := pem.Decode(privBytes)
		if block == nil {
			return nil, nil, fmt.Errorf("invalid private key PEM")
		}
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parse private key: %w", err)
		}
		return privKey, &privKey.PublicKey, nil
	}

	//create the directory for pub & priv
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, nil, fmt.Errorf("create keys dir: %w", err)
	}
	//create an RSA key pair as a struct
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate RSA key: %w", err)
	}

	//store priv key in a perm block
	//x509.MarshalPKCS1PrivateKey(privKey) converts that *rsa.PrivateKey struct into standard binary bytes
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	//writing private key to disk privPath
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		return nil, nil, fmt.Errorf("write private key: %w", err)
	}

	//converting the public key struct to bytes
	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}
	//putting public key in a perm block
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})
	//write it to disk
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return nil, nil, fmt.Errorf("write public key: %w", err)
	}

	return privKey, &privKey.PublicKey, nil
}
