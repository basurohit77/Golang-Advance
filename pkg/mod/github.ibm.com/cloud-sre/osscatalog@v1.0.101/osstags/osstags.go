package osstags

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// Package osstags is used to track various special characteristics of a OSS entry,
// that control validation rules that apply to this entry, reflect the overall status
// of the service/component from a OSS perpective, etc.

// Tag represents one Tag (possibly with expiration time info appended)
type Tag string

// TagMetaData represents all the metadata associated with one tag
type TagMetaData struct {
	tag        Tag
	group      TagGroup
	rmcManaged bool
	isStatus   bool
}

// HasExpiration returns true if this Tag has an expiration date (not necessarily expired)
func (t Tag) HasExpiration() bool {
	if ix := strings.IndexByte(string(t), '>'); ix >= 0 {
		return true
	}
	return false
}

// IsExpired returns true if this Tag has an expiration date and that date is past.
func (t Tag) IsExpired() bool {
	if ix := strings.IndexByte(string(t), '>'); ix >= 0 {
		str := string(t[ix+1:])
		expiration, err := time.Parse("060102", str)
		//		fmt.Printf("DEBUG: IsExpired(%s) -> expiration=%v   err=%v\n", t, expiration, err)
		if err == nil {
			if expiration.Before(time.Now()) {
				return true
			}
		} else {
			debug.PrintError(`Cannot parse OSS Tag with expiration date: "%s" -> %v`, str, err)
		}
	}
	return false
}

// BaseName returns the base tag (ignoring any expiration date)
func (t Tag) BaseName() Tag {
	if ix := strings.IndexByte(string(t), '>'); ix >= 0 {
		return t[:ix]
	}
	return t
}

// ComparableString forces comparisons of Tags using the special string slice comparator
func (t Tag) ComparableString() string {
	return string(t)
}

// All the possible OSS Tag values (many carried over from the service-validation python script)
const (
	// TODO: Provide detailed descriptions of each of the OSSStatus values

	OSSOnly            Tag = "oss_only_placeholder" // For items that exist only as OSS records, not in any other sources -- e.g. some OSSEnvironment entries or placeholders for future services
	OSSStaging         Tag = "oss_staging_only"     // For items that should exist only as OSS records in the Staging catalog but never in Production
	OSSDelete          Tag = "oss_delete"           // To force deletion of the OSS record at the next run of osscatimporter/osscatpublisher (for use in in backup/restore)
	OSSLock            Tag = "oss_lock"             // To lock the entry against publishing or updating in Production (but not affect already-published entry)
	CatalogNative      Tag = "catalog_native"       // For items that exist as "native" entries in the Global Catalog, with the OSS record built on the fly not explicitly stored -- e.g. some OSSEnvironment entries
	NotReady           Tag = "not_ready"            // from service-validation python script: NOTREADY
	SelectAvailability Tag = "select_availability"  // from service-validation python script: DARKLAUNCH or CF_OBSOLETE
	Deprecated         Tag = "deprecated"           // from service-validation python script: DEPRECATED
	Retired            Tag = "retired"              // service is known to be retired -- for info only
	Internal           Tag = "internal"             // To force OperationalStatus INTERNAL
	Invalid            Tag = "invalid"              // entry is known to be invalid (should not even be retired) -- for info only
	TypeOtherOSS       Tag = "type_otheross"        // Entry is of type "OtherOSS" i.e. not tracked with IBM Cloud proper
	TypeGaaS           Tag = "type_gaas"            // Entry is of type "GAAS" i.e. not tracked with IBM Cloud proper
	TypeComponent      Tag = "type_component"       // from service-validation python script: COMPONENT
	TypeSubcomponent   Tag = "type_subcomponent"    // from service-validation python script: SUBCOMPONENT
	TypeSupercomponent Tag = "type_supercomponent"  // Entry is of type "SUPERCOMPONENT"
	TypeVMware         Tag = "type_vmware"          // Services deployed and managed under VMware umbrella
	TypeIAMOnly        Tag = "type_iam_only"        // Entry is of type IAMONLY
	TypeContent        Tag = "type_content"         // Entry is of type "CONTENT"
	TypeConsulting     Tag = "type_consulting"      // Entry is of type "CONSULTING"
	TypeInternal       Tag = "type_internal"        // Entry is of type "INTERNALSERVICE"
	ClientFacing       Tag = "client_facing"        // from service-validation python script: CLIENTFACING
	NotClientFacing    Tag = "not_client_facing"    // from service-validation python script: ??
	//	CustomUI            Tag = "custom_ui"            // Service has a custom UI in the Catalog; ignore normal rules for displaying Catalog tiles
	NotCF              Tag = "not_in_cf"            // from service-validation python script: NOT_CF
	IaaSGen1           Tag = "iaas_gen1"            // from service-validation python script: IAAS_LEGACY
	IaaSGen1Other      Tag = "iaas_gen1_other"      // service-validation from python script: IAAS
	IaaSGen2           Tag = "iaas_gen2"            // service-validation from python script: IAAS_NEW
	OneCloud           Tag = "onecloud"             // All services currently badged as OneCloud
	OneCloudComponent  Tag = "onecloud_component"   // Components and subordinate entries associated with services currently badged as OneCloud (kept separate to maintain an accurate count of OneCloud badged services)
	OneCloudWave1      Tag = "onecloud_wave1"       // Wave 1 services targeted for OneCloud badging
	OneCloudWave2      Tag = "onecloud_wave2"       // Wave 2 services targeted for OneCloud badging
	OneCloudWave3      Tag = "onecloud_wave3"       // Wave 3 services targeted for OneCloud badging
	PalanteExclude     Tag = "palante_exclude"      // Force exclusion from Palante Heuristics Engine CIE generation
	PalanteInclude     Tag = "palante_include"      // Force inclusion in Palante Heuristics Engine CIE generation
	PalanteTest        Tag = "palante_test"         // Test entry for Palante Heuristics Engine - do not show on external status page
	PnPCandidate       Tag = "pnp_candidate"        // Candidate to be PnPEnabled - to validate if it actually can be
	PnPInclude         Tag = "pnp_include"          // Entry should be PnPEnabled even if it has outstanding issues
	PnPExclude         Tag = "pnp_exclude"          // Entry should not be PnPEnabled even if it meets all criteria
	PnPEnabled         Tag = "pnp_enabled"          // Entry is valid for use with PnP
	PnPEnabledIaaS     Tag = "pnp_enabled_iaas"     // Entry is valid for use with PnP and uses IaaS (SoftLayer) as a status source
	EDBInclude         Tag = "edb_include"          // Entry is required to have EDB Monitoring data, even it does not mean the normal conditions for EDB
	EDBExclude         Tag = "edb_exclude"          // Entry should not have any EDB Monitoring data
	OSSTest            Tag = "oss_test"             // To mark test records (never to be published to Production)
	LenientCRNName     Tag = "lenient_crn_name"     // Allow invalid words in CRN service name (e.g. "bluemix", "cloud", etc.)
	LenientDisplayName Tag = "lenient_display_name" // Allow invalid words in CRN Display name (e.g. "Bluemix", "Cloud", etc.)
	FSCloud            Tag = "fscloud"              // To track Environments that are part of the FS Cloud compliance
	USRegulated        Tag = "us_regulated"         // To track Environments that may be part of the US Regulated boundary
	FedRAMP            Tag = "fedramp"              // To track compliance regime for services/components (should mirror the Catalog tag if there is a Main Catalog entry)

	IBMCloudDefaultSegment Tag = "ibmcloud_default_segment" // To mark the default segment owning all IBM Public Cloud and Dedicated/Local environments

	ServiceNowApproved Tag = "servicenow_approved" // Allow creation of a ServiceNow entry

	StatusRed       Tag = "oss_status_red"        // Overall validation status: Red
	StatusYellow    Tag = "oss_status_yellow"     // Overall validation status: Yellow
	StatusGreen     Tag = "oss_status_green"      // Overall validation status: Green
	StatusCRNRed    Tag = "oss_status_crn_red"    // Main/CRN validation status: Red
	StatusCRNYellow Tag = "oss_status_crn_yellow" // Main/CRN validation status: Yellow
	StatusCRNGreen  Tag = "oss_status_crn_green"  // Main/CRN validation status: Green

	// Old types from the python script
	// TODO: Check which old OSSStatus values from service-validation python script are still needed
	//OneCloud1
	//OneCloud2
	//Deprecated -> moved to OperationalStatus
	//Component -> moved to EntryType?
	//Subcomponent
	//SubcomponentClientFacing
	//NotAService
	//Runtime
	//Reviewed
	//ClientFacing
	//RCNotCatalog
	//CFObsolete
	//ScorecardV1Deprecated -> now tracked as OperationalStatus
	//Placeholder
)

// TagGroup idenfies a group of related tags. Only one type from each TagGroup is allowed in each TagSet
type TagGroup string

// All possible TagGroup values
const (
	GroupStorageControl    TagGroup = "storage-control"
	GroupOperationalStatus TagGroup = "status"
	GroupEntryType         TagGroup = "type"
	GroupClientFacing      TagGroup = "client-facing"
	GroupIaaS              TagGroup = "iaas"
	GroupOneCloud          TagGroup = "onecloud"
	GroupPnpControl        TagGroup = "pnpcontrol"
	GroupEDBControl        TagGroup = "edbcontrol"
	GroupOSSStatus         TagGroup = "oss-status"
	GroupCRNStatus         TagGroup = "crn-status"
	GroupPnPStatus         TagGroup = "pnp-status"
	GroupPalanteControl    TagGroup = "palante-control"
)

// allValidTags specifies all the possible tags along with their metadata that they belong to
var allValidTags = make(map[Tag]*TagMetaData)

func registerTag(t Tag, g TagGroup, rmcManaged bool) {
	if _, found := allValidTags[t]; found {
		panic(fmt.Sprintf("Found duplicate osstags.Tag: %q", t))
	}
	var isStatus bool
	switch g {
	case GroupOSSStatus, GroupCRNStatus, GroupPnPStatus:
		isStatus = true
	default:
		isStatus = false
	}
	allValidTags[t] = &TagMetaData{tag: t, group: g, rmcManaged: rmcManaged, isStatus: isStatus}
}

func init() {
	registerTag(OSSOnly, GroupStorageControl, false)
	registerTag(OSSStaging, GroupStorageControl, false)
	registerTag(OSSDelete, GroupStorageControl, false)
	registerTag(OSSLock, GroupStorageControl, false)
	registerTag(OSSTest, "", false)
	registerTag(CatalogNative, GroupStorageControl, false)

	registerTag(NotReady, GroupOperationalStatus, true)
	registerTag(SelectAvailability, GroupOperationalStatus, true)
	registerTag(Deprecated, GroupOperationalStatus, true)
	registerTag(Retired, GroupOperationalStatus, true)
	registerTag(Internal, GroupOperationalStatus, true)
	registerTag(Invalid, GroupOperationalStatus, true)

	registerTag(TypeOtherOSS, GroupEntryType, true)
	registerTag(TypeGaaS, GroupEntryType, true)
	registerTag(TypeComponent, GroupEntryType, true)
	registerTag(TypeSubcomponent, GroupEntryType, true)
	registerTag(TypeSupercomponent, GroupEntryType, true)
	registerTag(TypeVMware, GroupEntryType, true)
	registerTag(TypeIAMOnly, GroupEntryType, true)
	registerTag(TypeContent, GroupEntryType, true)
	registerTag(TypeConsulting, GroupEntryType, false) // TODO: TypeConsulting is not yet RMC managed - update when RMC does support the CONSULTING type
	registerTag(TypeInternal, GroupEntryType, true)

	registerTag(NotCF, "", false)

	registerTag(ClientFacing, GroupClientFacing, false)
	registerTag(NotClientFacing, GroupClientFacing, false)
	//registerTag(CustomUI, "", false)

	registerTag(IaaSGen1, GroupIaaS, false)
	registerTag(IaaSGen1Other, GroupIaaS, false)
	registerTag(IaaSGen2, GroupIaaS, false)

	registerTag(OneCloud, GroupOneCloud, false)
	registerTag(OneCloudComponent, GroupOneCloud, false)
	registerTag(OneCloudWave1, GroupOneCloud, false)
	registerTag(OneCloudWave2, GroupOneCloud, false)
	registerTag(OneCloudWave3, GroupOneCloud, false)

	registerTag(PnPCandidate, GroupPnpControl, false)
	registerTag(PnPInclude, GroupPnpControl, true)
	registerTag(PnPExclude, GroupPnpControl, true)
	registerTag(PnPEnabled, GroupPnPStatus, false)
	registerTag(PnPEnabledIaaS, "", false)

	registerTag(EDBInclude, GroupEDBControl, true)
	registerTag(EDBExclude, GroupEDBControl, true)

	registerTag(LenientCRNName, "", false)
	registerTag(LenientDisplayName, "", false)
	registerTag(IBMCloudDefaultSegment, "", false)
	registerTag(ServiceNowApproved, "", true)

	registerTag(PalanteExclude, GroupPalanteControl, false)
	registerTag(PalanteInclude, GroupPalanteControl, false)
	registerTag(PalanteTest, "", false)

	registerTag(FSCloud, "", false)
	registerTag(USRegulated, "", false)
	registerTag(FedRAMP, "", false)

	registerTag(StatusRed, GroupOSSStatus, false)
	registerTag(StatusYellow, GroupOSSStatus, false)
	registerTag(StatusGreen, GroupOSSStatus, false)

	registerTag(StatusCRNRed, GroupCRNStatus, false)
	registerTag(StatusCRNYellow, GroupCRNStatus, false)
	registerTag(StatusCRNGreen, GroupCRNStatus, false)
}

// the list of Tags that correspond to the OSS validation status
// TODO: should we have a special type for the OSS validation status Tags
var statusOverallTags = map[Tag]string{StatusRed: "Overall:Red", StatusYellow: "Overall:Yellow", StatusGreen: "Overall:Green"}
var statusCRNTags = map[Tag]string{StatusCRNRed: "CRN:Red", StatusCRNYellow: "CRN:Yellow", StatusCRNGreen: "CRN:Green"}
var allStatusTags map[Tag]string
var allStatusTagsShort = map[Tag]string{StatusRed: "R", StatusYellow: "Y", StatusGreen: "G", StatusCRNRed: "R", StatusCRNYellow: "Y", StatusCRNGreen: "G"}

// TagSet represents the set of Tags associated with one OSS entry
type TagSet []Tag

// Contains returns true if the TagSet contains the specified Tag
func (ts *TagSet) Contains(t Tag) bool {
	if t == "" {
		return false
	}
	if t.HasExpiration() {
		panic(fmt.Sprintf("osstags.TagsSet.Contains(%s) called with Tag that has an expiration date", t))
	}
	// TODO: optimize TagSet search with a cached map?
	for _, e := range *ts {
		//		fmt.Printf("DEBUG: e=%v  HasBaseName(%s)=%v  IsExpired()=%v\n", e, t, e.HasBaseName(t), e.IsExpired())
		if e.BaseName() == t && !e.IsExpired() {
			return true
		}
	}
	return false
}

// AddTag adds the specified Tag in the TagSet
func (ts *TagSet) AddTag(t Tag) {
	if t == "" {
		panic("osstags.TagsSet.AddTag() - attempt to set empty tag")
	}
	if t.HasExpiration() {
		panic(fmt.Sprintf("osstags.TagsSet.AddTag(%s) called with Tag that has an expiration date", t))
	}
	if ts.Contains(t) {
		// the Tag is already in the TagSet - no-op
		return
	}
	*ts = append(*ts, t)
	return
}

// RemoveTag removes the specified Tag from the TagSet, if present
func (ts *TagSet) RemoveTag(t Tag) {
	if t.HasExpiration() {
		panic(fmt.Sprintf("osstags.TagsSet.RemoveTag(%s) called with Tag that has an expiration date", t))
	}
	ts1 := *ts
restart:
	for i, e := range ts1 {
		if e.BaseName() == t {
			ts1 = append(ts1[:i], ts1[i+1:]...)
			goto restart
		}
	}
	*ts = ts1
	return
}

// SetOverallStatus adds or replaces one Tag that represents the OSS overall validation status in the TagSet
func (ts *TagSet) SetOverallStatus(t Tag) {
	if t.HasExpiration() {
		panic(fmt.Sprintf("osstags.TagsSet.SetOverallStatus(%s) called with Tag that has an expiration date", t))
	}
	// TODO: Should ensure that the OSS validation status Tag is always first in the TagSet
	if _, found := statusOverallTags[t]; !found {
		panic(fmt.Sprintf("osstags.TagSet.SetOverallStatus() called with a non-status tag: %v", t))
	}
	ts1 := *ts
	for i, e := range ts1 {
		if _, found := statusOverallTags[e]; found {
			ts1 = append(ts1[:i], ts1[i+1:]...)
		}
	}
	ts1 = append([]Tag{t}, ts1...)
	*ts = ts1
}

// GetOverallStatus returns the one Tag that represents the OSS overall validation status in the TagSet,
// or the zero value if not status tag is present
func (ts *TagSet) GetOverallStatus() Tag {
	ts1 := *ts
	for _, e := range ts1 {
		if _, found := statusOverallTags[e]; found {
			return e
		}
	}
	return ""
}

// SetCRNStatus adds or replaces one Tag that represents the OSS main/CRN validation status in the TagSet
func (ts *TagSet) SetCRNStatus(t Tag) {
	if t.HasExpiration() {
		panic(fmt.Sprintf("osstags.TagsSet.SetCRNStatus(%s) called with Tag that has an expiration date", t))
	}
	// TODO: Should ensure that the OSS validation status Tag is always first in the TagSet
	if _, found := statusCRNTags[t]; !found {
		panic(fmt.Sprintf("osstags.TagSet.SetCRNStatus() called with a non-status tag: %v", t))
	}
	ts1 := *ts
	for i, e := range ts1 {
		if _, found := statusCRNTags[e]; found {
			ts1 = append(ts1[:i], ts1[i+1:]...)
		}
	}
	ts1 = append([]Tag{t}, ts1...)
	*ts = ts1
}

// GetCRNStatus returns the one Tag that represents the OSS main/CRN validation status in the TagSet,
// or the zero value if not status tag is present
func (ts *TagSet) GetCRNStatus() Tag {
	ts1 := *ts
	for _, e := range ts1 {
		if _, found := statusCRNTags[e]; found {
			return e
		}
	}
	return ""
}

// StringStatus returns a string representation of this Tag
// This method provides a short string for the Status-related tags;
// other tags are rendered as-is
func (t Tag) StringStatus() string {
	if s, found := allStatusTags[t]; found {
		return s
	}
	return string(t)
}

// StringStatusShort returns a very short string representation of this Tag
// This method provides a special short string for the Status-related tags;
// other tags are rendered as-is
func (t Tag) StringStatusShort() string {
	if s, found := allStatusTagsShort[t]; found {
		return s
	}
	return string(t)
}

// IsRMCManaged returns true if this Tag should be ignored when there is a OSS record managed by RMC
func (t Tag) IsRMCManaged() bool {
	if tmd, found := allValidTags[t]; found {
		return tmd.rmcManaged
	}
	baseTag, _, err := parseTagName(string(t), true)
	if err != nil {
		return false
	}
	if tmd, found := allValidTags[baseTag]; found {
		return tmd.rmcManaged
	}
	return false
}

// IsStatusTag returns true if this Tag is one of the status tags (computed by ossmerge, not used as input)
func (t Tag) IsStatusTag() bool {
	if tmd, found := allValidTags[t]; found {
		return tmd.isStatus
	}
	baseTag, _, err := parseTagName(string(t), true)
	if err != nil {
		return false
	}
	if tmd, found := allValidTags[baseTag]; found {
		return tmd.isStatus
	}
	return false
}

// String returns a string representation of this TagSet
func (ts *TagSet) String() string {
	return fmt.Sprintf("%v", *ts)
}

// Copy copies the given TagSet into a new TagSet
func (ts *TagSet) Copy() TagSet {
	ret := make(TagSet, len(*ts))
	copy(ret, *ts)
	return ret
}

// WithoutStatus returns this TagSet, all except the OSS Status tags
func (ts *TagSet) WithoutStatus() *TagSet {
	ts1 := make([]string, 0, len(*ts))
	for _, e := range *ts {
		if tmd, found := allValidTags[e]; found {
			if !tmd.isStatus {
				ts1 = append(ts1, string(e))
			}
		} else {
			ts1 = append(ts1, string(e))
		}
	}

	sort.Strings(ts1)

	ts2 := TagSet{}
	for _, e := range ts1 {
		ts2 = append(ts2, Tag(e))
	}

	return &ts2
}

// WithoutPureStatus returns this TagSet, all except the OSS Pure Status tags(GroupOSSStatus and GroupCRNStatus)
func (ts *TagSet) WithoutPureStatus() *TagSet {
	ts1 := make([]string, 0, len(*ts))
	for _, e := range *ts {
		if tmd, found := allValidTags[e]; found {
			if !tmd.isStatus {
				ts1 = append(ts1, string(e))
			} else if tmd.group != GroupOSSStatus && tmd.group != GroupCRNStatus {
				ts1 = append(ts1, string(e))
			}
		} else {
			ts1 = append(ts1, string(e))
		}
	}

	sort.Strings(ts1)

	ts2 := TagSet{}
	for _, e := range ts1 {
		ts2 = append(ts2, Tag(e))
	}

	return &ts2
}

// Validate ensures that the TagSet does not contain any invalid or conflicting tags.
// In addition, it also transforms each tag into its canonical name and sorts the tags in the TagSet
func (ts *TagSet) Validate(allowStatusTags bool) error {
	output := strings.Builder{}
	tagGroupFound := make(map[TagGroup]Tag)
	ts1 := make(TagSet, 0, len(*ts))

	for _, t := range *ts {
		// Check for invalid tags
		baseTag, fullTag, err := parseTagName(string(t), allowStatusTags)
		if err != nil {
			output.WriteString(fmt.Sprintf(`  Invalid tag "%s": %v\n`, t, err))
			continue
		}
		// TODO: Check for conflicts
		if tmd, found := allValidTags[baseTag]; found {
			if tmd.group != "" {
				if ot, found := tagGroupFound[tmd.group]; found {
					output.WriteString(fmt.Sprintf(`  Tag "%s" cannot be used in conjunction with Tag "%s"\n`, fullTag, ot))
				} else {
					tagGroupFound[tmd.group] = fullTag
				}
			}
		} else {
			panic(fmt.Sprintf("Unregistered osstags.Tag: %q", baseTag))
		}
		ts1 = append(ts1, fullTag)
	}

	// Ensure that the OSSTags are sorted
	sort.SliceStable(ts1, func(i, j int) bool {
		return ts1[i] < ts1[j]
	})

	if output.Len() != 0 {
		str := output.String()
		return fmt.Errorf("Invalid TagSet %v: \n%s", *ts, str)
	}
	*ts = ts1
	return nil
}

var foldNamePattern = regexp.MustCompile(`[._:-]`)

// foldTagName folds a tag name into a plain string that can be used to compare variations of the same tag name
func foldTagName(name string) string {
	trimmed := strings.ToLower(strings.TrimSpace(name))
	result := foldNamePattern.ReplaceAllLiteralString(trimmed, "")
	return result
}

var tagNameMap = make(map[string]Tag)

// Initialize the tagNameMap used to validate tag names and the combined list of all status tags
func init() {
	allStatusTags = make(map[Tag]string)
	for t, s := range statusOverallTags {
		allStatusTags[t] = s
	}
	for t, s := range statusCRNTags {
		allStatusTags[t] = s
	}
	for _, tmd := range allValidTags {
		if tmd.isStatus {
			if _, found := allStatusTags[tmd.tag]; !found {
				allStatusTags[tmd.tag] = string(tmd.tag)
			}
		}
		n := foldTagName(string(tmd.tag))
		if t0, found := tagNameMap[n]; found {
			panic(fmt.Sprintf("Found two OSSTags with similar names: %v / %v", t0, tmd.tag))
		}
		tagNameMap[n] = tmd.tag
	}

	tagNameMap[foldTagName("limited_availability")] = SelectAvailability
	tagNameMap[foldTagName("Red")] = StatusRed
	tagNameMap[foldTagName("StatusRed")] = StatusRed
	tagNameMap[foldTagName("OverallRed")] = StatusRed
	tagNameMap[foldTagName("Yellow")] = StatusYellow
	tagNameMap[foldTagName("StatusYellow")] = StatusYellow
	tagNameMap[foldTagName("OverallYellow")] = StatusYellow
	tagNameMap[foldTagName("Green")] = StatusGreen
	tagNameMap[foldTagName("StatusGreen")] = StatusGreen
	tagNameMap[foldTagName("OverallGreen")] = StatusGreen
	tagNameMap[foldTagName("CRNRed")] = StatusCRNRed
	tagNameMap[foldTagName("CRNYellow")] = StatusCRNYellow
	tagNameMap[foldTagName("CRNGreen")] = StatusCRNGreen
}

// parseTagName parses a tag name (allowing for variations in spelling + expiration date)
// it returns the base Tag name (as a string), the full Tag (including expiration date if any)
// or an error if the name cannot be parsed to a valid tag
func parseTagName(name string, allowStatusTags bool) (baseTag Tag, fullTag Tag, err error) {
	var folded string
	var expString string
	if ix := strings.IndexByte(string(name), '>'); ix >= 0 {
		folded = foldTagName(name[:ix])
		expString = string(name[ix+1:])
		if expString == "" {
			return "", "", fmt.Errorf(`Malformed expiration date string: "%s"`, name)
		}
	} else {
		folded = foldTagName(name)
		expString = ""
	}
	baseTag, ok := tagNameMap[folded]
	if !ok {
		return "", "", fmt.Errorf(`Invalid tag name: "%s"`, name)
	}
	if tmd, found := allValidTags[baseTag]; found {
		if tmd.isStatus {
			if !allowStatusTags {
				return "", "", fmt.Errorf(`Disallowed status tag: "%s"`, name)
			} else if expString != "" {
				return "", "", fmt.Errorf(`Expiration date not allowed on status tag: "%s"`, name)
			}
		}
	} else {
		panic(fmt.Sprintf("Tag %q found in tagNameMap but not in allValidTags", name))
	}
	if expString != "" {
		_, err := time.Parse("060102", expString)
		if err != nil {
			return "", "", debug.WrapError(err, `Invalid expiration date format in tag: "%s"`, name)
		}
		return baseTag, Tag(fmt.Sprintf("%s>%s", baseTag, expString)), nil
	}
	return baseTag, baseTag, nil
}

// GetExpiredTags finds all the expired tags in the given TagSet and returns
// a new TagSet containing only those expired tags
func (ts *TagSet) GetExpiredTags() TagSet {
	var result TagSet
	for _, t := range *ts {
		if t.IsExpired() {
			result = append(result, t)
		}
	}
	return result
}

// GetTagByGroup extract one Tag from the TagSet that belongs to the given TagGroup
// or returns the empty string if no matching Tag found
func (ts *TagSet) GetTagByGroup(g TagGroup) Tag {
	var result []Tag
	for _, t := range *ts {
		baseTag, _, err := parseTagName(string(t), true)
		if err != nil {
			panic(fmt.Sprintf(`GetTagByGroup(%s): invalid tag: %v`, g, err))
		}
		if tmd, found := allValidTags[baseTag]; found {
			if g == tmd.group {
				result = append(result, t)
			}
		}
	}
	switch len(result) {
	case 0:
		return ""
	case 1:
		return result[0]
	default:
		panic(fmt.Sprintf(`GetTagByGroup(%s): found multiple matching tags: %v`, g, result))
	}
}

/* XXX Disabled the Comparable interface: We want to use the slice comparator to see exact differences
// ComparableString forces comparisons of TagSets to be on a single line with all tags at once, rather than one slice element at a time
func (ts TagSet) ComparableString() string {
	return ts.String()
}
*/
