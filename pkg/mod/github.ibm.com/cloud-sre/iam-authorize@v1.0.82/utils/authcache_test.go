package utils

import (
	"testing"
	"time"
)

func TestBasicCache(t *testing.T) {

	token := "abc12345"

	AddBadAuth(token)

	if !IsBadAuth(token) {
		t.Fatal("Caching failed to find simple token")
	}

	if IsBadAuth("ABCDEFG") {
		t.Fatal("False positive on simple token")
	}
}

func TestCacheExpire(t *testing.T) {

	token := "abc12345"

	AddBadAuth(token)

	if !IsBadAuth(token) {
		t.Fatal("Caching failed to find simple token")
	}

	authCacheInstance.tokens[token] = time.Now().Add(time.Second * -2)

	if IsBadAuth(token) {
		t.Fatal("False positive on simple token")
	}
}
