package ossmerge

import (
	"fmt"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// NameGroup represents one group of ServiceInfo records and/or ClearingHouse Entries that share a common name,
// directly or indirecly. For example:
//   - Entry A has names N1 and N2
//   - Entry B has names N2 and N3
//   - Entry C has names N3 and N4
// --> all three entries A, B and C are part of the same NameGroup
type NameGroup struct {
	OSSNames         []ossrecord.CRNServiceName    // CRN-sn of one ServiceInfo entry
	CHIDs            []clearinghouse.DeliverableID // ID of one ClearingHouse entry
	traceSharedNames []*sharedNamesEntry           // for producing diagnostics information about the source of each NameGroup
}

var byOSSName = make(map[ossrecord.CRNServiceName]*NameGroup)
var byCHID = make(map[clearinghouse.DeliverableID]*NameGroup)
var allNameGroups = make([]*NameGroup, 0, 500)

func newNameGroup() *NameGroup {
	ng := new(NameGroup)
	allNameGroups = append(allNameGroups, ng)
	return ng
}

func (ng *NameGroup) addOSSName(name ossrecord.CRNServiceName, traceSharedNames []*sharedNamesEntry) {
	if debug.IsDebugEnabled(debug.Fine) { // Avoid the overhead of "defer" if not debugging
		defer debug.Debug(debug.Fine, `addOSSName(%s) - data=%+v/%p   traceSharedNames=%v`, name, ng, ng, traceSharedNames)
	}

	// Record the sharedNames contributing to this NameGroup
	for _, sne1 := range traceSharedNames {
		var found bool
		for _, sne2 := range ng.traceSharedNames {
			if sne1 == sne2 {
				found = true
				break
			}
		}
		if !found {
			ng.traceSharedNames = append(ng.traceSharedNames, sne1)
		}
	}

	// Record the new name in the NameGroup
	if r0, found := byOSSName[name]; !found {
		byOSSName[name] = ng
		for _, n := range ng.OSSNames {
			if n == name {
				// already there; do nothing
				return
			}
		}
		ng.OSSNames = append(ng.OSSNames, name)
		sort.Slice(ng.OSSNames, func(i, j int) bool {
			return ng.OSSNames[i] < ng.OSSNames[j]
		})
	} else if r0 != ng {
		panic(fmt.Sprintf(`NameGroup.addOSSName found multiple NameGroups for name="%s":  r0=%+v/%p  r=%+v/%p`, name, r0, r0, ng, ng))
	}
}

func (ng *NameGroup) addCHID(chid clearinghouse.DeliverableID, traceSharedNames []*sharedNamesEntry) {
	if debug.IsDebugEnabled(debug.Fine) { // Avoid the overhead of "defer" if not debugging
		defer debug.Debug(debug.Fine, `addCHID(%s) - data=%+v/%p   traceSharedNames=%v`, chid, ng, ng, traceSharedNames)
	}

	// Record the sharedNames contributing to this NameGroup
	for _, sne1 := range traceSharedNames {
		var found bool
		for _, sne2 := range ng.traceSharedNames {
			if sne1 == sne2 {
				found = true
				break
			}
		}
		if !found {
			ng.traceSharedNames = append(ng.traceSharedNames, sne1)
		}
	}

	// Record the new CHID in the NameGroup
	if r0, found := byCHID[chid]; !found {
		byCHID[chid] = ng
		for _, id1 := range ng.CHIDs {
			if id1 == chid {
				// already there; do nothing
				return
			}
		}
		ng.CHIDs = append(ng.CHIDs, chid)
	} else if r0 != ng {
		panic(fmt.Sprintf(`NameGroup.addCHID found multiple NameGroups for id=%s:  r0=%+v/%p  r=%+v/%p`, chid, r0, r0, ng, ng))
	}
}

// dumpTraceSharedNames produces a multi-line string with all the traceSharedNames info from this NameGroup
func (ng *NameGroup) dumpTraceSharedNames(firstPrefix, mainPrefix string) string {
	buf := strings.Builder{}
	sort.Slice(ng.traceSharedNames, func(i, j int) bool {
		return ng.traceSharedNames[i].ComparableName < ng.traceSharedNames[j].ComparableName
	})
	for _, sne := range ng.traceSharedNames {
		buf.WriteString(sne.dump(firstPrefix+"    -", mainPrefix+"    "))
	}
	return buf.String()
}

// ListNameGroups calls the handler function with each NameGroup from the list,
// that brings together all ServiceInfo and all ClearingHouse records that some common names
func ListNameGroups(handler func(ng *NameGroup)) error {
	for _, ng := range allNameGroups {
		handler(ng)
	}
	return nil
}

// LookupNameGroupByOSSName returns the NameGroup that contains a given ServiceInfo record
// Note: one NameGroup may contain multiple ServiceInfo records, but each ServiceInfo is in only one NameGroup
func LookupNameGroupByOSSName(name ossrecord.CRNServiceName) (ng *NameGroup, found bool) {
	ng, found = byOSSName[name]
	return ng, found
}

// LookupNameGroupByCHID returns the NameGroup that contains a given ClearingHouse Deliverable ID
// Note: one NameGroup may contain multiple ClearingHouse IDs, but each ClearingHouse ID is in only one NameGroup
func LookupNameGroupByCHID(chid clearinghouse.DeliverableID) (ng *NameGroup, found bool) {
	ng, found = byCHID[chid]
	return ng, found
}
