package handlers

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-nq2ds/testutils"
)

var (
	resourceReturned = `{
    "category_id":"test",
    "record_id":"%s",
    "source":"%s",
    "source_id":"%s",
    "status":"%s",
    "crn_full":"%s",
    "pnp_creation_time":"",
    "pnp_update_time":"",
    "source_creation_time":"",
    "source_update_time":"",
    "state":"",
    "operational_status":"",
    "status_update_time":"",
    "regulatory_domain":"",
    "category_id":"",
    "category_parent":false,
    "record_hash":"",
    "displayName":[{"name":"test status","language":"en"}],
    "visibility": ["test"],
    "tags": [{"id":"tag1"}]
    }`
	recordID = db.CreateRecordIDFromSourceSourceID(testutils.Source, testutils.SourceID)
)

func Test_decryptionError(t *testing.T) {
	log.Println(tlog.Log())
	var testData = `{"test":"test"}`
	dbConn, _, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, []byte(testData), &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have decryption error.")
}

func Test_unmarshalError(t *testing.T) {
	log.Println(tlog.Log())
	var testData = `{"test":"test"`
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have unmarshal error.")
}

func Test_invalidResouce_source(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(resourceReturned, recordID, "", testutils.SourceID, status, crnFull)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, "Should have validation error.")
}

func Test_invalidResouce_sourceID(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(resourceReturned, recordID, testutils.Source, "", status, crnFull)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have validation error.")
}

func Test_invalidResouce_status(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(resourceReturned, recordID, testutils.Source, testutils.SourceID, "", crnFull)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have validation error.")
}

func Test_invalidResouce_crn(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(resourceReturned, recordID, testutils.Source, testutils.SourceID, status, "")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have validation error.")
}

func Test_invalidResouce_recordid(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(resourceReturned, "123", testutils.Source, testutils.SourceID, status, crnFull)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have validation error.")
}

func Test_processStatus(t *testing.T) {
	log.Println(tlog.Log())
	status := "complete"
	var testData = fmt.Sprintf(resourceReturned, recordID, testutils.Source, testutils.SourceID, status, "crn:v1:bluemix:public:testService:eu-gb::::")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	isBadMessage := ProcessStatus(dbConn, msg, &mon)
	assert.False(t, isBadMessage, tlog.Log()+"Should have no error.")
	testData = fmt.Sprintf(resourceReturned, recordID, "badresource", testutils.SourceID, status, "crn:v1:bluemix:public:testService:eu-gb::::")
	dbConn, msg, mon = testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	isBadMessage = ProcessStatus(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"should have validation error")
	log.Println(tlog.Log() + "Completed")
}
