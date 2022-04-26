package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
)

//16 bytes (AES-128) or 32 (AES-256), let's use 32
func CreateCaptcha(size int) string {
	key := make([]byte, size)
	rand.Read(key)
	return fmt.Sprintf("%x", key)
}

func CFBEncrypter(aesKey string, plainText string) (string, error) {
	key, _ := hex.DecodeString(aesKey)
	plainBytes := []byte(plainText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the cipherText.
	cipherText := make([]byte, aes.BlockSize+len(plainBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainBytes)

	return fmt.Sprintf("%x", cipherText), nil
}

func CFBDecrypter(aesKey string, cipherText string) (string, error) {

	key, _ := hex.DecodeString(aesKey)
	cipherBytes, _ := hex.DecodeString(cipherText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the cipherText.
	if len(cipherBytes) < aes.BlockSize {
		log.Println("cipherText too short")
		return "", err
	}
	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherBytes, cipherBytes)
	return fmt.Sprintf("%s", cipherBytes), nil
}


