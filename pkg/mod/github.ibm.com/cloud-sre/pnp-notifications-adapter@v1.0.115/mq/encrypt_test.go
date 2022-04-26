package mq

import (
	"errors"
	"log"
	"os"
	"testing"

	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
)

func TestEncrypt(t *testing.T) {

	os.Setenv("MASTER_KEY", "1234567890")

	myProduce("{this is a test message}")
}

// Produce will produce a message to the queue
func myProduce(msg string) error {

	var err error
	var outputMsg []byte

	outputMsg, err = encryption.Encrypt(msg)
	if err != nil {
		log.Println("pnp-notifications-adapter.mq.Produce: Encryption failure: " + err.Error())
		return err
	}

	decryptedMsg := outputMsg

	decryptedMsg, err = encryption.Decrypt(outputMsg)
	if err != nil {
		log.Println(": failed to decrypt message err=" + err.Error())
		return err
	}

	if string(decryptedMsg) != msg {
		return errors.New("did not decrypt")
	}
	return nil
}
