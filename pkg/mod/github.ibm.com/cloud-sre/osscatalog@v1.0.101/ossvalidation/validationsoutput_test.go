package ossvalidation

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestHeader(t *testing.T) {
	ossv := setupOSSValidation()

	out := ossv.Header()

	if testing.Verbose() {
		fmt.Println(out)
	}

	if len(out) < 300 {
		t.Fatalf("Unexpectedly short output: %d chars - \n%s", len(out), out)
	}
}

func TestIssueString(t *testing.T) {
	ossv := setupOSSValidation()

	issues := ossv.GetIssues(nil)

	out := issues[0].String()

	//	out = strings.TrimSpace(out)

	if testing.Verbose() {
		fmt.Println(out)
	}

	testhelper.AssertEqual(t, "", "(warning)  [DataMismatch]  warning 2: warning 2 details\n", out)
}
