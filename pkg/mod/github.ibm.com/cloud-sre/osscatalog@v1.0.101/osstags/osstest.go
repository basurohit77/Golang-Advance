package osstags

import (
	"regexp"
	"strings"
)

// ossTestDisplayName is the prefix used to mark the display name of all OSS records used for testing (identified by the OSSTest tag)
const ossTestDisplayName = "TEST RECORD:"

//const ossTestDisplayNameAlt = "*TEST RECORD*"

var ossTestPattern = regexp.MustCompile(`(?i)^\s*[*]?TEST[ _-]+RECORD[*: ]+`)

// GetOSSTestBaseName returns the base name of a displayName that is prefixed as a TEST RECORD
func GetOSSTestBaseName(displayName string) (baseName string, isTest bool) {
	if prefix := ossTestPattern.FindString(displayName); prefix != "" {
		return displayName[len(prefix):], true
	}
	return displayName, false
}

// CheckOSSTestTag checks if a OSS record is tagged as being a test record.
// and adjusts the DisplayName and OSSTags accordingly if needed.
// It returns true of the record is a test record, false otherwise.
func CheckOSSTestTag(displayName *string, tags *TagSet) bool {
	if strings.HasPrefix(*displayName, ossTestDisplayName) {
		tags.AddTag(OSSTest) // May be already there
		return true
		/*
			case strings.HasPrefix(trimmed2, ossTestDisplayNameAlt):
				*displayName = ossTestDisplayName + trimmed[len(ossTestDisplayNameAlt):]
				tags.AddTag(OSSTest) // May be already there
				return true
		*/
	} else if baseName, isTest := GetOSSTestBaseName(*displayName); isTest {
		*displayName = ossTestDisplayName + " " + baseName
		tags.AddTag(OSSTest) // May be already there
		return true
	} else if tags.Contains(OSSTest) {
		*displayName = ossTestDisplayName + " " + *displayName
		return true
	}
	return false
}
