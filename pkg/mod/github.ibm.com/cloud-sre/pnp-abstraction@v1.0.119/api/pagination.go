package api

// Moved from PNP Status
import (
	"strconv"
	"strings"
)

// Pagination defines the attributes a pagination can send back to the client
type Pagination struct {
	Href   string   `json:"href"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
	Count  int      `json:"count"`
	First  *URLImpl `json:"first"`
	Last   *URLImpl `json:"last"`
	Prev   *URLImpl `json:"previous,omitempty"`
	Next   *URLImpl `json:"next,omitempty"`
}

// IBMPagination defines the attributes a pagination can send back to the client
// Follows IBM Cloud API Handbook for offset and limit pagination https://test.cloud.ibm.com/docs/api-handbook?topic=api-handbook-pagination#offset-and-limit-pagination
type IBMPagination struct {
	Href   string   `json:"href"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
	Count  int      `json:"total_count"`
	First  *URLImpl `json:"first"`
	Last   *URLImpl `json:"last"`
	Prev   *URLImpl `json:"previous,omitempty"`
	Next   *URLImpl `json:"next,omitempty"`
}

// URLImpl defines the href of a resource
type URLImpl struct {
	Href string `json:"href"`
}

const (
	// Limit is the default number of records that are returned per page in the pagination
	Limit = 25
	// MaxLimit is the maximum number of records that are returned per page in the pagination
	MaxLimit = 300
)

var (
	limit  = Limit
	offset = 0

	// APIOSSURL is the URL to reach the API
	APIOSSURL = "https://api-oss.bluemix.net"
	// CatalogClientID holds the base path after APIOSSURL
	CatalogClientID = ""
)

// CreatePagination sets up pagination in a server response to a client
func CreatePagination(inputLimit, inputOffset, numberOfRecords int, apiPath, query string) *Pagination {

	pagination := internalCreatePagination(inputLimit, inputOffset, numberOfRecords, apiPath, query, "v1")
	return pagination.(*Pagination)
}

// CreateIBMPagination sets up pagination in a server response to a client
func CreateIBMPagination(inputLimit, inputOffset, numberOfRecords int, apiPath, query string) *IBMPagination {

	pagination := internalCreatePagination(inputLimit, inputOffset, numberOfRecords, apiPath, query, "v2")
	return pagination.(*IBMPagination)
}

// internalCreatePagination sets up pagination in a server response to a client
func internalCreatePagination(inputLimit, inputOffset, numberOfRecords int, apiPath, query, version string) interface{} {

	limit, offset := limit, offset

	if inputLimit > 0 {
		limit = inputLimit
	}

	if inputOffset > 0 {
		offset = inputOffset
	}

	var returnPagination interface{}

	href := APIOSSURL + CatalogClientID + apiPath +
		(func() string {
			if query == "" {
				return ""
			}
			return "?" + query
		})()

	if query != "" {
		splitQuery := strings.Split(query, "&")

		query = ""

		for _, q := range splitQuery {
			if !strings.HasPrefix(q, "offset") && !strings.HasPrefix(q, "limit") {
				if query != "" {
					query += "&" + q
				} else {
					query += q
				}
			}
		}
		if query != "" {
			query += "&"
		}
	}

	baseURL := APIOSSURL + CatalogClientID + apiPath + "?" + query

	first := "limit=" + strconv.Itoa(limit)

	lOffset := numberOfRecords - numberOfRecords%limit

	if lOffset == numberOfRecords && numberOfRecords > 0 {
		lOffset--
	}

	last := "offset=" + strconv.Itoa(lOffset) + "&" + first

	var pagination Pagination
	pagination.Href = href
	pagination.Count = numberOfRecords
	pagination.Offset = offset
	pagination.Limit = limit
	pagination.First = paginationURL(baseURL + first)
	pagination.Last = paginationURL(baseURL + last)
	//The next link MUST be included for all pages except the last page
	if offset+limit < numberOfRecords {
		n := offset + limit
		next := "offset=" + strconv.Itoa(n) + "&" + first
		pagination.Next = paginationURL(baseURL + next)
	}

	//The prev link SHOULD be included for all pages except the first page.
	if offset != 0 {
		p := offset - limit
		if p < 0 {
			p = 0
		}
		prev := "offset=" + strconv.Itoa(p) + "&" + first
		pagination.Prev = paginationURL(baseURL + prev)
	}

	// v1 refers to struct Pagination
	if version == "v1" {
		returnPagination = &pagination

	}
	// v2 refers to struct IBMPagination
	if version == "v2" {
		var paginationIBM = IBMPagination{
			Href:   pagination.Href,
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Count:  pagination.Count,
			First:  pagination.First,
			Last:   pagination.Last,
			Prev:   pagination.Prev,
			Next:   pagination.Next,
		}
		returnPagination = &paginationIBM
	}
	return returnPagination
}

// need to instantiate/create each of the next, prev, first and last URL's for the pagination
func paginationURL(url string) *URLImpl {
	var hrefURL URLImpl
	hrefURL.Href = url

	return &hrefURL
}
