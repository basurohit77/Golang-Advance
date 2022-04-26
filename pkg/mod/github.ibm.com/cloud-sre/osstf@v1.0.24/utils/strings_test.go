package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUniqueString(t *testing.T) {
	expected := "must1"
	actual := GenerateUniqueString("must", []string{"stop", "test", "must", "musting"})
	assert.Equal(t, expected, actual)
}

func TestExtractAlphaNumericString(t *testing.T) {
	expected := "Aabb45"
	actual := ExtractAlphaNumericString("Aabb45*", 7)
	assert.Equal(t, expected, actual)
}
