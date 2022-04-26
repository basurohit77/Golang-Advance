package bgcache

import (
	"os"
	"testing"
)

var (
	encryptedText []byte
	data          = "dummy data"
)

func TestEncrypt(t *testing.T) {
	setEnv(t)
	var err error
	var encryptionKeyID int64

	encryptedText, encryptionKeyID, err = Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	if encryptionKeyID != currentKeyID {
		t.Fail()
	}
}

func TestDecrypt(t *testing.T) {
	decryptedData, err := Decrypt(encryptedText, currentKeyID)
	if err != nil {
		t.Fatal(err)
	}

	if data != string(decryptedData) {
		t.Fail()
	}
}

func setEnv(t *testing.T) {
	// We can add new master key value pair to BGCACHE_MASTER_KEY for encryption key rotation
	masterKeyInVault := `{"keys":{"1262303999":"000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201001","2366841599":"000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000"}}`

	if err := os.Setenv("BGCACHE_MASTER_KEY", masterKeyInVault); err != nil {
		t.Fatal(err)
	}
}

