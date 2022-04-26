package bgcache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sort"
	"time"

	"github.ibm.com/cloud-sre/oss-secrets/secret"

	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"

	"encoding/json"
)

// EncryptionKeys is the environment variable stored in vault
type EncryptionKeys struct {
	Keys map[int64]string
}

// getMasterKey is used for obtaining the right encryption key and its corresponding key id
// To obtain the current encryption key, use time.Now().Unix() as parameter
func getMasterKey(timeInUnix int64) (string, int64, error) {
	// Get the EncryptionKeys object from the environment
	var encryptionKeys EncryptionKeys
	err := json.Unmarshal([]byte(secret.Get("BGCACHE_MASTER_KEY")), &encryptionKeys)
	if err != nil {
		return "", 0, errors.New("failed to unmarshal master key")
	}

	// Find the encryption key ID
	masterKeyID, keyIDErr := getMasterKeyID(timeInUnix, encryptionKeys.Keys)
	if keyIDErr != nil {
		return "", 0, keyIDErr
	}

	// Find the encryption key based on key ID
	masterKey := encryptionKeys.Keys[masterKeyID]
	if masterKey == "" {
		// Do not proceed if the encryption key in vault is empty
		return "", masterKeyID, errors.New("masterKey is empty")
	}

	return masterKey, masterKeyID, nil
}

// getMasterKeyID is used for obtaining the right encryption key ID
// To obtain the current encryption key ID, use time.Now().Unix() as parameter
func getMasterKeyID(timeInUnix int64, keyMap map[int64]string) (int64, error) {
	if len(keyMap) == 0 {
		return 0, errors.New("masterKeyID is empty")
	}

	// Store keyIDs
	keyIDs := make([]int64, 0, len(keyMap))
	for k := range keyMap {
		keyIDs = append(keyIDs, k)
	}

	// Sort keyIDs in descending order
	sort.Slice(keyIDs, func(i, j int) bool { return keyIDs[i] >= keyIDs[j] })

	// Find the largest encryption key ID less than or equal to timeInUnix
	for i := 0; i < len(keyIDs); i++ {
		if keyIDs[i] > timeInUnix {
			continue
		} else {
			// Encryption key ID is found
			return keyIDs[i], nil
		}
	}

	return 0, errors.New("masterKeyID is empty")
}

// Encrypt encrypts data using AES_256_GCM, having a unique derivated key+nonce per message
// and returns the encrypted text, encryption key id, and an error
func Encrypt(data string) ([]byte, int64, error) {
	// masterKey, err := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here
	// Use the current encryption key for encryption
	keyFromEnv, keyIDFromEnv, err := getMasterKey(time.Now().Unix())
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}

	masterKey, err := hex.DecodeString(keyFromEnv)
	if err != nil {
		err = errors.New("masterKey is invalid for encryption")
		log.Println(err)
		return nil, keyIDFromEnv, err
	}

	// Generate a random nonce to derive an encryption key from the master key
	// This nonce must be saved to be able to decrypt the data again - it is not required to keep it secret
	// Must have one nonce per encryption, as well as one derivated key
	nonce := make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		err = errors.New("failed to read random data")
		log.Println(err)
		return nil, keyIDFromEnv, err
	}

	// Create a new key based on the nonce and the master key
	// Must have one key per encryption
	key := make([]byte, 32)
	if key, err = newKDF(masterKey, nonce); err != nil {
		err = errors.New("failed to derive encryption key")
		log.Println(err)
		return nil, keyIDFromEnv, err
	}

	// Create an AES-256 block in GCM
	aesGCM, err := newGCM(key)
	if err != nil {
		err = errors.New("newGCM failed for encryption")
		log.Println(err)
		return nil, keyIDFromEnv, err
	}

	// Encrypt data
	encryptedText := aesGCM.Seal([]byte(data)[:0], nonce, []byte(data), nil)

	// Append nonce to the encrypted data, as the nonce will be extracted when decrypting the data
	encryptedText = append(encryptedText, nonce...)

	return encryptedText, keyIDFromEnv, nil
}

// Decrypt decrypts data (nonce is expected to be at the end of the payload)
// and returns the decrypted data, and an error
// keyID is the version of encryption key used at time of encryption
func Decrypt(data []byte, keyID int64) ([]byte, error) {
	msg := make([]byte, len(data))
	// Make a copy of data so that we only change the copy, not the original data
	// See https://github.ibm.com/cloud-sre/pnp-data-encryption/issues/6
	copy(msg, data)

	// masterKey, err := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here
	// Get the encryption key according to keyID
	keyFromEnv, _, err := getMasterKey(keyID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	masterKey, err := hex.DecodeString(keyFromEnv)
	if err != nil {
		err = errors.New("masterKey is invalid for decryption")
		log.Println(err)
		return nil, err
	}

	// Ensure msg is long enough to contain nonce
	if len(msg) < 12 {
		err = errors.New("data is invalid for decryption")
		log.Println(err)
		return nil, err
	}

	// Retrieve nonce from msg - it's the last 12 bytes
	nonce := msg[len(msg)-12:]

	// Derive an encryption key from the master key and the nonce
	key := make([]byte, 32)
	if key, err = newKDF(masterKey, nonce); err != nil {
		err = errors.New("failed to derive encryption key for decryption")
		log.Println(err)
		return nil, err
	}

	// Create an AES-256 block in GCM
	aesGCM, err := newGCM(key)
	if err != nil {
		err = errors.New("newGCM failed for decryption")
		log.Println(err)
		return nil, err
	}

	// Decrypt msg
	decryptedData, err := aesGCM.Open(msg[:0], nonce, msg[:len(msg)-12], nil)
	if err != nil {
		err = errors.New("failed to decrypt data")
		log.Println(err)
		return nil, err
	}

	return decryptedData, nil
}

// newKDF derives an encryption key from the master key and the nonce
func newKDF(masterKey []byte, nonce []byte) ([]byte, error) {
	key := make([]byte, 32)

	kdf := hkdf.New(sha256.New, masterKey, nonce, nil)

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

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM, nil
}
