package ossvalidation

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRegisterPanicDuplicateTag(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got none")
		} else {
			testhelper.AssertEqual(t, "", "Found duplicate ossvalidation.Tag: DataMismatch (datamismatch)", r)
		}
	}()

	registerTag(TagDataMismatch)
}

func TestRegisterPanicDuplicateRunActionTag(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got none")
		} else {
			testhelper.AssertEqual(t, "", "Found duplicate ossvalidation.Tag: ProductInfo-Parts (productinfo-parts)", r)
		}
	}()

	registerTag("ProductInfo-Parts")
}

func TestAddTag(t *testing.T) {
	v := &ValidationIssue{Title: "test title", Details: "test details", Severity: WARNING}

	v.AddTag(TagPriorOSS, TagCRN)
	testhelper.AssertEqual(t, "first set of tags", []Tag{TagPriorOSS, TagCRN}, v.Tags)

	v.AddTag(TagDataMismatch, TagPriorOSS)
	testhelper.AssertEqual(t, "second set of tags np dups", []Tag{TagPriorOSS, TagCRN, TagDataMismatch}, v.Tags)
}
