package encryption

import (
	"encoding/hex"
	"errors"
	"log"
)

// Decrypt decrypts data. nonce is expected to be at the end of the payload.
// returns the decrypted data and an error
func Decrypt(data []byte) ([]byte, error) {

	msg := make([]byte, len(data))
	// make a copy of data so that we only change the copy, not the orignal data. See https://github.ibm.com/cloud-sre/pnp-data-encryption/issues/6
	copy(msg, data)

	// masterkey, err := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here
	masterkey, err := hex.DecodeString(masterKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// ensure msg is long enough to contain nonce
	if len(msg) < 12 {
		err = errors.New("Data is not long enough to be decrypted")
		log.Println(err)
		return nil, err
	}

	// retrieve nonce from msg, it's the last 12 bytes
	nonce := msg[len(msg)-12:]

	// derive an encryption key from the master key and the nonce
	key := make([]byte, 32)
	if key, err = newKDF(masterkey, nonce); err != nil {
		log.Println(err)
		return nil, err
	}

	// creates an AES-256 block in GCM
	aesgcm, err := newGCM(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// decrypt msg
	decryptedData, err := aesgcm.Open(msg[:0], nonce, msg[:len(msg)-12], nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return decryptedData, nil
}
