package ossmerge

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/iam"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"
	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

// checkAllAdditionalSources checks if we have more than one record from each source
// (Main Catalog, ServiceNow, ScorecardV1) and ensures that we select the "best"
// primary record from each source, using a stable and repetable algorithm
func (si *ServiceInfo) checkAllAdditionalSources() {

	// Note we must go through sources in this order (Catalog, ServiceNow, ScorecardV1)
	// because we use that logic to determine the most likely "primary" record,
	// and later to compute the ReferenceResourceName (in mergeReferenceResourceName())

	/*
		//		if si.PriorOSS.ReferenceResourceName == "cli-repo" {
		if strings.HasPrefix(si.SourceServiceNow.CRNServiceName, "iam") {
			previousDebugFlags := debug.SetDebugFlags(debug.Fine)
			debug.Debug(debug.Fine, `checkAllAdditionalSources(%s) enter: PriorOSS="%s"`, si.String(), si.PriorOSS.ReferenceResourceName)
			defer func() { debug.SetDebugFlags(previousDebugFlags) }()
		}
	*/

	si.checkAdditionalMainCatalog()

	si.checkAdditionalServiceNow()

	si.checkAdditionalScorecardV1()

	si.checkAdditionalIAM()
}

func (si *ServiceInfo) checkAdditionalMainCatalog() {
	debug.Debug(debug.Fine, `checkAdditionalMainCatalog(%s) enter: SourceMainCatalog="%s"   AdditionalMainCatalog=%s`, si.String(), si.SourceMainCatalog.String(), debugPrintAdditionalNames(si.AdditionalMainCatalog))
	if len(si.AdditionalMainCatalog) == 0 {
		return
	}

	srcs := make([]*catalogapi.Resource, len(si.AdditionalMainCatalog), len(si.AdditionalMainCatalog)+1)
	copy(srcs, si.AdditionalMainCatalog)
	if si.HasSourceMainCatalog() {
		x := *si.GetSourceMainCatalog() // Must make a copy of the record, because we may later overwrite the original record inside ServiceInfo
		srcs = append(srcs, &x)
	}
	sort.SliceStable(srcs, func(i, j int) bool {
		if srcs[i].IsPublicVisible() && !srcs[j].IsPublicVisible() {
			return true
		}
		if !srcs[i].IsPublicVisible() && srcs[j].IsPublicVisible() {
			return false
		}
		if srcs[i].EffectiveVisibility.Restrictions == string(catalogapi.VisibilityPublic) && srcs[j].EffectiveVisibility.Restrictions != string(catalogapi.VisibilityPublic) {
			return true
		}
		if srcs[i].EffectiveVisibility.Restrictions != string(catalogapi.VisibilityPublic) && srcs[j].EffectiveVisibility.Restrictions == string(catalogapi.VisibilityPublic) {
			return false
		}
		if srcs[i].EffectiveVisibility.Restrictions == string(catalogapi.VisibilityIBMOnly) && srcs[j].EffectiveVisibility.Restrictions != string(catalogapi.VisibilityIBMOnly) {
			return true
		}
		if srcs[i].EffectiveVisibility.Restrictions != string(catalogapi.VisibilityIBMOnly) && srcs[j].EffectiveVisibility.Restrictions == string(catalogapi.VisibilityIBMOnly) {
			return false
		}
		if si.HasPriorOSS() {
			if srcs[i].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return true
			} else if srcs[j].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return false
			}
		}
		return srcs[i].Name < srcs[j].Name
	})

	si.SourceMainCatalog = *srcs[0]
	si.AdditionalMainCatalog = srcs[1:]
	debug.Debug(debug.Fine, `checkAdditionalMainCatalog(%s) exit: SourceMainCatalog="%s"   AdditionalMainCatalog=%s`, si.String(), si.SourceMainCatalog.String(), debugPrintAdditionalNames(si.AdditionalMainCatalog))
}

func (si *ServiceInfo) checkAdditionalServiceNow() {
	debug.Debug(debug.Fine, `checkAdditionalServiceNow(%s) enter: SourceServiceNow="%s"   AdditionalServiceNow=%s`, si.String(), si.SourceServiceNow.String(), debugPrintAdditionalNames(si.AdditionalServiceNow))
	if len(si.AdditionalServiceNow) == 0 {
		return
	}

	srcs := make([]*servicenow.ConfigurationItem, len(si.AdditionalServiceNow), len(si.AdditionalServiceNow)+1)
	copy(srcs, si.AdditionalServiceNow)
	if si.HasSourceServiceNow() {
		x := *si.GetSourceServiceNow() // Must make a copy of the record, because we may later overwrite the original record inside ServiceInfo
		srcs = append(srcs, &x)
	}
	sort.SliceStable(srcs, func(i, j int) bool {
		if !srcs[i].IsRetired() {
			if srcs[j].IsRetired() {
				return true
			}
		} else if !srcs[j].IsRetired() {
			return false
		}
		if si.HasPriorOSS() {
			if srcs[i].CRNServiceName == string(si.GetPriorOSS().ReferenceResourceName) {
				return true
			} else if srcs[j].CRNServiceName == string(si.GetPriorOSS().ReferenceResourceName) {
				return false
			}
		}
		if si.HasSourceMainCatalog() {
			if srcs[i].CRNServiceName == si.GetSourceMainCatalog().Name {
				return true
			} else if srcs[j].CRNServiceName == si.GetSourceMainCatalog().Name {
				return false
			}
		}
		return srcs[i].CRNServiceName < srcs[j].CRNServiceName
	})

	si.SourceServiceNow = *srcs[0]
	si.AdditionalServiceNow = srcs[1:]
	debug.Debug(debug.Fine, `checkAdditionalServiceNow(%s) exit: SourceServiceNow="%s"   AdditionalServiceNow=%s`, si.String(), si.SourceServiceNow.String(), debugPrintAdditionalNames(si.AdditionalServiceNow))
}

func (si *ServiceInfo) checkAdditionalScorecardV1() {
	debug.Debug(debug.Fine, `checkAdditionalScorecardV1(%s) enter: SourceScorecardV1="%s"   AdditionalScorecardV1=%s`, si.String(), si.SourceScorecardV1Detail.String(), debugPrintAdditionalNames(si.AdditionalScorecardV1Detail))
	if len(si.AdditionalScorecardV1Detail) == 0 {
		return
	}

	srcs := make([]*scorecardv1.DetailEntry, len(si.AdditionalScorecardV1Detail), len(si.AdditionalScorecardV1Detail)+1)
	copy(srcs, si.AdditionalScorecardV1Detail)
	if si.HasSourceScorecardV1Detail() {
		x := *si.GetSourceScorecardV1Detail() // Must make a copy of the record, because we may later overwrite the original record inside ServiceInfo
		srcs = append(srcs, &x)
	}
	sort.SliceStable(srcs, func(i, j int) bool {
		// No check for retired/deleted ScorecardV1 entries (no such thing)

		if si.HasPriorOSS() {
			if srcs[i].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return true
			} else if srcs[j].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return false
			}
		}
		if si.HasSourceMainCatalog() {
			if srcs[i].Name == si.GetSourceMainCatalog().Name {
				return true
			} else if srcs[j].Name == si.GetSourceMainCatalog().Name {
				return false
			}
		}
		if si.HasSourceServiceNow() {
			if srcs[i].Name == si.GetSourceServiceNow().CRNServiceName {
				return true
			} else if srcs[j].Name == si.GetSourceServiceNow().CRNServiceName {
				return false
			}
		}
		return srcs[i].Name < srcs[j].Name
	})

	si.SourceScorecardV1Detail = *srcs[0]
	si.AdditionalScorecardV1Detail = srcs[1:]
	debug.Debug(debug.Fine, `checkAdditionalScorecardV1(%s) exit: SourceScorecardV1="%s"   AdditionalScorecardV1=%s`, si.String(), si.SourceScorecardV1Detail.String(), debugPrintAdditionalNames(si.AdditionalScorecardV1Detail))
}

func (si *ServiceInfo) checkAdditionalIAM() {
	debug.Debug(debug.Fine, `checkAdditionalIAM(%s) enter: SourceIAM="%s"   AdditionalIAM=%s`, si.String(), si.SourceIAM.String(), debugPrintAdditionalNames(si.AdditionalIAM))
	if len(si.AdditionalIAM) == 0 {
		return
	}

	srcs := make([]*iam.Service, len(si.AdditionalIAM), len(si.AdditionalIAM)+1)
	copy(srcs, si.AdditionalIAM)
	if si.HasSourceIAM() {
		x := *si.GetSourceIAM() // Must make a copy of the record, because we may later overwrite the original record inside ServiceInfo
		srcs = append(srcs, &x)
	}
	sort.SliceStable(srcs, func(i, j int) bool {
		if srcs[i].Enabled {
			if !srcs[j].Enabled {
				return true
			}
		} else if srcs[j].Enabled {
			return false
		}
		if si.HasPriorOSS() {
			if srcs[i].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return true
			} else if srcs[j].Name == string(si.GetPriorOSS().ReferenceResourceName) {
				return false
			}
		}
		if si.HasSourceMainCatalog() {
			if srcs[i].Name == si.GetSourceMainCatalog().Name {
				return true
			} else if srcs[j].Name == si.GetSourceMainCatalog().Name {
				return false
			}
		}
		// XXX Should we also look for names from ServiceNow and Scorecard?
		// Not much relevant for IAM, which links more closely to Catalog
		return srcs[i].Name < srcs[j].Name
	})

	si.SourceIAM = *srcs[0]
	si.AdditionalIAM = srcs[1:]
	debug.Debug(debug.Fine, `checkAdditionalIAM(%s) exit: SourceIAM="%s"   AdditionalIAM=%s`, si.String(), si.SourceIAM.String(), debugPrintAdditionalNames(si.AdditionalIAM))
}

func debugPrintAdditionalNames(list interface{}) string {
	buf := strings.Builder{}
	first := true
	buf.WriteString("[")
	listVal := reflect.ValueOf(list)
	listLen := listVal.Len()
	for i := 0; i < listLen; i++ {
		if !first {
			buf.WriteString("  ")
		} else {
			first = false
		}
		buf.WriteString(fmt.Sprintf(`%v`, listVal.Index(i)))
	}
	buf.WriteString("]")
	return buf.String()
}
