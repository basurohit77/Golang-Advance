package iammock

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain will setup the test environment for iammock.
func TestMain(m *testing.M) {
	SetupIAMMockGeneric()

	os.Exit(m.Run())
}

// TestEnvVarsValues will checks whether the value of the IAM's environment variables is correct.
func TestEnvVarsValues(t *testing.T) {
	assert.Equal(t, `test`, os.Getenv(`IAM_MODE`), `value of $IAM_MODE`)
	assert.Equal(t, `true`, os.Getenv(`IAM_BYPASS_FLAG`), `value of $IAM_BYPASS_FLAG`)
	assert.Regexp(t, regexp.MustCompile(`^http://127.0.0.1:\d\d\d\d\d$`), os.Getenv(`IAM_URL`), `value of $IAM_URL`)
}
