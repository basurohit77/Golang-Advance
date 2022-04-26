package ossuid

import (
	"fmt"
	"strings"
)

// Functions to allocate and manipulate a unique OSS UID for each entry, that can be used in lieu of ProductID wheh no ProductID is available

// UID is a OSS UID (as an opaque type, that can be converted into a string)
type UID struct {
	numeric int
}

// BaseValue is the first value for counting all OSS UIDs (as a string)
const BaseValue = "OSS0A00"

const prefix = "OSS"
const digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numericLength = 4

var radix = len(digits)

// Parse converts a string into a OSS UID
func Parse(in string) UID {
	if !strings.HasPrefix(in, prefix) {
		panic(fmt.Sprintf(`Invalid prefix in OSS UID "%s" (expected "%s")`, in, prefix))
	}
	in0 := strings.TrimPrefix(in, prefix)
	if len(in0) != numericLength {
		panic(fmt.Sprintf(`Unexpected length for numeric part of OSS UID "%s" (expected %d got %d)`, in, numericLength, len(in0)))
	}
	var result UID
	for _, c := range in0 {
		ix := strings.IndexRune(digits, c)
		if ix == -1 {
			panic(fmt.Sprintf(`Invalid digit character in OSS UID "%s": "%s"`, in, string(c)))
		}
		result.numeric = (result.numeric * radix) + ix
	}
	return result
}

// String returns this UID as a string
func (uid UID) String() string {
	runes := make([]rune, 0, numericLength)
	num := uid.numeric
	for num > 0 {
		ix := num % radix
		num = num / radix
		runes = append(runes, rune(digits[ix]))
	}
	result := strings.Builder{}
	result.WriteString(prefix)
	for i := numericLength; i > len(runes); i-- {
		result.WriteString("0")
	}
	for i := len(runes) - 1; i >= 0; i-- {
		result.WriteRune(runes[i])
	}
	if len(runes) > numericLength {
		panic(fmt.Sprintf(`Numeric value of out range for OSS UID "%s" (numeric=%d, max=%d)`, result.String(), uid.numeric, radix^numericLength))
	}
	return result.String()
}

// Increment increments this UID to the next available value and returns a new UID
func (uid UID) Increment() UID {
	return UID{uid.numeric + 1}
}
