package tools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateCaptcha(t *testing.T) {
	key := CreateCaptcha(32)
	assert.Equal(t, 64, len(key))
}

func TestSymmetry(t *testing.T) {
	key := CreateCaptcha(32)
	plainText := "test 123"
	cipherText, _ := CFBEncrypter(key, plainText)
	assert.Equal(t, true, len(cipherText) > 0)

	decrpytedText, _ := CFBDecrypter(key, cipherText)
	assert.Equal(t, plainText, decrpytedText)
}