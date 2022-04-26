package main

import (
	"errors"
	"testing"

	"github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
)

func TestRestarter(t *testing.T) {
	restarter(0, "ut", goodRestart)
}

var retryCount = 0

func goodRestart() {
	if retryCount < 1 {
		retryCount++
		panic("test panic")
	}
}

func TestStartup(t *testing.T) {
	initadapter.Initialize = myutInitialize
	startup()
}

func myutInitialize() (*api.SourceConfig, error) {
	return nil, nil
}

func TestStartup2(t *testing.T) {
	initadapter.Initialize = myutInitialize2
	startup()
}

func myutInitialize2() (*api.SourceConfig, error) {
	return nil, errors.New("ut error")
}

func TestRestarter2(t *testing.T) {
	retryCount = 0
	restarter(0, "ut", goodRestart2)
}

func goodRestart2() {
	if retryCount < 1 {
		retryCount++
		restarter(11, "ut2", badRestart)
	}
}

func badRestart() {
	panic("really bad")
}
