package initadapter

import (
	"os"
	"testing"

	"github.ibm.com/cloud-sre/pnp-notifications-adapter/cloudant"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
)

func TestInitialize(t *testing.T) {
	Initialize()
}

func TestInitialize2(t *testing.T) {
	os.Setenv(cloudant.AccountID, "accountID")
	os.Setenv(cloudant.AccountPW, "accountPW")
	os.Setenv(cloudant.NotificationsURLEnv, "NotificationsURL")
	os.Setenv(cloudant.ServicesURLEnv, "ServicesURL")
	os.Setenv(cloudant.RuntimesURLEnv, "RuntimesURL")
	os.Setenv(cloudant.PlatformURLEnv, "PlatformURL")
	os.Setenv(globalcatalog.GCResourceURL, "GCURL")
	os.Setenv(osscatalog.OSSCatalogCatYPKeyLabel, "catkeylabel")
	os.Setenv(osscatalog.OSSCatalogCatYPKeyValue, "catkeyvalue")
	Initialize()
}
