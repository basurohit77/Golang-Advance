package db

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
)

type DbConfiguration struct {
	Host                string
	Port                int
	Database            string
	User                string
	Password            string
	SslMode             string
	SslRootCertFilePath string
}

var Config *DbConfiguration
var configStruct DbConfiguration

// GetDatabaseConfiguration - return the database information string used to connect to PostGreSQL using the GO database driver gopkg.in/lib/pq.v0
func GetDatabaseConfiguration(pathname string) string {
	// read the configure file into bytes
	configJson, err := ioutil.ReadFile(filepath.Clean(pathname))
	if err != nil {
		// this would be an installation defect
		log.Fatal("Fail to read database configuration file. ", err)
	}
	return GetDatabaseConfiguration2(configJson)
}

// GetDatabaseConfiguration2 - return the database information string used to connect to PostGreSQL using the GO database driver gopkg.in/lib/pq.v0
func GetDatabaseConfiguration2(configJson []byte) string {
	// parse JSON
	err := json.Unmarshal(configJson, &Config)
	if err != nil {
		// this would be a programming defect
		log.Fatal("Fail to unmarshal database configuration. ", err)
	}

	return "host=" + Config.Host +
		" port=" + strconv.Itoa(Config.Port) +
		" dbname=" + Config.Database +
		" user=" + Config.User +
		" password=" + Config.Password +
		" sslmode=" + Config.SslMode
}

// GetDatabaseConfiguration3 - return the database information string used to connect to PostGreSQL using the GO database driver gopkg.in/lib/pq.v0
func GetDatabaseConfiguration3(host string, port int, database string, user string, password string, sslmode string) string {

	Config = &configStruct

	Config.Host = host
	Config.Port = port
	Config.Database = database
	Config.User = user
	Config.Password = password
	Config.SslMode = sslmode

	return "host=" + Config.Host +
		" port=" + strconv.Itoa(Config.Port) +
		" dbname=" + Config.Database +
		" user=" + Config.User +
		" password=" + Config.Password +
		" sslmode=" + Config.SslMode
}

// GetDatabaseConfigurationWithSSLRootCert - return the database information string used to connect to PostGreSQL that requires SSLROOTCERT
// Note that sslrootcertFilePath value is the file path of the SSLROOTCERT certificate
func GetDatabaseConfigurationWithSSLRootCert(host string, port int, database string, user string, password string, sslMode string, sslrootcertFilePath string) string {

	Config = &configStruct

	Config.Host = host
	Config.Port = port
	Config.Database = database
	Config.User = user
	Config.Password = password
	Config.SslMode = sslMode
	Config.SslRootCertFilePath = sslrootcertFilePath

	return "host=" + Config.Host +
		" port=" + strconv.Itoa(Config.Port) +
		" dbname=" + Config.Database +
		" user=" + Config.User +
		" password=" + Config.Password +
		" sslmode=" + Config.SslMode +
		" sslrootcert=" + Config.SslRootCertFilePath
}
