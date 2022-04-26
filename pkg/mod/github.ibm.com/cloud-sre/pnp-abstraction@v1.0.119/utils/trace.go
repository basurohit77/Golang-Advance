/* Package utils implements utility routines for pnp-abstration services. This library is for debuging use only;
   displays the relative path of the package, funtion name and line number at the ouput logs.
   Set an environment varible DEBGUG = "false" in the Makefile and Jenkins file if you don't want to see the full
   line description

   DEBUG = true or not set at all, will display something like [cloud-sre/pnp-abstraction/db/read.go->GetResourceByQuerySimple:4139]
   DEBUG = false will display something like [GetResourceByQuerySimple:4139]
*/

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IBMgitHubEndPoint IBM Git Hub end point
const IBMgitHubEndPoint = "github.ibm.com/"

/*TraceLog Will get the package and function names from the function is called.
If DEBUG environment variable is set at the Makefile, returs [relative path of the package and file name ->funtion name:line number]
such as: [cloud-sre/pnp-abstraction/db/insert.go->InsertResource:45] otherwise will return  [funtion name:line number], like: [InsertResource:45]

   package: "github.ibm.com/cloud-sre/pnp-abstraction/utils"

   Use: log.Println(utils.TraceLog()+"Some message here ")
*/
func TraceLog() string {

	debugFlag := strings.ToLower(os.Getenv("DEBUG"))
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fileName := frame.File[strings.Index(frame.File, IBMgitHubEndPoint)+len(IBMgitHubEndPoint) : len(frame.File)]
	funcName := filepath.Base(frame.Function)
	funcName = funcName[strings.Index(funcName, ".")+1 : len(funcName)]
	if debugFlag == "false" {
		return fmt.Sprintf("[%s:%d] ", funcName, frame.Line)
	}
	return fmt.Sprintf("[%s->%s:%d] ", fileName, funcName, frame.Line)

}
