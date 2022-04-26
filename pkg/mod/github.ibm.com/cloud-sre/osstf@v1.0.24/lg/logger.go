package lg

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"
)

var logFileDirectory string = "/var/log/"

func SetLogFileDirectory(directoryName string) {
	logFileDirectory = directoryName
}

func Setup(folderName string, fileName string, maxSize int, maxBackups int, maxAge int) {

	logFolder := logFileDirectory + folderName // put log files under here
	// 0666 file mode is for rw-rw-rw-
	// ensure log folder is ready
	if _, err := os.Stat(logFolder); err != nil { // Get info for the folder, if exists, err==nil
		if os.IsNotExist(err) { // folder doesn't exist, will attempt to create
			if err = os.Mkdir(logFolder, 0600); err != nil { // couldn't create the folder -- usually requries super user authority
				log.Println("Please create a folder " + logFolder + " and restart the API Catalog")
				log.Fatal("Failed to create folder "+logFolder+": ", err)
			} else {
				log.Println(logFolder + " created")
			}
		} else { // something is wrong, cannot access log folder
			log.Fatal("Failed to get "+logFolder+" info: ", err)
		}
		// } else {
		// log.Println("Folder " + logFolder + " already exist")
	}

	logFile := filepath.Join(logFolder, fileName)
	_, err := os.Stat(logFile)
	if err != nil { // Get info for the file, if exists, err==nil
		if os.IsNotExist(err) { // file doesn't exist, will attempt to create
			if _, err = os.Create(filepath.Clean(logFile)); err != nil { // couldn't create the file -- probably requries super user authority
				log.Println("Please create an empty writable file " + logFile + " and restart the API Catalog")
				log.Fatal("Failed to create "+logFile+": ", err)
			} else {
				if _, err = os.Stat(logFile); err != nil {
					log.Fatal("Filed to get new "+logFile+" info: ", err)
				}
			}
		} else {
			// something wrong, cannot access log file
			log.Fatal("Failed to get info for "+logFile+": ", err)
		}
		// } else {
		// log.Println(logFile + " already exist")
	}

	// Ensure we have write permission
	// if info != nil {
	// log.Println(logFile + " permission is :" + info.Mode().String())
	// }
	f, err := os.OpenFile(filepath.Clean(logFile), os.O_WRONLY, 0600)
	if err != nil {
		if os.IsPermission(err) {
			if err := os.Chmod(logFile, 0600); err != nil {
				log.Fatal("Failed to grant "+logFile+" write permission", err)
			} else {
				log.Println(logFile + " permission granted")
			}
		}
		// } else {
		// log.Println(logFile + " already have read&write permission")
	}
	_ = f.Close()

	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	})
}

// An identifier to associate a pair of Request and Response.
// The int is incremented for each Reqeust
var count = 1

// METHOD =====================> COMPONENT
func SendRequest(componentName string, req *http.Request) int {
	id := count
	count++

	indent1 := strconv.Itoa(id) + " >   "
	indent2 := strconv.Itoa(id) + " >       "

	Log(strconv.Itoa(id), componentName, " =====================> ", req.Method)

	logRequest(req, indent1, indent2, false)

	return id
}

// COMPONENT <--------------------- STATUS
func ReceiveResponse(id int, componentName string, resp *http.Response, body []byte) {

	indent1 := strconv.Itoa(id) + " <   "
	indent2 := strconv.Itoa(id) + " <       "

	if resp == nil {
		Log(strconv.Itoa(id), componentName, " <--------------------- NO RESPONSE")
		Log(indent1, "Client may be unavailable or bad address")
	} else {
		Log(strconv.Itoa(id), componentName, " <--------------------- ", strconv.Itoa(resp.StatusCode))
		logResponse(resp, body, indent1, indent2)
	}
}

// COMPONENT <===================== METHOD
func ReceiveRequest(componentName string, req *http.Request) int {
	id := count
	count++

	indent1 := strconv.Itoa(id) + " <   "
	indent2 := strconv.Itoa(id) + " <       "

	Log(strconv.Itoa(id), componentName, " <===================== ", req.Method)

	logRequest(req, indent1, indent2, true)

	return id
}

// COMPONENT ---------------------> STATUS
func SendResponse(id int, componentName string, statusCode int, body []byte, headers http.Header) {

	indent1 := strconv.Itoa(id) + " >   "
	indent2 := strconv.Itoa(id) + " >       "

	Log(strconv.Itoa(id))
	Log(strconv.Itoa(id), componentName, " ---------------------> ", strconv.Itoa(statusCode))

	if body != nil {
		Log(indent1, "Body:")
		Log(indent2, string(body))

	}

	logHeaders(headers, indent1, indent2)
}

// COMPONENT ---------------------> STATUS
func SendResponseError(id int, componentName string, statusCode int, err string) {

	indent1 := strconv.Itoa(id) + " >   "
	indent2 := strconv.Itoa(id) + " >       "

	Log(strconv.Itoa(id))
	Log(strconv.Itoa(id), componentName, " ---------------------> ", strconv.Itoa(statusCode))

	Log(indent1, "Error:")
	Log(indent2, err)
}

func logRequest(req *http.Request, indent1, indent2 string, incoming bool) {

	Log(indent1, "URL:")
	Log(indent2, req.URL.String())
	if incoming {
		Log(indent1, "RemoteAddr:")
		Log(indent2, req.RemoteAddr)
	}

	if req.Body != nil {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			Log("logRequest: ERROR could not read body ", err.Error())
		}
		_ = req.Body.Close()

		rdr := bytes.NewReader(body)
		req.Body = ioutil.NopCloser(rdr)

		Log(indent1, "Body:")
		Log(indent2, string(body))

	}

	logHeaders(req.Header, indent1, indent2)
}

func logResponse(resp *http.Response, body []byte, indent1, indent2 string) {

	if resp.Request != nil {
		Log(indent1, "URL:")
		Log(indent2, resp.Request.URL.String())
	}

	if body != nil {
		Log(indent1, "Body:")
		Log(indent2, string(body))
	}

	logHeaders(resp.Header, indent1, indent2)
}

func logHeaders(header http.Header, indent1, indent2 string) {

	if header != nil && len(header) > 0 {
		Log(indent1, "Headers:")

		for key, values := range header {
			h := key + " = "
			for _, value := range values {
				if key == "Authorization" {
					value = "****" // Don't write sensitive information to the log
				}
				h += value + ";"
			}
			Log(indent2, " - ", h)
		}
	}
}

func OpenDatabse(i int) {
	Log(">>>>>>>>>> OPEN  DATABASE :: ", strconv.Itoa(i), ":: <<<<<<<<<<")
}

func CloseDatabase(i int) {
	Log("<<<<<<<<<< CLOSE DATABASE :: ", strconv.Itoa(i), ":: >>>>>>>>>>")
}

//func OpenRows(i int, j int){
//	Log(">.>.>.>.>. OPEN  DB ROWS :: ", strconv.Itoa(i), ".",  strconv.Itoa(j), ":: <.<.<.<.<.")
//}
//
//func CloseRows(i int, j int){
//	Log("<.<.<.<.<. CLOSE DB ROWS :: ", strconv.Itoa(i), ".",  strconv.Itoa(j), ":: >.>.>.>.>.")
//}

func SqlQuery(source string, query string, err error) {
	id := count
	count++

	indent := strconv.Itoa(id) + " >   "

	arrow1 := " <===================== "
	arrow2 := " ---------------------> "

	Log(strconv.Itoa(id), "SELECT", arrow1, source)
	Log(indent, query)
	Log(strconv.Itoa(id), "SELECT", arrow2, source)

	if err != nil {
		Log(indent, err.Error())
	}
}

func SqlExecuteStatement(statement string, result sql.Result, err error) {
	SqlExecuteTable("SQL", "", statement, result, err)
}

func SqlExecuteTable(verb string, target string, statement string, result sql.Result, err error) {
	id := count
	count++

	indent := strconv.Itoa(id) + " >   "

	Log(strconv.Itoa(id), verb, " =====================> ", target)
	Log(indent, statement)
	Log(strconv.Itoa(id), verb, " <--------------------- ", target)
	if result != nil {
		lastInsertId, _ := result.LastInsertId()
		Log(indent, "Last Insert ID: ", strconv.FormatInt(int64(lastInsertId), 10))
		rowsAffected, _ := result.RowsAffected()
		Log(indent, "Rows Affected:  ", strconv.FormatInt(int64(rowsAffected), 10))
	}
	if err != nil {
		Log(indent, "Error: ", err.Error())
	}
}

func SqlError(query string, err error) {
	id := count
	count++

	Log(strconv.Itoa(id), "SQL   ====================== ", query)
	Log(strconv.Itoa(id), "ERROR ---------------------- ", err.Error())
}

func Log(str ...string) {

	l := ""
	for _, s := range str {
		l += " " + s
	}

	log.Println(l)
}
