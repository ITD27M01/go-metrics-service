package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"io"
	"os"
)

var ErrBadPublicKeyFormat = errors.New("wrong public key format, ensure -----BEGIN PUBLIC KEY----- header")

// ReadPrivateKey reads private key from file
func ReadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	if privateKeyPath == "" {
		return nil, nil
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(privateKeyBytes)
	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ReadPublicKey reads public key from file
func ReadPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	if publicKeyPath == "" {
		return nil, nil
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(publicKeyBytes)
	untypedKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := untypedKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrBadPublicKeyFormat
	}

	return key, nil
}

// RSAEncrypt encrypts data with public key
func RSAEncrypt(msg []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return msg, nil
	}

	ciphertext, err := EncryptOAEP(sha512.New(), rand.Reader, publicKey, msg, nil)

	return ciphertext, err
}

// RSADecrypt decrypts data with private key
func RSADecrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return ciphertext, nil
	}

	plaintext, err := DecryptOAEP(sha512.New(), rand.Reader, privateKey, ciphertext, nil)

	return plaintext, err
}

// EncryptOAEP is a chanked format of rsa func
func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

// DecryptOAEP is a chanked format of rsa func
func DecryptOAEP(hash hash.Hash, random io.Reader, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, private, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}
