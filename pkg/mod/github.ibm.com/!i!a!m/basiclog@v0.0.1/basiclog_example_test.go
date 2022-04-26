package basiclog_test

import (
	"fmt"
	"log"
	"os"

	basiclog "github.ibm.com/IAM/basiclog"
)

func ExampleLogger() {

	logger, err := basiclog.NewBasicLogger(basiclog.LevelDebug, os.Stdout)

	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Debug")
	logger.Info("Info")
	logger.Error("Error")

	// Output: DEBUG: basiclog_example_test.go:19: Debug
	// INFO: basiclog_example_test.go:20: Info
	// ERROR: basiclog_example_test.go:21: Error

}

func ExampleLogger_Error() {

	logger, err := basiclog.NewBasicLogger(basiclog.LevelNotSet, os.Stdout)

	if logger != nil {
		log.Fatal("invalid logger should be nil")
	}

	fmt.Println(err.Error())
	// Output: invalid logging level 0 provided, use Logging level definitions LevelDebug, LevelInfo, or LevelError
}
