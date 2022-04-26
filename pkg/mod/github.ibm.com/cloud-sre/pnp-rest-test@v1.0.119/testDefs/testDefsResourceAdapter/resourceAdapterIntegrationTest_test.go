package resourceAdapterTest

import (
	"database/sql/driver"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-rest-test/config"
)

const (
	// This turns allows the mock servers to work.  It's useful to turn off/set to false if you want to test against actual servers
	isLocal = true
)

func TestRunImportResources(t *testing.T) {
	if isLocal {
		os.Setenv("disableSkipTransport", "true")
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		r1 := httpmock.NewStringResponder(http.StatusOK, "")
		httpmock.RegisterResponder("GET", "http://api-pnp-resource-adapter/importResources", r1)
	}
	err := runImportResources()
	assert.Nil(t, err, "Should be nil")
}

func TestDeleteResources(t *testing.T) {
	const FCT = "TestDeleteResources: "

	var (
		err  error
		mock sqlmock.Sqlmock
	)

	if isLocal {
		config.Pdb, mock, err = sqlmock.New()
		mock.ExpectExec("DELETE .*").WillReturnResult(sqlmock.NewResult(1, 1))

	} else {
		config.Pdb, err = db.Connect(config.PgHost, 5432, config.PgDB, config.PgDBUser, config.PgPass, "disable")
		if err != nil {
			log.Print(err)
		}
	}
	log.Print("main db: ", config.Pdb)

	err, _ = deleteResources(config.Pdb)

	assert.Nil(t, err)
	log.Println(FCT+" err: ", err)
}

func TestDeleteResource(t *testing.T) {
	const FCT = "TestDeleteResource: "

	var (
		err  error
		mock sqlmock.Sqlmock
	)

	if isLocal {
		config.Pdb, mock, err = sqlmock.New()
		mock.ExpectExec("DELETE .*").WillReturnResult(sqlmock.NewResult(1, 1))

	} else {
		config.Pdb, err = db.Connect(config.PgHost, 5432, config.PgDB, config.PgDBUser, config.PgPass, "disable")
		if err != nil {
			log.Print(err)
		}
	}
	log.Print("main db: ", config.Pdb)

	err, _ = deleteResource(config.Pdb, "record_id")

	assert.Nil(t, err)
	log.Println(FCT+" err: ", err)
}

func TestResourceAdapter(t *testing.T) {
	const FCT = "[TestResourceAdapter]"
	var (
		err      error
		mock     sqlmock.Sqlmock
		recordID = "1234"
	)
	var isLocal = true
	if isLocal {
		config.Pdb, mock, err = sqlmock.New()
		rows := sqlmock.NewRows([]string{"record_id", "source_id"})
		var rowVals = []driver.Value{recordID, "source_id_value"}
		rows.AddRow(rowVals...)
		mock.ExpectQuery(".*").
			WillReturnRows(rows)
	} else {
		config.Pdb, err = db.Connect(config.PgHost, 5432, config.PgDB, config.PgDBUser, config.PgPass, "disable")
		if err != nil {
			log.Print(err)
		}
	}
	log.Println(FCT, "main db:", config.Pdb)

	recordID1, _, err, _ := GetRandomResource(config.Pdb)
	if err != nil {
		log.Println(FCT, "Error getting Resource:", err.Error())
	}
	assert.Equal(t, recordID, recordID1)
}
