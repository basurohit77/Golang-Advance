package token_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.ibm.com/IAM/basiclog"
	"github.ibm.com/IAM/token/v5"
)

func ExampleTokenValidator() {

	/* #nosec G101 */
	expiredToken := "eyJraWQiOiIyMDIxMDgyNTE1MDkiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTIxMjcxZDczLWFkYWUtNDEwZC04MjBiLTdhNTgyYjAwNjZmYyIsImlkIjoiaWFtLVNlcnZpY2VJZC0yMTI3MWQ3My1hZGFlLTQxMGQtODIwYi03YTU4MmIwMDY2ZmMiLCJyZWFsbWlkIjoiaWFtIiwianRpIjoiYTZmZmFmZjctY2FjZi00ZGMxLTg4NDgtZDcyOGZlODVlNTQxIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC0yMTI3MWQ3My1hZGFlLTQxMGQtODIwYi03YTU4MmIwMDY2ZmMiLCJuYW1lIjoiaXNtLXJlc291cmNlLWNvbnRyb2xsZXItYXBpLSBTZXJ2aWNlSWQiLCJzdWIiOiJTZXJ2aWNlSWQtMjEyNzFkNzMtYWRhZS00MTBkLTgyMGItN2E1ODJiMDA2NmZjIiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhdXRobiI6eyJzdWIiOiJTZXJ2aWNlSWQtMjEyNzFkNzMtYWRhZS00MTBkLTgyMGItN2E1ODJiMDA2NmZjIiwiaWFtX2lkIjoiaWFtLVNlcnZpY2VJZC0yMTI3MWQ3My1hZGFlLTQxMGQtODIwYi03YTU4MmIwMDY2ZmMiLCJzdWJfdHlwZSI6IlNlcnZpY2VJZCIsIm5hbWUiOiJpc20tcmVzb3VyY2UtY29udHJvbGxlci1hcGktIFNlcnZpY2VJZCJ9LCJhY2NvdW50Ijp7ImJvdW5kYXJ5IjoiZ2xvYmFsIiwidmFsaWQiOnRydWUsImJzcyI6IjQ2MzUyYzM1ZTEyODVmMTY4ZGIzMzk2OTAyMmNkODQxIiwiZnJvemVuIjp0cnVlfSwiaWF0IjoxNjMwMDk0NTU2LCJleHAiOjE2MzAwOTgxNTYsImlzcyI6Imh0dHBzOi8vaWFtLnRlc3QuY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImRlZmF1bHQiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.Bh1i0Yg0BWWycJv_78UxeO5PNpg9gSEgccYULNOhjPbc4vR7zLtNpA4idmYlop93CnAOoHIOc7Q_8UUHuDGUgQgJRptK7x4asPY77jjFn9gnfOrSWtdlszT6Nd1L1BsDU5dAbzKiYmHHmNgxPQrqOD_Ibjrjst_et0lTe4_gpOjRMPMI-2B2UcXtGvzXnke6m8YLTxL3MWHcQ5WTIRFuzJfib_KTG-HI-n6ldjVn7vHGJWwNAcFZrXN8oAYC9FXAK6D0uT2_XcYbyDLeoyBWdozxaRyZC211AXxwJB2ElBGMiZiOnKW6I90sCbk6enx7qi5tkz9ipJpiS8RCjTvqrA"

	// create a token validator
	tv, err := token.NewTokenValidator(token.StagingHostURL + token.KeyPath)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	// use a token validator to extract claims and validate the token
	validatorClaims, err := tv.GetClaims(expiredToken, false)

	fmt.Println("claims:", validatorClaims)
	fmt.Println("Key does not exist for the token's KID:", strings.Contains(err.Error(), "Key does not exist for the token's KID"))

	// Output: claims: <nil>
	// Key does not exist for the token's KID: true
}
func ExampleNewTokenManager() {

	// create a new token manager with your API key, here it is fetched from an environment variable
	// but it is just a string being passed to NewToken()

	tm, err := token.NewTokenManager(os.Getenv("API_KEY"), token.Staging, nil)

	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	// get the token from the token manager
	tk, err := tm.GetToken()
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Token fetched: %t\n", (len(tk) > 0))

	// extract the claims from the encoded token and validate the token at the same time
	claims, err := tm.GetClaims(tk, false)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	// we can also validate that token using this convenience function
	err = claims.Valid()
	fmt.Println("error is nil:", err == nil)

	if !strings.Contains(claims.IAMID, "iam-ServiceId") {
		log.Fatal("error: claims are invalid", claims.IAMID)
	}

	// We can also validate this token with the TokenValidator

	// create a token validator
	tv, err := token.NewTokenValidator(token.StagingHostURL + token.KeyPath)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	// use a token validator to extract claims and validate the token
	validatorClaims, err := tv.GetClaims(tk, false)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("IAMID is valid and contains the expected substring:", strings.Contains(validatorClaims.IAMID, "iam-ServiceId"))
	// Output: Token fetched: true
	// error is nil: true
	// IAMID is valid and contains the expected substring: true
}

func ExampleGetDelegationToken() {
	tm, err := token.NewTokenManager(os.Getenv("API_KEY"), token.Staging, nil)

	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	iamID := "crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance123456::"

	delegationToken, err := tm.GetDelegationToken(iamID)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	// extract the claims from the encoded token
	claims, err := tm.GetClaims(delegationToken, false)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("iam_id in delegation token: %v\n", claims.IAMID)

	// Output: iam_id in delegation token: crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance123456::
}

func ExampleExtendedConfig_Logger() {
	// Shows how to enable the printing of error messages for situations that do not get automatically re-run
	// on a short cadence when errors occurs

	endpoints := &token.Endpoints{
		TokenEndpoint: "http://localhost:59669/badurl",
		KeyEndpoint:   token.StagingHostURL + token.KeyPath,
	}

	logger, err := basiclog.NewBasicLogger(basiclog.LevelDebug, os.Stdout)

	if err != nil {
		log.Fatal(err)
	}

	tokenConfig := &token.ExtendedConfig{
		ClientID:     "bx", // optional ClientID and ClientSecret
		ClientSecret: "bx",
		Endpoints:    *endpoints,
		Logger:       logger,
	}

	_, err = token.NewTokenManager(os.Getenv("API_KEY"), token.Custom, tokenConfig)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	// ERROR: token.go:178:  Transaction-Id: 2e3a445a-81a9-41cd-ab8e-adc1cafda99b: Post "http://localhost:59669/badurl": dial tcp [::1]:59669: connect: connection refused unable to retrieve token
}
