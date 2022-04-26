package ossrecord

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestConstructPerson(t *testing.T) {
	oneTest := func(input, w3id, name string) func(t *testing.T) {
		return func(t *testing.T) {
			p := ConstructPerson(input)
			testhelper.AssertEqual(t, t.Name()+".W3ID", w3id, p.W3ID)
			testhelper.AssertEqual(t, t.Name()+".Name", name, p.Name)
		}
	}

	t.Run("simple-email", oneTest("foo@bar.baz.com", "foo@bar.baz.com", ""))
	t.Run("simple-email", oneTest("John Doe foo@bar.baz.com", "foo@bar.baz.com", "John Doe"))
	t.Run("simple-email", oneTest("foo@bar.baz.com: John Doe", "foo@bar.baz.com", ": John Doe"))
	t.Run("simple-email", oneTest("John <foo@bar.baz.com> Doe", "foo@bar.baz.com", "John <> Doe"))
	t.Run("simple-email", oneTest("John Doe foo@", "", "John Doe foo@"))
	t.Run("simple-email", oneTest("John Doe @foo", "", "John Doe @foo"))
	t.Run("simple-email", oneTest("", "", ""))
	t.Run("simple-email", oneTest("  ", "", ""))
}

func TestNormalizeProductID(t *testing.T) {
	testhelper.AssertEqual(t, "", NormalizeProductID("5737E34"), "5737E34")
	testhelper.AssertEqual(t, "", NormalizeProductID(" 5737-e34 "), "5737E34")
	testhelper.AssertEqual(t, "", NormalizeProductID(" 5737 e34 "), "5737E34")
}

func TestDeepCopy(t *testing.T) {
	src := CreateTestRecord()
	dest := src.DeepCopy()

	originalName := src.ReferenceResourceName
	originalTags := src.GeneralInfo.OSSTags
	fmt.Println("originalTags:", originalTags)

	testhelper.AssertEqual(t, "ReferenceResourceName after copy", originalName, dest.ReferenceResourceName)
	testhelper.AssertEqual(t, "OSSTags after copy", originalTags, dest.GeneralInfo.OSSTags)

	dest.ReferenceResourceName = dest.ReferenceResourceName + "-modified"
	dest.GeneralInfo.OSSTags = append(dest.GeneralInfo.OSSTags, "modified")

	testhelper.AssertEqual(t, "ReferenceResourceName original", originalName, src.ReferenceResourceName)
	testhelper.AssertEqual(t, "OSSTags original", originalTags, src.GeneralInfo.OSSTags)
}
