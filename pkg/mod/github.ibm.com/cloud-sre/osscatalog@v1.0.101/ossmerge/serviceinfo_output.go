package ossmerge

import (
	"fmt"
	"strings"
)

// Dump returns a string summary representation of this ServiceInfo record
func (si *ServiceInfo) Dump() string {
	buf := strings.Builder{}

	buf.WriteString(fmt.Sprintf(`ServiceInfo(%s)`, si.ComparableName))
	if si.DuplicateOf != "" {
		buf.WriteString(fmt.Sprintf(`  DuplicateOf="%s"`, si.DuplicateOf))
	}
	buf.WriteString(fmt.Sprintf(`  CatalogMain="%s"`, si.SourceMainCatalog.Name))
	if len(si.AdditionalMainCatalog) > 0 {
		buf.WriteString("+[")
		for _, r := range si.AdditionalMainCatalog {
			buf.WriteString(fmt.Sprintf(`"%s",`, r.Name))
		}
		buf.WriteString("]")
	}

	buf.WriteString(fmt.Sprintf(`  ScorecardV1="%s"`, si.SourceScorecardV1Detail.Name))
	if len(si.AdditionalScorecardV1Detail) > 0 {
		buf.WriteString("+[")
		for _, r := range si.AdditionalScorecardV1Detail {
			buf.WriteString(fmt.Sprintf(`"%s",`, r.Name))
		}
		buf.WriteString("]")
	}

	buf.WriteString(fmt.Sprintf(`  ServiceNow="%s"`, si.SourceServiceNow.CRNServiceName))
	if len(si.AdditionalServiceNow) > 0 {
		buf.WriteString("+[")
		for _, r := range si.AdditionalServiceNow {
			buf.WriteString(fmt.Sprintf(`"%s",`, r.CRNServiceName))
		}
		buf.WriteString("]")
	}

	if si.OSSMergeControl != nil && !si.OSSMergeControl.IsEmpty() {
		buf.WriteString("   OSSMergeControl=<non-empty>")
		if si.OSSMergeControl.Notes != "" {
			buf.WriteString(fmt.Sprintf(`"%.20s"`, si.OSSMergeControl.Notes))
		}
	}

	if si.IgnoredMainCatalog != nil {
		buf.WriteString(fmt.Sprintf(`  IgnoredMainCatalog="%s"`, si.IgnoredMainCatalog.Name))
	}

	return buf.String()
}

// Header returns a short header representing this ServiceInfo record, suitable for printing in logs and reports
func (si *ServiceInfo) Header() string {
	result := strings.Builder{}
	dupof := si.GetDupOfServiceName()
	if dupof != "" {
		result.WriteString(fmt.Sprintf("DUPLICATE-OF:%s  ", dupof))
	}
	result.WriteString(si.OSSService.Header())
	if si.OSSValidation != nil {
		result.WriteString(si.OSSValidation.Header())
	}
	return result.String()
}

// Details returns a long text representing this ServiceInfo record (including a Header), suitable for printing in logs and reports
func (si *ServiceInfo) Details() string {
	result := strings.Builder{}
	dupof := si.GetDupOfServiceName()
	if dupof != "" {
		result.WriteString(fmt.Sprintf("DUPLICATE-OF:%s  ", dupof))
	}
	result.WriteString(si.OSSService.Header())
	if si.OSSValidation != nil {
		result.WriteString(si.OSSValidation.Details())
	}
	return result.String()
}
