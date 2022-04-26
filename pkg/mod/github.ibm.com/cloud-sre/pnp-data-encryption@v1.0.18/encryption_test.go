package encryption

import (
	"os"
	"testing"
)

var (
	encryptedText []byte
	data          = "some dummy data"
)

func TestEncrypt(t *testing.T) {

	setEnv(t)
	var err error

	encryptedText, err = Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}

}

func TestDecrypt(t *testing.T) {

	decryptedData, err := Decrypt(encryptedText)
	if err != nil {
		t.Fatal(err)
	}

	if data != string(decryptedData) {
		t.Fail()
	}

}

func setEnv(t *testing.T) {

	if err := os.Setenv("MASTER_KEY", "000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000"); err != nil {
		t.Fatal(err)
	}

}
