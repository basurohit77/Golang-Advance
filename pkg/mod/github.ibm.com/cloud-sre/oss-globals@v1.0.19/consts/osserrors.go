package consts

import "errors"

var (
	// ErrMalformedCRN is an error for malformed CRN
	ErrMalformedCRN = errors.New("malformed CRN")
	// ErrMalformedScope is an error for malformed scopes
	ErrMalformedScope = errors.New("malformed scope in CRN")
	// ErrEmptyCRN is an error for no CRN to process
	ErrEmptyCRN = errors.New("no CRN to process")
	//PrettyJSONFail error message printing JSON
	PrettyJSONFail = "Failed pretty printing json"
)

const (
	//ErrJSONDecode Error occurred when trying to decode JSON
	ErrJSONDecode = "Error trying to decode: "
	//ErrJSONEncode Error occurred when trying to encode JSON
	ErrJSONEncode = "Error trying to encode: "
	// ErrBadCRNFmt  CRN  has incorrect format"
	ErrBadCRNFmt = "CRN has incorrect format"
)
