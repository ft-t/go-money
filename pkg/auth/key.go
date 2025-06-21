package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type KeyGenerator struct {
}

func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

func (k *KeyGenerator) Generate() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return privateKey
}

func (k *KeyGenerator) Serialize(key *rsa.PrivateKey) []byte {
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		},
	)
}

func (k *KeyGenerator) Save(key *rsa.PrivateKey, name string) error {
	return os.WriteFile(name, k.Serialize(key), 0777)
}
