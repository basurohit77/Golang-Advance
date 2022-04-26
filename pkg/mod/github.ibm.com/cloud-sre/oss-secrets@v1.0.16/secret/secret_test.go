package secret

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGet tests the Get function
func TestGet(t *testing.T) {
	// Change the mount path for testing to a path that exists and set the path back at the end of the test:
	mountPath = "./testdata"
	defer func() { mountPath = defaultMountPath }()

	// Secret value not found (file does not exist):
	value := Get("secretDNE")
	assert.Equal(t, "", value)

	// Secret value found (file exists):
	value = Get("secret1")
	assert.Equal(t, "value1", value)
}

// TestLookup tests the Lookup function
func TestLookup(t *testing.T) {
	// Change the mount path for testing to a path that exists and set the path back at the end of the test:
	mountPath = "./testdata"
	defer func() { mountPath = defaultMountPath }()

	// Secret value found (file exists):
	value, found := Lookup("secret1")
	assert.Equal(t, "value1", value)
	assert.Equal(t, true, found)

	// Secret value found not found (file does not exist):
	value, found = Lookup("secretDNE")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value is found as environment variable:
	err := os.Setenv("secretAsEnvVar", "secretAsEnvVarValue")
	assert.Nil(t, err)
	value, found = Lookup("secretAsEnvVar")
	assert.Equal(t, "secretAsEnvVarValue", value)
	assert.Equal(t, true, found)
}

// TestLookupWithPath tests the lookupWithPath function
func TestLookupWithPath(t *testing.T) {
	// Secret value does not exist (empty name):
	value, found := lookupWithPath("", "./testdata")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value does not exist (empty mount path):
	value, found = lookupWithPath("secret1", "")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value does not exist (empty name and mount path):
	value, found = lookupWithPath("", "")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value does not exist (invalid name):
	value, found = lookupWithPath("secret/1", "./testdata")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value found (file exists):
	value, found = lookupWithPath("secret1", "./testdata")
	assert.Equal(t, "value1", value)
	assert.Equal(t, true, found)

	// Secret value found (file exists):
	value, found = lookupWithPath("secret2", "./testdata")
	assert.Equal(t, "value2", value)
	assert.Equal(t, true, found)

	// Secret value found (empty value) (file exists):
	value, found = lookupWithPath("secret3", "./testdata")
	assert.Equal(t, "", value)
	assert.Equal(t, true, found)

	// Secret value not found (file does not exist):
	value, found = lookupWithPath("secretDNE", "./testdata")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)

	// Secret value not found (file and path do not exist):
	value, found = lookupWithPath("secretDNE", "./testdataDNE")
	assert.Equal(t, "", value)
	assert.Equal(t, false, found)
}
