package cos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"strconv"

	"github.com/pkg/errors"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

func main() {
	isUploaded, err := Upload("", "", "./README.MD", "text/plain")
	fmt.Println(isUploaded, err)
}

const (
	CONTENT_TYPE_JSON string = "application/json"
	CONTENT_TYPE_TEXT string = "text/plain"
	CONTENT_TYPE_LOG  string = "text/x-log"
	CONTENT_TYPE_XLSX string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	COS_IAM_KEY       string = "cos-iam-key"
)

func ContentType(fileExtension string) string {
	switch fileExtension {
	case "xlsx":
		return CONTENT_TYPE_XLSX
	case "txt":
		return CONTENT_TYPE_TEXT
	case "json":
		return CONTENT_TYPE_JSON
	}
	return CONTENT_TYPE_TEXT
}

func TestConnection(endpoint, bucketName string) (bool, error) {
	if endpoint == "" {
		return false, errors.New("cos endpoint is required param.")
	}
	if bucketName == "" {
		return false, errors.New("cos bucket name is required param.")
	}
	apiKey, err := rest.GetKey(COS_IAM_KEY)
	if apiKey == "" {
		return false, errors.New("cos_api_key is required variable.")
	}
	accessToken, err := iamAccessToken(apiKey)
	if err != nil {
		return false, err
	}
	headers := make(map[string]string)
	headers["Authorization"] = "bearer " + accessToken
	result, code, err := fireHTTPRequest("https://"+endpoint+"/"+bucketName, headers, http.MethodGet, bytes.NewReader([]byte("")))
	if err != nil {
		fmt.Println("failed to get objects from bucket.", err)
		return false, err
	}
	if code < 200 || code > 399 {
		fmt.Println("failed to get objects from bucket. http statu code is ", code)
		return false, errors.New(string(result))
	}
	fmt.Println("Test COS connection passed:", bucketName)
	return true, nil
}
func Delete(endpoint, bucketName, fileName string) (bool, error) {
	if endpoint == "" {
		return false, errors.New("cos endpoint is required param.")
	}
	if bucketName == "" {
		return false, errors.New("cos bucket name is required param.")
	}
	if fileName == "" {
		return false, errors.New("fileName is required param.")
	}
	apiKey, err := rest.GetKey(COS_IAM_KEY)
	if apiKey == "" {
		return false, errors.New("cos_api_key is required variable.")
	}
	accessToken, err := iamAccessToken(apiKey)
	if err != nil {
		return false, err
	}
	cosEndpoint := "https://" + endpoint + "/" + bucketName + "/" + fileName
	cosHeaders := make(map[string]string)
	cosHeaders["Authorization"] = "bearer " + accessToken
	result, code, err := fireHTTPRequest(cosEndpoint, cosHeaders, http.MethodDelete, bytes.NewReader([]byte("")))
	if err != nil {
		fmt.Println("failed to delete file to COS.", err)
		return false, err
	}
	if code < 200 || code > 399 {
		if result != nil {
			fmt.Println(string(result))
		}
		return false, errors.New("bad http response code:" + strconv.Itoa(code))
	}
	return true, nil
}
func Download(endpoint, bucketName, fileName string) ([]byte, error) {
	if endpoint == "" {
		return nil, errors.New("cos endpoint is required param.")
	}
	if bucketName == "" {
		return nil, errors.New("cos bucket name is required param.")
	}
	if fileName == "" {
		return nil, errors.New("fileName is required param.")
	}
	apiKey, err := rest.GetKey(COS_IAM_KEY)
	if apiKey == "" {
		return nil, errors.New("cos_api_key is required variable.")
	}
	accessToken, err := iamAccessToken(apiKey)
	if err != nil {
		return nil, err
	}
	cosEndpoint := "https://" + endpoint + "/" + bucketName + "/" + fileName
	cosHeaders := make(map[string]string)
	cosHeaders["Authorization"] = "bearer " + accessToken
	fileContent, code, err := fireHTTPRequest(cosEndpoint, cosHeaders, http.MethodGet, bytes.NewReader([]byte("")))
	if err != nil {
		fmt.Println("failed to download file to COS.", err)
		return nil, err
	}
	if code < 200 || code > 399 {
		if fileContent != nil {
			fmt.Println(string(fileContent))
		}
		return nil, errors.New("bad http response code:" + strconv.Itoa(code))
	}
	return fileContent, nil
}
func Overwrite(endpoint, bucketName, filePath, contentType string) (bool, error) {
	return upload(endpoint, bucketName, filePath, contentType, true)
}
func upload(endpoint, bucketName, filePath, contentType string, overwrite bool) (bool, error) {
	if endpoint == "" {
		return false, errors.New("cos endpoint is required param.")
	}
	if bucketName == "" {
		return false, errors.New("cos bucket name is required param.")
	}
	if filePath == "" {
		return false, errors.New("filePath is required param.")
	}
	filePath = filepath.Clean(filePath)
	if !fileExists(filePath) {
		return false, errors.New(filePath + " does not exist (or is a directory).")
	}
	if contentType == "" {
		contentType = "text/plain"
		fmt.Println("using default contentType for file :", filePath, contentType)
	}
	apiKey, err := rest.GetKey(COS_IAM_KEY)
	if apiKey == "" {
		return false, errors.New("cos_api_key is required variable.")
	}
	accessToken, err := iamAccessToken(apiKey)
	if err != nil {
		return false, err
	}
	fileName := filepath.Base(filePath)
	// try delete it first
	if overwrite {
		Delete(endpoint, bucketName, fileName)
	}
	cosEndpoint := "https://" + endpoint + "/" + bucketName + "/" + fileName
	cosHeaders := make(map[string]string)
	cosHeaders["Authorization"] = "bearer " + accessToken
	cosHeaders["Content-Type"] = contentType
	theFileContent, err := ioutil.ReadFile(filePath) // #nosec
	bytes.NewReader(theFileContent)
	result, code, err := fireHTTPRequest(cosEndpoint, cosHeaders, http.MethodPut, bytes.NewReader(theFileContent))
	if err != nil {
		fmt.Println("failed to upload file to COS.", err)
		return false, err
	}
	if code < 200 || code > 399 {
		if result != nil {
			fmt.Println(string(result))
		}
		return false, errors.New(string(result))
	}
	return true, nil
}

// filePath : ./osscatpublisher-ro-log-2019-12-05T2340Z.log
// contentType : text/plain
func Upload(endpoint, bucketName, filePath, contentType string) (bool, error) {
	return upload(endpoint, bucketName, filePath, contentType, false)
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func fireHTTPRequest(url string, headers map[string]string, method string, bodyReader *bytes.Reader) ([]byte, int, error) {
	if url == "" {
		return nil, 0, fmt.Errorf("url is required")
	}
	if method == "" {
		return nil, 0, fmt.Errorf("method is required")
	}
	/* #nosec */
	transport := &http.Transport{DisableKeepAlives: true}
	req, _ := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	for k, v := range headers {
		if k == "BasicAuth" {
			usernamePwdPair := strings.Split(v, "::")
			if len(usernamePwdPair) == 2 {
				u := usernamePwdPair[0]
				p := usernamePwdPair[1]
				req.SetBasicAuth(u, p)
			}
			continue
		}
		req.Header.Set(k, v)
	}
	res, err := transport.RoundTrip(req)
	if err != nil {
		return nil, 0, err
	}
	if transport != nil {
		defer transport.CloseIdleConnections()
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		return data, res.StatusCode, err
	}
	return nil, 0, err
}

func iamAccessToken(apiKey string) (string, error) {
	iamHeaders := make(map[string]string)
	iamHeaders["Accept"] = "application/json"
	iamHeaders["Content-Type"] = "application/x-www-form-urlencoded"
	body := "apikey=" + apiKey + "&response_type=cloud_iam&grant_type=urn:ibm:params:oauth:grant-type:apikey" // pragma: whitelist secret
	var mapResult map[string]interface{}
	bodyReader := bytes.NewReader([]byte(body))
	var iamTokenURL string
	if os.Getenv("IAM_URL") != "" {
		iamTokenURL = os.Getenv("IAM_URL") + "/identity/token"
	} else {
		iamTokenURL = "https://iam.cloud.ibm.com/identity/token" // #nosec G101
	}
	result, code, err := fireHTTPRequest(iamTokenURL, iamHeaders, http.MethodPost, bodyReader)
	if err != nil {
		fmt.Println("failed to get IAM token.", err)
		return "", err
	}
	if code < 200 || code > 399 {
		if result != nil {
			fmt.Println(string(result))
		}
		return "", errors.New(string(result))
	}
	err = json.Unmarshal([]byte(result), &mapResult)
	if err != nil {
		fmt.Println("failed to unmarshal IAM token.", string(result), err)
		return "", err
	}
	accessToken := mapResult["access_token"].(string)
	return accessToken, nil
}
