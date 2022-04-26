package health

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestCheckCatalogHealth(t *testing.T) {
	httpmock.Activate()

	res := `{"href":"https://api-oss.bluemix.net/catalog/api/catalog/healthz","code":0,"description":"The API is available and operational."}`

	r := httpmock.NewStringResponder(http.StatusOK, res)

	httpmock.RegisterResponder("GET", "http://localhost/health", r)

	c := NewCatalogHealth("http://localhost/health")

	healthy, status := c.CheckCatalogHealth()
	if !healthy {
		t.Fatal("catalog is not healthy")
	}
	if status != "200" {
		t.Fatalf("expected code %s, got %s for catalog health", "200", status)
	}
}
