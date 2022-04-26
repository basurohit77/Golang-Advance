package ossrunactions

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

var TestAction1 = newRunAction("TestAction1", false, nil)
var TestAction2 = newRunAction("Test-Action2", false, TestAction1)
var TestAction3 = newRunAction("Test-Action3", false, nil)

func setup() {
	/*
		allValidRunActionsMap = make(map[string]*RunAction)
		allValidRunActionsList = nil
		allValidRunActionsNames = nil
	*/
	for _, ra := range ListValidRunActions() {
		ra.enabled = false
	}
}

func TestEnableDisable(t *testing.T) {
	setup()
	err := Enable([]string{"testaction1", "test-action3"})
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, TestAction1.Name(), true, TestAction1.IsEnabled())
	testhelper.AssertEqual(t, TestAction2.Name(), false, TestAction2.IsEnabled())
	testhelper.AssertEqual(t, TestAction3.Name(), true, TestAction3.IsEnabled())

	testhelper.AssertEqual(t, TestAction1.Name()+" parent", (*RunAction)(nil), TestAction1.Parent())
	testhelper.AssertEqual(t, TestAction2.Name()+" parent", TestAction1, TestAction2.Parent())

	err = Disable([]string{"test-action3"})
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, TestAction1.Name()+"after disable", true, TestAction1.IsEnabled())
	testhelper.AssertEqual(t, TestAction2.Name()+"after disable", false, TestAction2.IsEnabled())
	testhelper.AssertEqual(t, TestAction3.Name()+"after disable", false, TestAction3.IsEnabled())
}

func TestEnableEmpty(t *testing.T) {
	setup()
	err := Enable([]string{""})
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, TestAction1.Name(), false, TestAction1.IsEnabled())
	testhelper.AssertEqual(t, TestAction2.Name(), false, TestAction2.IsEnabled())
	testhelper.AssertEqual(t, TestAction3.Name(), false, TestAction3.IsEnabled())
}

func TestEnableInvalid(t *testing.T) {
	setup()
	err := Enable([]string{"testaction1", "xyz", "test-action2"})
	testhelper.AssertEqual(t, "error from Enable()", fmt.Sprintf(`Invalid RunAction(s): ["xyz"] -- allowed values: %q`, ListValidRunActionNames()), err.Error())

	testhelper.AssertEqual(t, TestAction1.Name(), false, TestAction1.IsEnabled())
	testhelper.AssertEqual(t, TestAction2.Name(), false, TestAction2.IsEnabled())
	testhelper.AssertEqual(t, TestAction3.Name(), false, TestAction3.IsEnabled())
}

func TestRegisterPanicDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got none")
		} else {
			testhelper.AssertEqual(t, "", "Found duplicate ossrunactions.RunAction: Test-Action2 (test-action2)", r)
		}
	}()

	setup()

	newRunAction(TestAction2.Name(), false, nil)
}
