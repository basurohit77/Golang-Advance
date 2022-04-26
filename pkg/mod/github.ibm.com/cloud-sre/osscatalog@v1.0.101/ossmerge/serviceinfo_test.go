package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestIsDeletable(t *testing.T) {
	si := &ServiceInfo{}
	globalMergePhase = mergePhaseServicesTwo
	si.OSSMergeControl = ossmergecontrol.New("oss-catalog-testing")
	si.OSSValidation = ossvalidation.New("", "test-timestamp")
	si.OSSValidation.AddSource("oss-catalog-testing", ossvalidation.PRIOROSS)
	si.OSSService.ReferenceResourceName = "oss-catalog-testing"
	si.OSSValidation.SetSourceNameCanonical("oss-catalog-testing")
	si.mergeWorkArea.mergePhase = mergePhaseFinalized
	testhelper.AssertEqual(t, "empty record", true, si.IsDeletable())
	si.OSSMergeControl.Notes = "something"
	testhelper.AssertEqual(t, "record with merge control notes", true, si.IsDeletable())
	si.OSSMergeControl.Notes = ""
	testhelper.AssertEqual(t, "empty record(2)", true, si.IsDeletable())
	si.OSSValidation.AddSource("oss-catalog-testing", ossvalidation.CATALOG)
	testhelper.AssertEqual(t, "record with one source", false, si.IsDeletable())
}
