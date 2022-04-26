package options

import (
	"fmt"
	"os"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// Test functions for logging information that is sent to LogDNA.
// This must happen in this package rather than the "debug" package, to avoid
// a circular package dependency

func TestLoggingWithLogDNA(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestLoggingWithLogDNA() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Fine)
	}

	LoadGlobalOptions("-keyfile DEFAULT -lenient", true)
	debug.SetLogFile(os.Stdout)

	err := BootstrapLogDNA("testing", false, 0)
	testhelper.AssertError(t, err)

	debug.Info("Testing %s", "debug.Info()")

	err = debug.InfoWithOptions("test-oss-id", "Testing %s", "debug.InfoWithOptions()")
	testhelper.AssertError(t, err)

	debug.Audit("Testing %s", "debug.Audit()")

	err = debug.AuditWithOptions("test-oss-id", "Testing %s", "debug.AuditWithOptions()")
	testhelper.AssertError(t, err)

	debug.PrintError("Testing %s", "debug.PrintError()")

	err = debug.PlainLogEntry(debug.LevelINFO, "test-oss-id", "Testing %s\n", "debug.PlainLogEntry()")
	testhelper.AssertError(t, err)

	err = debug.FlushLogDNA()
	testhelper.AssertError(t, err)
}

func DISABLEDTestLogDNABuffering(t *testing.T) {
	if testing.Short() /* && false /*XXX */ {
		t.Skip("Skipping test TestLogDNABuffering() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.LogDNA | debug.Fine /* XXX */)
	}

	LoadGlobalOptions("-keyfile DEFAULT -lenient", true)
	debug.SetLogFile(os.Stdout)

	err := BootstrapLogDNA("testing", false, 10)
	testhelper.AssertError(t, err)

	debug.Info("Testing %s", "First INFO line")

	for i := 0; i < 30; i++ {
		debug.PlainLogEntry(debug.LevelINFO, "testing-oss-id", fmt.Sprintf("Plain log entry %d\n   continued\n", i))
	}

	debug.Audit("Finished writing a series of plain log entries")

	err = debug.FlushLogDNA()
	testhelper.AssertError(t, err)
}
