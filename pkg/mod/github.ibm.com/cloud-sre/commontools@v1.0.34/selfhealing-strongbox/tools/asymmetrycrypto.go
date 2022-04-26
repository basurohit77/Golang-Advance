package tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

// GenerateKeyPair generates a new key pair
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Println(err)
	}
	return privateKey, &privateKey.PublicKey
}

// PrivateKeyToBytes private key to bytes
func PrivateKeyToBytes(privateKey *rsa.PrivateKey) []byte {
	privateBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return privateBytes
}

// PublicKeyToBytes public key to bytes
func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Println(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

func ReadRSAPrivateKeyFromFile(file string) (*rsa.PrivateKey, error) {
	if file == "" {
		return nil, fmt.Errorf("read rsa private key failed")
	}
	data, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return BytesToPrivateKey(data), nil
}

func ReadRSAPublicKeyFromFile(file string) (*rsa.PublicKey, error) {
	if file == "" {
		return nil, fmt.Errorf("read rsa public key failed")
	}
	data, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return BytesToPublicKey(data), nil
}

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Println(err)
	}
	return key
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Println(err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Println("not ok")
	}
	return key
}

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) (string, error) {
	hash := sha512.New()
	cipherText, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return fmt.Sprintf("%x", cipherText), nil
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(cipherText string, privateKey *rsa.PrivateKey) (string, error) {
	cipher, _ := hex.DecodeString(cipherText)
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, cipher, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(plaintext), nil
}
