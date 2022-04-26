package ossmerge

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

type statusCategoryTestSpec struct {
	name           ossrecord.CRNServiceName
	id             string
	parent         string
	si             *ServiceInfo
	expectedIssues []string
}

var statusCategoryTestData = []*statusCategoryTestSpec{
	{"service-1", "category-1", "", nil, nil},

	{"service-2", "category-2", "service-2", nil, []string{
		`(info)     [StatusPage]    This entry is the Status Page Category Parent for its CategoryID: CategoryID="category-2"  number of entries with this CategoryID=1`,
	}},

	{"service-3-a", "category-3", "service-3-b", nil, nil},
	{"service-3-b", "category-3", "service-3-b", nil, []string{
		`(info)     [StatusPage]    This entry is the Status Page Category Parent for its CategoryID: CategoryID="category-3"  number of entries with this CategoryID=3`,
	}},
	{"service-3-c", "category-3", "service-3-b", nil, nil},

	{"service-4-a", "category-4", "service-4-b", nil, []string{
		`(severe)   [StatusPage]    More than one Status Page Category Parent for this CategoryID (voiding): CategoryID="category-4"  parents=[service-4-b <empty>]`,
	}},
	{"service-4-b", "category-4", "service-4-b", nil, []string{
		`(severe)   [StatusPage]    More than one Status Page Category Parent for this CategoryID (voiding): CategoryID="category-4"  parents=[service-4-b <empty>]`,
	}},
	{"service-4-c", "category-4", "", nil, []string{
		`(severe)   [StatusPage]    More than one Status Page Category Parent for this CategoryID (voiding): CategoryID="category-4"  parents=[service-4-b <empty>]`,
	}},

	{"service-5-a", "category-5", "", nil, []string{
		`(severe)   [StatusPage]    Status Page Notification CategoryID is used in more than one entry but Category Parent is blank (voiding): CategoryID="category-5"  number of entries with same ID=2`,
	}},
	{"service-5-b", "category-5", "", nil, []string{
		`(severe)   [StatusPage]    Status Page Notification CategoryID is used in more than one entry but Category Parent is blank (voiding): CategoryID="category-5"  number of entries with same ID=2`,
	}},

	{"service-6-a", "category-6-prime", "service-6-a", nil, []string{
		`(info)     [StatusPage]    This entry is the Status Page Category Parent for its CategoryID: CategoryID="category-6-prime"  number of entries with this CategoryID=1`,
	}},
	{"service-6-b", "category-6", "service-6-a", nil, []string{
		`(severe)   [StatusPage]    Status Page Notification CategoryID does not match the CategoryID of the CategoryParent (voiding): this.CategoryID="category-6"  parent.CategoryID="category-6-prime"`,
	}},

	{"service-7", "category-7", "service7", nil, []string{
		`(severe)   [StatusPage]    Status Page Notification Category Parent is found but does not use the canonical name (voiding): CategoryParent="service7"  CanonicalName="service-7"`,
	}},
}

func (s *statusCategoryTestSpec) creationPhase() {
	ossr := ossrecordextended.NewOSSServiceExtended(s.name)
	s.si, _ = LookupService(MakeComparableName(string(s.name)), true)
	s.si.OSSServiceExtended = *ossr
	s.si.SourceServiceNow.CRNServiceName = string(s.name)
	s.si.OSSService.Compliance.ServiceNowOnboarded = true
	s.si.SourceServiceNow.StatusPage.Group = "Group/" + s.id
	s.si.SourceServiceNow.StatusPage.CategoryID = s.id
	if s.parent != "" {
		s.si.OSSMergeControl.AddOverride("StatusPage.CategoryParent", s.parent)
	}
	s.si.PriorOSS.ReferenceResourceName = s.name
	s.si.PriorOSS.StatusPage.CategoryParent = ""
	s.si.mergeWorkArea.mergePhase = mergePhaseServicesOne
	globalMergePhase = mergePhaseServicesOne

	s.si.mergeStatusPage()
}

func (s *statusCategoryTestSpec) executionPhase(t *testing.T) {
	s.si.mergeWorkArea.mergePhase = mergePhaseServicesTwo
	globalMergePhase = mergePhaseServicesTwo

	s.si.checkStatusCategoryParent()
}

func (s *statusCategoryTestSpec) checkingPhase(t *testing.T) {
	var i = 0
	var diffs bool
	var errors bool
	var actualIssues = make([]string, 0, len(s.si.OSSValidation.Issues))
	for _, issue := range s.si.OSSValidation.Issues {
		//			fmt.Printf("DEBUG: i=%d  len(expected)=%d   len(actual)=%d\n", i, len(expectedIssues), len(si.OSSValidation.Issues))
		if issue.Severity == ossvalidation.IGNORE {
			continue
		}
		if issue.Title == "Status Page Notification Category Parent value overriden with OSS MergeControl record" {
			continue
		}
		if issue.Severity == ossvalidation.SEVERE {
			errors = true
		}
		actualIssues = append(actualIssues, issue.String())
		if i < len(s.expectedIssues) {
			if actualIssues[i] != s.expectedIssues[i]+"\n" {
				diffs = true
			}
		} else {
			diffs = true
		}
		i++
	}
	if len(s.expectedIssues) > i {
		diffs = true
	}
	if diffs {
		t.Errorf("%s: mismatched validation issues", s.si.OSSService.ReferenceResourceName)
		fmt.Printf("FAIL: %s: mismatched validation issues:\n", s.si.OSSService.ReferenceResourceName)
		fmt.Println("  -- Expected:")
		for _, issue := range s.expectedIssues {
			fmt.Printf("       %s\n", issue)
		}
		fmt.Println("  -- Actual:")
		for _, issue := range actualIssues {
			fmt.Printf("       %s", issue)
		}
	}
	if errors {
		testhelper.AssertEqual(t, string(s.si.OSSService.ReferenceResourceName)+".CategoryID", "", s.si.OSSService.StatusPage.CategoryID)
		testhelper.AssertEqual(t, string(s.si.OSSService.ReferenceResourceName)+".CategoryParent", ossrecord.CRNServiceName(""), s.si.OSSService.StatusPage.CategoryParent)
	} else {
		testhelper.AssertEqual(t, string(s.si.OSSService.ReferenceResourceName)+".CategoryID", s.si.SourceServiceNow.StatusPage.CategoryID, s.si.OSSService.StatusPage.CategoryID)
		expectedParent := s.si.OSSMergeControl.Overrides["StatusPage.CategoryParent"]
		if expectedParent != nil {
			testhelper.AssertEqual(t, string(s.si.OSSService.ReferenceResourceName)+".CategoryParent", expectedParent.(ossrecord.CRNServiceName), s.si.OSSService.StatusPage.CategoryParent)
		} else {
			testhelper.AssertEqual(t, string(s.si.OSSService.ReferenceResourceName)+".CategoryParent", ossrecord.CRNServiceName(""), s.si.OSSService.StatusPage.CategoryParent)
		}
	}
}

func TestCheckStatusCategoryParent(t *testing.T) {

	for _, s := range statusCategoryTestData {
		s.creationPhase()
	}
	for _, s := range statusCategoryTestData {
		s.executionPhase(t)
	}
	runAllDeferredFunctions()
	for _, s := range statusCategoryTestData {
		s.checkingPhase(t)
	}
}
