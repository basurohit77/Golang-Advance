package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

// DecodeByte will decode a []byte to JSON
func DecodeByte(fct, objName string, str []byte, intfce interface{}) error {
	return Decode(fct, objName, bytes.NewReader(str), intfce)
}

// DecodeStr will decode a string to JSON
func DecodeStr(fct, objName string, str string, intfce interface{}) error {
	return Decode(fct, objName, strings.NewReader(str), intfce)
}

// Decode is a wrapper for JSON Decoder
func Decode(fct, objName string, rdr io.Reader, intfce interface{}) error {
	METHOD := fct + "->[Decode]"

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Println(METHOD, "Failed to read response body:", err.Error())
		return err
	}

	log.Println(METHOD, "Response returned:", string(data))

	err = json.NewDecoder(bytes.NewReader(data)).Decode(intfce)
	if err != nil {
		log.Println(METHOD, "Failed to decode JSON object", objName, ":", err.Error())
		return err
	}

	return nil
}
