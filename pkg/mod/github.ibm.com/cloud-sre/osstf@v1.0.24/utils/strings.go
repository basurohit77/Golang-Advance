package utils

import (
	"strings"
	"strconv"
)

// characters allowed in the AlphaNumericString
var alphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
var alphaNumericSplit = strings.Split(alphaNumeric, "")

// From the given template, extra acceptable characters up to the specified max length
func ExtractAlphaNumericString(template string, maxLength int) string {
	// Take all alphabets and numbers from the template
	
	var index = 0
	var baseBytes = make([]byte,maxLength)
	var clientIdSplit = strings.Split(template, "")
	for _, input := range clientIdSplit {
		for _, char :=range alphaNumericSplit {
			if char == input {
				baseBytes[index] = []byte(input)[0]
				index++
				break
			}
		}
		if index > maxLength {
			break
		}
	}

	// in case there is neither alphabets nor digits in the template (non-Latin template)
	// fill in with something
	if index == 0 {
		baseBytes[0] = 'i'
		baseBytes[1] = 'd'
		index = 2
	}
	
	return string(baseBytes[:index])
}

// Generate a new string based on the given hint, 
// and ensure it is different from any string in a given collection
// by appending numbers to the end
func GenerateUniqueString(hint string, existingString []string) string {
	// suffix with number if necessary
	var suffixNumber = 1
	var newString = hint
	var test = true
	for test {
		test = false
		for _, id := range existingString {
			if id == newString {
				test = true
				newString = hint + strconv.Itoa(suffixNumber)
				suffixNumber++
				break
			}
		}
	}
	
	return newString
}
