package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// LoadPublicKey загружает публичный ключ из файла
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
    keyBytes, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyBytes)
    if block == nil {
        return nil, errors.New("failed to decode PEM block")
    }

    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    rsaPub, ok := pub.(*rsa.PublicKey)
    if !ok {
        return nil, errors.New("not an RSA public key")
    }

    return rsaPub, nil
}

// LoadPrivateKey загружает приватный ключ из файла
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
    keyBytes, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyBytes)
    if block == nil {
        return nil, errors.New("failed to decode PEM block")
    }

    privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
        // Пробуем PKCS1
        privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
        if err != nil {
            return nil, err
        }
    }

    rsaPrivate, ok := privateKey.(*rsa.PrivateKey)
    if !ok {
        return nil, errors.New("not an RSA private key")
    }

    return rsaPrivate, nil
}

// Encrypt шифрует данные с помощью публичного ключа
func Encrypt(data []byte, pubKey *rsa.PublicKey) ([]byte, error) {
    hash := sha256.New()
    ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pubKey, data, nil)
    if err != nil {
        return nil, err
    }
    return ciphertext, nil
}

// Decrypt расшифровывает данные с помощью приватного ключа
func Decrypt(ciphertext []byte, privKey *rsa.PrivateKey) ([]byte, error) {
    hash := sha256.New()
    plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privKey, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    return plaintext, nil
}