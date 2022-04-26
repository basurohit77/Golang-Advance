package helper

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

func TestCompare(t *testing.T) {

	newTime := "2018-10-31T12:12:12Z"
	oldTime := "2018-10-31T12:12:12Z"

	c, err := CompareTime(newTime, time.RFC3339, oldTime, time.RFC3339)
	if err != nil || c != 0 {
		t.Fatal("Wrong time value.  Not equal")
	}

	newTime = "2018-11-01T12:12:12Z"
	c, err = CompareTime(newTime, "", oldTime, time.RFC3339)
	if err != nil || c <= 0 {
		t.Fatal("Wrong time value.  Not greater than")
	}

	oldTime = "2018-11-02T12:12:12Z"
	c, err = CompareTime(newTime, time.RFC3339, oldTime, "")
	if err != nil || c >= 0 {
		t.Fatal("Wrong time value.  Not less than")
	}

	newTime = "Some wierd string"
	c, err = CompareTime(newTime, time.RFC3339, oldTime, time.RFC3339)
	if err != nil || c >= 0 {
		t.Fatal("Wrong time value with wierd string.  Not less than")
	}

	oldTime = "Some wierd string"
	newTime = "2018-11-01T12:12:12Z"
	c, err = CompareTime(newTime, time.RFC3339, oldTime, time.RFC3339)
	if c <= 0 {
		t.Fatal("Wrong time value with wierd string.  Not greater than")
	}

	newTime = "Some wierd string"
	c, err = CompareTime(newTime, time.RFC3339, oldTime, time.RFC3339)
	if c != 0 {
		t.Fatal("Wrong time value wierd string.  Not equal")
	}

	newTime = "2019-04-24 21:49:26"
	oldTime = "2019-04-24 21:48:57Z0500"
	c, err = CompareTime(newTime, "2006-01-02 15:04:05", oldTime, time.RFC3339)
	if c <= 0 {
		t.Fatal(c, "Wrong time value wierd string.  Not equal")
	}

	newTime = "2019-04-24 21:49:56-05"
	oldTime = "2019-04-25 02:49:57+00"
	c, err = CompareTime(newTime, "2006-01-02 15:04:05", oldTime, time.RFC3339)
	if c > -1 {
		t.Fatal(c, "newTime should be less than oldtime", newTime, oldTime)
	}

	newTime = "2019-04-25 02:49:58-00"
	oldTime = "2019-04-24 21:49:57-05"
	c, err = CompareTime(newTime, "2006-01-02 15:04:05", oldTime, "")
	if c < 1 {
		t.Fatal(c, "newTime should be newer than oldtime", newTime, oldTime)
	}

	newTime = "2019-04-24 17:49:57-00"
	oldTime = "2019-04-24 12:49:57-05"
	c, err = CompareTime(newTime, "", oldTime, "")
	if c != 0 {
		t.Fatal(c, "newTime should be equal to oldtime", newTime, oldTime)
	}
}

func TestInArrayObj(t *testing.T) {
	str := "aa"
	strArray := []string{"aa", "bb"}
	exist := InArrayObj(str, strArray)
	assert.True(t, cmp.Equal(exist, true))
}

func TestRecordInWatchArray(t *testing.T) {
	watches := []datastore.WatchReturn{}
	err := json.Unmarshal([]byte(watches_JSON), &watches)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	exist, pos := RecordInWatchArray("3f344815-2793-4780-9b1b-a5658a5dee20", watches)
	assert.True(t, cmp.Equal(exist, true))
	assert.True(t, cmp.Equal(pos, 0))

	exist, pos = RecordInWatchArray("3f344815-2793-4780-9b1b-a5658a5dee34", watches)
	assert.True(t, cmp.Equal(exist, false))
	assert.True(t, cmp.Equal(pos, -1))
}
func TestRemoveWatchFromArray(t *testing.T) {
	i := 1
	watches := []datastore.WatchReturn{}
	err := json.Unmarshal([]byte(watches_JSON), &watches)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	watchesRes := RemoveWatchFromArray(watches, i)
	log.Print(GetPrettyJson(watchesRes))

}

func TestSubscriptionInWatchArray(t *testing.T) {
	watches := []datastore.WatchReturn{}
	err := json.Unmarshal([]byte(watches_JSON), &watches)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}

	exist := SubscriptionInWatchArray("https://pnp-subconsumer.us-east.dev.cloud-oss.bluemix.net/api/v1/pnp/subscriptions/77df6e86-d600-4e72-81fe-cad26cccde5f", watches)
	assert.True(t, cmp.Equal(exist, true))

	exist = SubscriptionInWatchArray("https://pnp-subconsumer.us-east.dev.cloud-oss.bluemix.net/api/v1/pnp/subscriptions/77df6e86-d600-4e72-81fe-cad26cccde11", watches)
	assert.True(t, cmp.Equal(exist, false))
}

func TestRemoveSubscriptionReturnFromArray(t *testing.T) {
	subArr := []datastore.SubscriptionReturn{}
	err := json.Unmarshal([]byte(sampleSubs), &subArr)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	subsRes := RemoveSubscriptionReturnFromArray(subArr, 1)
	log.Print(GetJson(subsRes))
}

func TestRestarter(t *testing.T) {
	Restarter(1, "aa", testFun)
}

func testFun() {
	log.Print("TestRestarter...")
}

func TestGetJson(t *testing.T) {
	testdata := `{"name": "aa"`
	log.Print(GetJson(testdata))
}
func TestUnmarshalAnything(t *testing.T) {
	unmarshalRes := UnmarshalAnything(watches_JSON)
	log.Print(GetJson(unmarshalRes))
}

func TestUnmarshalAnything_err(t *testing.T) {
	testdata := `{"name": "aa"}`
	unmarshalRes := UnmarshalAnything(testdata)
	log.Print(unmarshalRes)
}

func Test(t *testing.T) {
	current := time.Now()
	currentStr := current.UTC().Format("2006-01-02T15:04:05Z")
	resBool := IsNewTimeAfterExistingTime("2006-01-02T15:04:05Z", time.RFC3339, currentStr, time.RFC3339)
	assert.True(t, cmp.Equal(resBool, false))
}

func TestIsNewTimeBeforeExistingTime(t *testing.T) {
	current := time.Now()
	currentStr := current.UTC().Format("2006-01-02T15:04:05Z")
	resBool := IsNewTimeBeforeExistingTime("2006-01-02T15:04:05Z", time.RFC3339, currentStr, time.RFC3339)
	assert.True(t, cmp.Equal(resBool, true))

	current = time.Now()
	currentStr = current.UTC().Format("2006-01-02T15:04:05Z")
	currentPlusOneMin := current.Add(time.Minute * 1)
	currentPlusOneMinStr := currentPlusOneMin.UTC().Format("2006-01-02T15:04:05Z")
	resBool = IsNewTimeBeforeExistingTime(currentPlusOneMinStr, time.RFC3339, currentStr, time.RFC3339)
	assert.True(t, cmp.Equal(resBool, false))

	// Test using format received in servicenow incidents (for example, 2019-05-23 16:33:39Z) [note "Z" is added at end by nq2ds code]:

	current = time.Now()
	currentStr = current.UTC().Format("2006-01-02T15:04:05Z")
	resBool = IsNewTimeBeforeExistingTime("2019-05-23 16:33:39Z", "2006-01-02 15:04:05Z", currentStr, time.RFC3339)
	assert.True(t, cmp.Equal(resBool, true))

	current = time.Now()
	currentStr = current.UTC().Format("2006-01-02T15:04:05Z")
	resBool = IsNewTimeBeforeExistingTime("2100-05-23 16:33:39Z", "2006-01-02 15:04:05Z", currentStr, time.RFC3339)
	assert.True(t, cmp.Equal(resBool, false))
}

var watches_JSON = `[{"record_id":"3f344815-2793-4780-9b1b-a5658a5dee20","subscription":{"href":"https://pnp-subconsumer.us-east.dev.cloud-oss.bluemix.net/api/v1/pnp/subscriptions/77df6e86-d600-4e72-81fe-cad26cccde5f"},"kind":"case","path":"/testpath/2","crns":["null"],"recordIDToWatch":["CS0052652"]},{"record_id":"c8ebd47f-d58a-4c88-8c6f-04473796fab9","subscription":{"href":"https://pnp-subconsumer.us-east.dev.cloud-oss.bluemix.net/api/v1/pnp/subscriptions/6bced181-9c4d-449e-af8a-26088db11fee"},"kind":"case","path":"/testpath/b","crns":["null"],"recordIDToWatch":["CS0052652"]},{"record_id":"df6ecf70-785a-48bf-b719-afe5ee7edaa4","subscription":{"href":"http://pnp-subscriptions.us-east.dev.cloud-oss.bluemix.net/api/v1/pnp/subscriptions/6bced181-9c4d-449e-af8a-26088db11fee"},"kind":"case","crns":["null"],"recordIDToWatch":["CS0052652"]}]`

var sampleSubs = `[{"record_id":"abcdd181-9c4d-449e-af8a-26088db11fee","name":"testSubs","targetAddress":"http://api-pnp-labserver","targetToken":"","expiration":""},{"record_id":"cf8571a3-01cb-46f5-bb82-1dff5c4d45be","name":"name 1","targetAddress":"http://hello","targetToken":"","expiration":""}]`
