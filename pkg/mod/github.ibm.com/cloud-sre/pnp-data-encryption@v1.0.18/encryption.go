package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"
)

var (
	masterKey = os.Getenv("MASTER_KEY")
)

// newKDF derives an encryption key from the master key and the nonce
func newKDF(masterkey []byte, nonce []byte) ([]byte, error) {
	key := make([]byte, 32)

	kdf := hkdf.New(sha256.New, masterkey, nonce, nil)

	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}

	return key, nil

}

// newGCM creates an AES-256 block cipher, and returns the block in Galois Counter Mode, and an error
func newGCM(key []byte) (cipher.AEAD, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm, nil

}
