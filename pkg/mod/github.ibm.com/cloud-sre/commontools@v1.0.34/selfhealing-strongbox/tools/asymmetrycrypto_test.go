package tools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	s := CreateCaptcha(32)
	privateKey, publicKey := GenerateKeyPair(2048)

	encryptedData,err := EncryptWithPublicKey([]byte(s), publicKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, len(encryptedData) > 0)

	decryptedData, err := DecryptWithPrivateKey(encryptedData, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, s, string(decryptedData))

}


func TestReadRSAPrivateKeyFromFile(t *testing.T) {

}