package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCryptoRandomHex(t *testing.T) {
	assert := assert.New(t)
	data, err := CryptoRandomHex(16)
	assert.Empty(err)
	assert.NotEmpty(data)
}

func TestIsValidEmail(t *testing.T) {
	assert := assert.New(t)
	not := IsValidEmail("111114543")
	assert.False(not)

	not = IsValidEmail("2222@")
	assert.False(not)


	yes := IsValidEmail("xxx@cn.ibm.com")
	assert.True(yes)
}