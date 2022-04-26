package ossmerge

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// Functions for converting service/component names to their canonical or comparable forms

// allDoNotMergeNames contains all the entry that should not be folded into the common "comparable name";
// constructed from the "DoNotMergeNames attributes of all OSS entries"
var allDoNotMergeNames = make(map[string]string)

func registerDoNotMergeNames(src *ossmergecontrol.OSSMergeControl, names ...string) {
	// XXX for now we assume that the "target" is always the same as the do-not-merge name itself
	for _, n := range names {
		debug.Info(`Registering "do-not-merge" name %q from OSSMergeControl(%s)`, n, src.CanonicalName)
		// TODO: check that the do-not-merge name belongs in the OSSMergeControl entry that declares it
		if prior, found := allDoNotMergeNames[n]; found {
			if prior != n {
				debug.PrintError(`registerDoNotMergeNames(): conflicting registrations for name "%s": target="%s"  prior="%s"`, n, n, prior)
			}
		} else {
			allDoNotMergeNames[n] = n
		}
	}
}

// isNameNotMergeable returns true if the name is a on special list of "do-not-merge" names,
// that should not be folded into the common "comparable" form for merging with other entries
// with similar names
func isNameNotMergeable(name string) bool {
	_, found := allDoNotMergeNames[name]
	return found
}

/*
def makeCanonicalName(name):
    if (name in exceptionsDoNotMerge):
        return name
    prefixMatch = canonicalPrefixPattern.match(name)
    if prefixMatch:
        prefix = prefixMatch.group('prefix')
        rest = prefixMatch.group('rest')
        return prefix + canonicalPattern.sub('-', rest.strip().lower())
    else:
        return canonicalPattern.sub('-', name.strip().lower())
*/

var canonicalNamePattern = regexp.MustCompile(`[^a-z0-9-]`)

//var canonicalPrefixPattern = regexp.MustCompile(`^(?P<prefix>(is\.)|(oss\.))(?P<rest>.*)`)

var compositeNamePattern = regexp.MustCompile(`^([a-z0-9]+)\.([a-z0-9-]+)$`)

// MakeCanonicalName converts a raw service name into its canonical form according to the CRN spec
func MakeCanonicalName(name string) ossrecord.CRNServiceName {
	if n, found := allDoNotMergeNames[name]; found {
		return ossrecord.CRNServiceName(n)
	}
	return internalComputeCanonicalName(name)
}

// internalComputeCanonicalName is an internal function to compute a canonical name, ignoring
// the possibility of do-not-merge names
// (also used in the implementation of IsNameCanonical())
func internalComputeCanonicalName(name string) ossrecord.CRNServiceName {
	var result string
	trimmed := strings.ToLower(strings.TrimSpace(name))
	/*
		prefixMatch := canonicalPrefixPattern.FindStringSubmatch(trimmed)
		if prefixMatch != nil {
			prefix := strings.TrimSpace(prefixMatch[1])
			rest := strings.TrimSpace(prefixMatch[4])
			result = prefix + canonicalNamePattern.ReplaceAllLiteralString(rest, "-")
		} else {
			result = canonicalNamePattern.ReplaceAllLiteralString(trimmed, "-")
		}
	*/
	result = canonicalNamePattern.ReplaceAllLiteralString(trimmed, "-")
	return ossrecord.CRNServiceName(result)
}

// IsNameCanonical returns true if the name is in canonical name format
func IsNameCanonical(name string) bool {
	cn := internalComputeCanonicalName(name)
	return string(cn) == name
}

/*
comparisonPattern = re.compile('(ibm-?cloud)|(^ibm-?bluemix)|(^ibm)|(^bluemix)|([-.])') # remove common prefixes and '-' and '.' for comparisons
def makeComparableName(name):
    """Construct the "comparable name" for a service entry, i.e. a simplified name in which all common variations have been suppressed"""
    if (name in exceptionsDoNotMerge):
        return name
    else:
        canonicalName = makeCanonicalName(name)
		return comparisonPattern.sub('', canonicalName)
*/

var comparableNamePattern = regexp.MustCompile(`[^a-z0-9]`)
var collapseDashesPattern = regexp.MustCompile(`-{2,}`)

type prefixSpec struct {
	Original    string
	Replacement string
}

var comparablePrefixes = []prefixSpec{
	// NOTE: the order of these entries is important, as we will only process the first match
	{"argonauts-ibm-cloud-", ""}, // must come before just "argonauts"
	{"argonauts-ibm-", ""},       // must come before just "argonauts"
	{"argonauts-", ""},
	//	{"cloud", ""},
	{"ibm-cloud-", ""}, // must come before just "ibm"
	{"ibmcloud-", ""},  // must come before just "ibm"
	{"ibm-bluemix-", ""},
	{"ibm-", ""},
	{"bluemix-", ""},
	{"testxyz-", "xxyyzz"}, // for testing only
}

var comparableSuffixes = []prefixSpec{
	// NOTE: the order of these entries is important, as we will only process the first match
	{"-for-ibm-cloud", ""},
	{"-for-cloud", ""},
	{"-for-the-cloud", ""},
	{"-for-the-ibm-cloud", ""},
	{"-in-ibm-cloud", ""},
	{"-in-the-cloud", ""},
	{"-in-the-ibm-cloud", ""},
	{"-for-bluemix", ""},
	{"-for-ibm-bluemix", ""},
	{"-in-bluemix", ""},
	{"-in-ibm-bluemix", ""},
	{"-paygo", ""},
	{"-sqo", ""},
	{"billing-subscription", "billing-subscription"},
	{"-subscription", ""},
	{"-continuous-delivery", ""},
	//	{"-for-bluemix-continuous-delivery", ""},
	{"-for-bluemix-continuos-delivery", ""}, // Note typo
	{"-testzyx", "zzyyxx"},                  // for testing only
	// TODO: handle Global Catalog group entries "-group"
	// TODO: do something about "platform" ?
	// TODO: do something about "SaaS" ?
}

// MakeComparableName converts a raw service name into a simplified name that can be compared
// with other services that have similar names
func MakeComparableName(name string) string {
	if n, found := allDoNotMergeNames[name]; found {
		return n
	}
	trimmed := strings.ToLower(strings.TrimSpace(name))
	trimmed = canonicalNamePattern.ReplaceAllLiteralString(trimmed, "-")  // Replace all non standard chars with "-"
	trimmed = collapseDashesPattern.ReplaceAllLiteralString(trimmed, "-") // Replace sequence of "-" with a single "-"
	for _, p := range comparablePrefixes {
		if strings.HasPrefix(trimmed, p.Original) {
			trimmed = p.Replacement + strings.TrimPrefix(trimmed, p.Original)
			break
		}
	}
	for _, p := range comparableSuffixes {
		if strings.HasSuffix(trimmed, p.Original) {
			trimmed = strings.TrimSuffix(trimmed, p.Original) + p.Replacement
			break
		}
	}
	result := comparableNamePattern.ReplaceAllLiteralString(trimmed, "") // Remove the remaining "-"
	return result
}

// ParseCompositeName parses the name of a Catalog composite child entry (e.g. "is.volume") and extracts its parts
func ParseCompositeName(name string) (base, suffix string) {
	m := compositeNamePattern.FindStringSubmatch(name)
	if m != nil {
		return m[1], m[2]
	}
	return "", ""
}

// ConvertCompositeToCanonicalName converts the name of a Catalog entry that might be a child of a composite (e.g. "is.xxx")
// and returns a canonical name (e.g. "is-xxx").
// It also returns a bool to indicate if the original name was in fact a composite child name or not
func ConvertCompositeToCanonicalName(name string) (canonicalName ossrecord.CRNServiceName, isComposite bool) {
	base, suffix := ParseCompositeName(name)
	if base != "" {
		return ossrecord.CRNServiceName(base + "-" + suffix), true
	}
	return MakeCanonicalName(name), false
}

// CompareCompositeAndCanonicalName compares the name of a Catalog entry that might be a child of a composite (e.g. "is.xxx")
// with a corresponding canonical name (e.g. "is-xxx") and returns true if they match.
func CompareCompositeAndCanonicalName(name string, canonicalName ossrecord.CRNServiceName) bool {
	if name == string(canonicalName) {
		return true
	}
	base, suffix := ParseCompositeName(name)
	if base != "" {
		if base+"-"+suffix == string(canonicalName) {
			return true
		}
	}
	return false
}

var invalidCRNNamePattern = regexp.MustCompile(`(ibm-?cloud)|(\bibm\b)|(bluemix)|(ibm-?bluemix)|(-prod$)|(-dev$)|(-test$)`)

// CheckValidCRNServiceName checks that a given CRN service-name meets all the naming requirements for CRN service-names
func checkValidCRNServiceName(name ossrecord.CRNServiceName) error {
	m := invalidCRNNamePattern.FindString(strings.ToLower(string(name)))
	if m != "" {
		return fmt.Errorf(`Name contains "%s"   (invalid names pattern="%s")`, m, invalidCRNNamePattern.String())
	}
	return nil
}
