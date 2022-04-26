package common

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SermoDigital/jose"
	log "github.com/sirupsen/logrus"
	h "github.ibm.com/cloud-sre/commontools/http"
)

const (
	devDal10 = "oss-dev-dal10"
	devWdc06 = "oss-dev-wdc06"
	devFra02 = "oss-dev-fra02"

	stagingDal10 = "oss-stage-dal10"
	stagingWdc06 = "oss-stage-wdc06"
	stagingFra02 = "oss-stage-fra02"

	prodDal10 = "oss-prod-dal10"
	prodWdc06 = "oss-prod-wdc06"
	prodFra02 = "oss-prod-fra02"

	devUSEastPath  = VaultEndpoint + VaultPathPrefix + "dev/us-east/KUBECONFIG"
	devUSSouthPath = VaultEndpoint + VaultPathPrefix + "dev/us-south/KUBECONFIG"
	devEUDEPath    = VaultEndpoint + VaultPathPrefix + "dev/eu-de/KUBECONFIG"

	stagingUSEastPath  = VaultEndpoint + VaultPathPrefix + "staging/us-east/KUBECONFIG"
	stagingUSSouthPath = VaultEndpoint + VaultPathPrefix + "staging/us-south/KUBECONFIG"
	stagingEUDEPath    = VaultEndpoint + VaultPathPrefix + "staging/eu-de/KUBECONFIG"

	prodUSEastPath  = VaultEndpoint + VaultPathPrefix + "prod/us-east/KUBECONFIG"
	prodUSSouthPath = VaultEndpoint + VaultPathPrefix + "prod/us-south/KUBECONFIG"
	prodEUDEPath    = VaultEndpoint + VaultPathPrefix + "prod/eu-de/KUBECONFIG"
)

type Payload struct {
	AsUser  bool      `json:"as_user"`
	Mrkdwn  bool      `json:"mrkdwn"`
	Channel string    `json:"channel"`
	Text    string    `json:"text"`
	Blocks  []*Blocks `json:"blocks"`
}
type TextPayload struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type Blocks struct {
	Type string       `json:"type"`
	Text *TextPayload `json:"text"`
}

type LegacyAttachment struct {
	MrkdwnIn   []string       `json:"mrkdwn_in"`
	Color      string         `json:"color"`
	Pretext    string         `json:"pretext"`
	AuthorName string         `json:"author_name"`
	AuthorIcon string         `json:"author_icon"`
	Title      string         `json:"title"`
	Text       string         `json:"text"`
	Fields     []*LegacyField `json:"fields"`
	ThumbURL   string         `json:"thumb_url"`
}

type LegacyField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type LegacySlackMsg struct {
	Channel     string              `json:"channel"`
	Text        string              `json:"text"`
	Attachments []*LegacyAttachment `json:"attachments"`
}

//CryptoRandomHex generate hex random number
func CryptoRandomHex(len int) (string, error) {
	randomBytes := make([]byte, len)
	if _, err := rand.Reader.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// Sample:
// vaultConf := "/login/vault/vault.conf"
// vaultProps, err := ReadINI(vaultConf)
// endpoint := vaultProps["endpoint"]
// if os.IsNotExist(err) || strings.TrimSpace(endpoint) == "" {
// 	beego.Info("getVaultInstance ", err)
// 	panic("please check vault properties")
// }
//ReadINI read ini file into map
func ReadINI(iniPath string) (map[string]string, error) {
	props := make(map[string]string)
	_, err := os.Stat(iniPath)

	if os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(filepath.Clean(iniPath))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineText := scanner.Text()
		if strings.HasPrefix(lineText, ";") { //; means comments
			continue
		}
		kv := strings.Split(lineText, "=")
		if len(kv) > 1 {
			k := kv[0]
			v := kv[1]
			vWithComments := strings.Split(v, ";")
			if len(vWithComments) > 1 {
				v = vWithComments[0]
			}
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			props[k] = v
		}
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return nil, err
	}

	return props, file.Close()
}

func GetEmailFromAccessToken(token *string) (string, error) {
	if token == nil {
		return "", fmt.Errorf("token is empty")
	}

	arr := strings.Split(*token, ".")
	if len(arr) == 0 {
		return "", fmt.Errorf("token is not valid")
	}

	body := arr[1]
	decodeBytes, err := jose.Base64Decode([]byte(body))
	if err != nil {
		return "", err
	}

	var t map[string]interface{}
	err = json.Unmarshal(decodeBytes, &t)
	if err != nil {
		return "", err
	}

	email := t["email"]
	if email != nil {
		return email.(string), err
	}
	return "", nil
}

/*
	blocks sample:

	titleBlock := &Blocks{
		Type: "section",
		Text: &TextPayload{
			Type: "mrkdwn",
			Text: "*Bastionhost bastionhost-sample installation complete*",
		},
	}
	bodyBlock := &Blocks{
		Type: "section",
		Text: &TextPayload{
			Type: "mrkdwn",
			Text: "ClusterID: xxx\nClusterName: xxx\nLogin URL: `tsh login --proxy xxxxx:443`",
		},
	}

	blocks := make([]*Blocks, 0)
	blocks = append(blocks, titleBlock)
	blocks = append(blocks, bodyBlock)
*/
func TriggerSlack(slackToken, channelName, text string, blocks []*Blocks) error {

	if slackToken == "" {
		return fmt.Errorf("slackToken is required")
	}
	if channelName == "" {
		return fmt.Errorf("channelName is required")
	}
	if text == "" {
		return fmt.Errorf("text is required")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
	}
	client := &http.Client{Transport: tr}

	data := &Payload{
		AsUser:  true,
		Mrkdwn:  true,
		Channel: channelName,
		Text:    text,
		Blocks:  blocks,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func TriggerSlackOutdated(slackToken string, legacySlackMsg *LegacySlackMsg) error {

	if slackToken == "" {
		return fmt.Errorf("slackToken is required")
	}
	if legacySlackMsg == nil {
		return fmt.Errorf("legacySlackMsg is required")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
	}
	client := &http.Client{Transport: tr}

	payloadBytes, err := json.Marshal(legacySlackMsg)
	if err != nil {
		return err
	}
	fmt.Println(string(payloadBytes))
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func IsValidEmail(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

func GetFunctionKubeConfig(env string, vaultRoleID string, vaultSecretID string) (string, error) {
	log.Debug("Get kube config of function id.")
	var vaultURL string
	if strings.HasPrefix(env, devDal10) {
		vaultURL = devUSSouthPath
	} else if strings.HasPrefix(env, devWdc06) {
		vaultURL = devUSEastPath
	} else if strings.HasPrefix(env, devFra02) {
		vaultURL = devEUDEPath
	} else if strings.HasPrefix(env, stagingDal10) {
		vaultURL = stagingUSSouthPath
	} else if strings.HasPrefix(env, stagingWdc06) {
		vaultURL = stagingUSEastPath
	} else if strings.HasPrefix(env, stagingFra02) {
		vaultURL = stagingEUDEPath
	} else if strings.HasPrefix(env, prodDal10) {
		vaultURL = prodUSSouthPath
	} else if strings.HasPrefix(env, prodWdc06) {
		vaultURL = prodUSEastPath
	} else if strings.HasPrefix(env, prodFra02) {
		vaultURL = prodEUDEPath
	} else {
		log.WithFields(log.Fields{"cluster": env}).Errorln("Not a valid cluster.")
		return "", fmt.Errorf("not a valid cluster")
	}
	var err error
	var token string
	cacheKey := base64.StdEncoding.EncodeToString([]byte("auth-" + vaultRoleID + "-" + vaultSecretID))
	cacheValue, found := VaultCache.Get(cacheKey)
	if found {
		log.Info("found cached vault auth")
		token = cacheValue.(string)
	} else {
		log.Info("get client token from vault.")
		token, err = vaultToken(cacheKey, vaultRoleID, vaultSecretID)
		if err != nil {
			return "", err
		}
	}
	header := make(map[string]string)
	header["X-Vault-Token"] = token
	log.WithFields(log.Fields{"vaultURL": vaultURL}).Debug("Request to get kube config from vault.")
	md := &h.Metadata{
		URL:     vaultURL,
		Headers: header,
		Method:  http.MethodGet,
		Body:    nil,
		Cookies: nil,
		Timeout: 0,
	}
	result := md.FireHTTPRequest()
	data := result.Data
	err = result.Err
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Failed to get kube config from vault.")
		return "", err
	}

	str := strings.ReplaceAll(string(data), "\\\\n", "")
	var vd *VaultData
	log.Debug("Unmarshal http response data to vaultData.")
	err = json.Unmarshal([]byte(str), &vd)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Failed to unmarshal http response data to vaultData.")
		return "", err
	}

	value := vd.Data.Value
	rawKubeConfig, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Failed to base64 decode.")
		return "", err
	}
	log.Debug("Get kube config of function id successfully.")
	return string(rawKubeConfig), nil
}
