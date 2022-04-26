package servicenow

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// TestGetServiceNowCallInfo tests the URL returned by the getServiceNowCallInfo function
func TestGetServiceNowCallInfo(t *testing.T) {
	url, _ := getServiceNowCallInfo("", 0, 0, true, WATSON)
	testhelper.AssertEqual(t, "watson test endpoint", serviceNowWatsonTestURLv2, url)

	url, _ = getServiceNowCallInfo("", 0, 0, false, WATSON)
	testhelper.AssertEqual(t, "watson production endpoint", serviceNowWatsonURLv2, url)

	url, _ = getServiceNowCallInfo("", 0, 0, true, CLOUDFED)
	testhelper.AssertEqual(t, "cloudfed test endpoint", serviceNowCloudfedTestURLv2, url)

	url, _ = getServiceNowCallInfo("", 0, 0, false, CLOUDFED)
	testhelper.AssertEqual(t, "cloudfed production endpoint", serviceNowCloudfedURLv2, url)
}

// TestGetTokenName tests the token name returned by the getTokenName function
func TestGetTokenName(t *testing.T) {
	tokenName := getTokenName(true, WATSON)
	testhelper.AssertEqual(t, "watson test", serviceNowWatsonTestTokenName, tokenName)

	tokenName = getTokenName(false, WATSON)
	testhelper.AssertEqual(t, "watson production", serviceNowWatsonTokenName, tokenName)

	tokenName = getTokenName(true, CLOUDFED)
	testhelper.AssertEqual(t, "cloudfed test", serviceNowCloudfedTestTokenName, tokenName)

	tokenName = getTokenName(false, CLOUDFED)
	testhelper.AssertEqual(t, "cloudfed production", serviceNowCloudfedTokenName, tokenName)
}
