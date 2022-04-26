/* Package tlog implements utility routines for all services. This library is for debuging use only;
   displays the relative path of the package, funtion name and line number at the ouput logs.
   Set an environment varible DEBGUG = "false" in the Makefile and Jenkins file if you don't want to see the full
   line description

   DEBUG = true or not set at all, will display something like [pnp-abstraction/db/read.go->GetResourceByQuerySimple:4139]
   DEBUG = false will display something like [GetResourceByQuerySimple:4139]
*/

package tlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.ibm.com/cloud-sre/oss-globals/consts"
)

// Log Will get the package and function names from the function is called.
func Log() string {
	// If DEBUG environment variable is set at the Makefile, returs [relative path of the package and file name ->funtion name:line number]
	// such as: [cloud-sre/pnp-abstraction/db/insert.go->InsertResource:45] otherwise will return  [funtion name:line number], like: [InsertResource:45]

	//    package: "github.ibm.com/cloud-sre/pnp-abstraction/utils"

	//    Use: log.Println(tlog.Log()+"Some message here ")

	debugFlag := strings.ToLower(os.Getenv("DEBUG"))
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fileName := frame.File[strings.Index(frame.File, consts.IBMgitHubEndPoint)+len(consts.IBMgitHubEndPoint) : len(frame.File)]
	funcName := filepath.Base(frame.Function)
	funcName = funcName[strings.Index(funcName, ".")+1:]
	if debugFlag == "false" {
		return fmt.Sprintf("[%s:%d] ", funcName, frame.Line)
	}
	return fmt.Sprintf("[%s->%s:%d] ", fileName, funcName, frame.Line)
}

// FuncName returns the name of the function where FuncName is called
func FuncName() string {

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	funcName := filepath.Base(frame.Function)
	funcName = funcName[strings.Index(funcName, ".")+1:]

	return funcName
}
