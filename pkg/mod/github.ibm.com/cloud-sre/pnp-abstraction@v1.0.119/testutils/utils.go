package testutils

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

// const catalogAPIPathImpls = "/catalog/api/catalog/impls"
// const catalogCategoryPath = "/catalog/api/catalog/category"
// const catalogCategoriesPath = "/catalog/api/catalog/categories"

//HttpGet Mock httpGet
func HttpGet(t *testing.T, url, expectedResponse string, expectedStatus int, handlerfunc http.HandlerFunc) {

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerfunc)

	handler.ServeHTTP(rr, req)

	// ensure expected status and returned status match
	assert.Equal(t, expectedStatus, rr.Code)

	// ensure expected response and returned response match
	assert.Equal(t, expectedResponse, strings.TrimSuffix(rr.Body.String(), "\n"))

}

// AssertEqual performs a reflect.DeepEqual on the expected (e) and actual(a) values. If the
// check fails, it outputs a colored diff between the two.
func AssertEqual(t *testing.T, e, a interface{}) {
	t.Helper()

	if e != nil && a != nil {
		eKind := reflect.TypeOf(e).Kind()
		aKind := reflect.TypeOf(a).Kind()

		if eKind != aKind {
			t.Errorf("Given arguments are not the same kind. %v and %v given.", eKind, aKind)
			return
		}
	}

	dmp := diffmatchpatch.New()
	scs := spew.ConfigState{
		Indent:                  "\t",
		DisablePointerAddresses: true,
		SortKeys:                true,
	}

	es := scs.Sdump(e)
	as := scs.Sdump(a)

	if !reflect.DeepEqual(e, a) {
		diffs := dmp.DiffMain(es, as, false)
		t.Errorf(dmp.DiffPrettyText(diffs))
	}
}
