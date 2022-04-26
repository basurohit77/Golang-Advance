package encryption

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
)

// Encrypt encrypts data using AES_256_GCM, having a unique derivated key+nonce per message
// and returns the encrypted text, and an error
func Encrypt(data string) ([]byte, error) {

	// masterkey, err := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here
	masterkey, err := hex.DecodeString(masterKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// generate a random nonce to derive an encryption key from the master key
	// this nonce must be saved to be able to decrypt the data again - it is not required to keep it secret
	// must have one nonce per encryption, as well as one derivated key
	nonce := make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Printf("Failed to read random data: %v", err)
		return nil, err
	}

	// create a new key, based on the nonce and the master key
	// must have one key per encryption
	key := make([]byte, 32)
	if key, err = newKDF(masterkey, nonce); err != nil {
		fmt.Printf("Failed to derive encryption key: %v", err)
		return nil, err
	}

	// creates an AES-256 block in GCM
	aesgcm, err := newGCM(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// encrypt data
	encryptedText := aesgcm.Seal([]byte(data)[:0], nonce, []byte(data), nil)

	// append nonce to the encrypted msg, as the nonce will be extracted when decrypting the data
	encryptedText = append(encryptedText, nonce...)

	return encryptedText, nil
}
