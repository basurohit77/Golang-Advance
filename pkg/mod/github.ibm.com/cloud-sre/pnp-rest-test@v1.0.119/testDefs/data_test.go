package testDefs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

func TestData(t *testing.T) {

	actual := db.CreateRecordIDFromString("globalCatalog" + "crn:v1:bluemix:public:cloud-object-storage:us-south::::" + "crn:v1:bluemix:public:cloud-object-storage:us-south::::")
	assert.Equal(t, "19e6aa1c78037a183b583404116f730370c8415f0cf9ce09e2783d495a6f0f1d", actual, "CreateRecordIDFromString: Does not match") // # pragma: whitelist secret
	actual = db.CreateRecordIDFromSourceSourceID("globalCatalog", "crn:v1:bluemix:public:cloud-object-storage:us-east::::")
	assert.Equal(t, "ce8b8ae171f34d666a3b6164eb61886477d7baa715418ab7818900d246bb0c59", actual, "CreateRecordIDFromSourceSourceID: Does not match") // # pragma: whitelist secret
	actual = db.CreateRecordIDFromString("servicenow" + "INC0311852" + "crn:v1:bluemix:public:cloud-object-storage:us-east::::")
	assert.Equal(t, "e4111d2511e2a426adc158e1e4f28f144055c6bbc4aedd94f09691077d607f08", actual, "CreateRecordIDFromString: Does not match") // # pragma: whitelist secret
}
