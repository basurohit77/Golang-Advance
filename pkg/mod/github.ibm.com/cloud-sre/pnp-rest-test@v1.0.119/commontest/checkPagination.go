package commontest

import (
	"errors"

	"github.ibm.com/cloud-sre/pnp-rest-test/common"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

// PageCheck is a structure that contains only pagination information
type PageCheck struct {
	common.Pagination
}

// CheckPagination will verify that pagination looks right
func CheckPagination(fct, objectLabel string, server *rest.Server, firstPageURL string) (err error) {

	METHOD := fct + "->CheckPagination"

	checkList := new(PageCheck)
	err = server.GetAndDecode(METHOD, objectLabel, firstPageURL, checkList)
	if err != nil {
		return lg.Err(METHOD, err, "Failed to get very first pagination link %s", firstPageURL)
	}

	count := checkList.Count
	limit := checkList.Limit

	if checkList.Offset != 0 {
		return lg.Err(METHOD, nil, "Failed to match offset (0!=%d) on first url pagination link %s", checkList.Offset, firstPageURL)
	}

	if nil != isSame(METHOD, objectLabel, firstPageURL, checkList.First.Href, server) { // The input URL should be the first page
		return lg.Err(METHOD, nil, "Failed go get a match (%s) on first page URL and input URL for %s", objectLabel, firstPageURL)
	}
	if nil != isSame(METHOD, objectLabel, firstPageURL, checkList.Href, server) { // The first page should be the same as href
		return lg.Err(METHOD, nil, "Failed go get a match (%s) on first page URLs for %s", objectLabel, firstPageURL)
	}

	next := checkList.Next.Href
	for next != "" {

		checkList = new(PageCheck)
		err = server.GetAndDecode(METHOD, objectLabel+"->next", next, checkList)
		if err != nil {
			return lg.Err(METHOD, err, "Failed to get next url pagination link %s", next)
		}

		if checkList.Count != count {
			return lg.Err(METHOD, nil, "Failed to match count (%d!=%d) on next url pagination link %s", checkList.Count, count, next)
		}
		if checkList.Limit != limit {
			return lg.Err(METHOD, nil, "Failed to match limit (%d!=%d) on next url pagination link %s", checkList.Limit, limit, next)
		}
		if checkList.Offset == 0 {
			return lg.Err(METHOD, nil, "Failed to get an offset  (%d) on next url pagination link %s", checkList.Offset, next)
		}

		if nil != isSame(METHOD, objectLabel+"->next", next, checkList.Href, server) { // The first page should be the same as href
			return lg.Err(METHOD, nil, "Failed to match (%s) on next page URL for %s", objectLabel, next)
		}

		next = checkList.Next.Href
	}

	return nil
}

func isSame(fct, label, url1, url2 string, server *rest.Server) (err error) {

	METHOD := fct + "->isSame"

	result := true

	checkList1 := new(PageCheck)
	err = server.GetAndDecode(METHOD, label, url1, checkList1)
	if err != nil {
		return lg.Err(METHOD, err, "Failed to retrieve url1 during comparison %s", url1)
	}

	checkList2 := new(PageCheck)
	err = server.GetAndDecode(METHOD, label, url2, checkList2)
	if err != nil {
		return lg.Err(METHOD, err, "Failed to retrieve url2 during comparison %s", url2)
	}

	// it is ok for hrefs to be different
	//result = result && checkList1.Href == checkList2.Href
	result = result && checkList1.Offset == checkList2.Offset
	result = result && checkList1.Limit == checkList2.Limit
	result = result && checkList1.Count == checkList2.Count
	result = result && checkList1.First == checkList2.First
	result = result && checkList1.Last == checkList2.Last
	result = result && checkList1.Previous == checkList2.Previous
	result = result && checkList1.Next == checkList2.Next

	if result {
		return nil
	}

	return errors.New("mismatch of urls")
}
