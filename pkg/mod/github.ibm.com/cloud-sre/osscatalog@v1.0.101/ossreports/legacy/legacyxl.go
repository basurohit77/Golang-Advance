package legacy

import (
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

type recordType struct {
	CRNServiceName          string `column:"CRN service-name,15"`
	DisplayName             string `column:"Display Name,20"`
	EntryType               string `column:"Type"`
	LegacyEntryType         string `column:"Legacy Type"`
	OperationalStatus       string `column:"Status,10"`
	LegacyOperationalStatus string `column:"Legacy Status,10"`
	CRNValidationStatus     string `column:"CRN Validation Status,10"`
	LegacyValidationStatus  string `column:"Legacy Validation Status,18"`
	LegacyExceptions        string `column:"Legacy Exceptions,10"`
	LegacyNotes             string `column:"Legacy Notes,10"`
	CatalogName             string `column:"Catalog Name\n(if not CRN),15"`
	ServiceNowName          string `column:"ServiceNow Name\n(if not CRN),15"`
	ScorecardV1Name         string `column:"ScorecardV1 Name\n(if not CRN),15"`
	IAMName                 string `column:"IAM Name\n(if not CRN),15"`
	Category                string `column:"Category,15"`
	OfferingManager         string `column:"Offering Manager"`
	OSSTags                 string `column:"OSS Tags"`
}

// RunReport generates a summary report for all services/components from a merge, comparing it with the legacy (python-based) CRN Validation report
func RunReport(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	var err error
	data := make([]interface{}, 0, 500)

	var handler = func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		oss := &si.OSSService
		r := recordType{}
		r.CRNServiceName = string(oss.ReferenceResourceName)
		r.DisplayName = oss.ReferenceDisplayName
		r.EntryType = oss.GeneralInfo.EntryType.ShortString()
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		r.CRNValidationStatus = oss.GeneralInfo.OSSTags.GetCRNStatus().StringStatus()
		if cat := si.GetSourceMainCatalog(); cat != nil {
			if len(si.AdditionalMainCatalog) > 0 {
				r.CatalogName = "*MULTIPLE*"
			} else if !ossmerge.CompareCompositeAndCanonicalName(cat.Name, oss.ReferenceResourceName) {
				r.CatalogName = cat.Name
			} else {
				r.CatalogName = "="
			}
			r.Category = oss.CatalogInfo.CategoryTags
		} else if si.IgnoredMainCatalog != nil {
			r.CatalogName = fmt.Sprintf("*HIDDEN/INACTIVE*(%s)", si.IgnoredMainCatalog.Name)
		} else {
			r.CatalogName = "*MISSING*"
		}
		if sn := si.GetSourceServiceNow(); sn != nil {
			if r.CRNServiceName != sn.CRNServiceName {
				r.ServiceNowName = sn.CRNServiceName
			} else {
				r.ServiceNowName = "="
			}
		} else {
			r.ServiceNowName = "*MISSING*"
		}
		if sc := si.GetSourceScorecardV1Detail(); sc != nil {
			if r.CRNServiceName != sc.Name {
				r.ScorecardV1Name = sc.Name
			} else {
				r.ScorecardV1Name = "="
			}
		} else {
			r.ScorecardV1Name = "*MISSING*"
		}
		if iam := si.GetSourceIAM(); iam != nil {
			if !ossmerge.CompareCompositeAndCanonicalName(iam.Name, oss.ReferenceResourceName) {
				r.IAMName = iam.Name
			} else {
				r.IAMName = "="
			}
		} else {
			r.IAMName = "*MISSING*"
		}
		r.OfferingManager = oss.Ownership.OfferingManager.String()
		if si.Legacy != nil {
			if si.Legacy.EntryType.ShortString() == r.EntryType {
				r.LegacyEntryType = "="
			} else if si.Legacy.EntryType == "" {
				//				r.LegacyEntryType = "*MISSING*"
			} else {
				r.LegacyEntryType = si.Legacy.EntryType.ShortString()
			}
			if string(si.Legacy.OperationalStatus) == r.OperationalStatus {
				r.LegacyOperationalStatus = "="
			} else if si.Legacy.OperationalStatus == "" {
				//				r.LegacyOperationalStatus = "*MISSING*"
			} else {
				r.LegacyOperationalStatus = string(si.Legacy.OperationalStatus)
			}
			r.LegacyExceptions = si.Legacy.Exceptions
			r.LegacyNotes = si.Legacy.Notes
			r.LegacyValidationStatus = si.Legacy.ValidationStatus
		}
		r.OSSTags = oss.GeneralInfo.OSSTags.String()
		data = append(data, r)
	}

	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].(recordType).CRNServiceName < data[j].(recordType).CRNServiceName
	})

	err = ossreports.GenerateExcel(w, data, 65.0, "LegacyXL")
	if err != nil {
		return err
	}

	return nil
}
