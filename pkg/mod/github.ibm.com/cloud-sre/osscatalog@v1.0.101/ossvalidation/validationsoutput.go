package ossvalidation

import (
	"fmt"
	"sort"
	"strings"
)

// Header produces a short-multiline header summarizing the name variants and the number of ValidationIssues in this OSS ValidationInfo object
func (ossv *OSSValidation) Header() string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("-- Canonical Name:     %-20s   Found in: %v\n", fmt.Sprintf(`"%s"`, ossv.CanonicalName), ossv.CanonicalNameSources))
	otherNames := make([]string, 0, len(ossv.OtherNamesSources))
	for k := range ossv.OtherNamesSources {
		otherNames = append(otherNames, k)
	}
	sort.Strings(otherNames)
	for _, name := range otherNames {
		result.WriteString(fmt.Sprintf("-- Other Name:         %-20s   Found in: %v\n", fmt.Sprintf(`"%s"`, name), ossv.OtherNamesSources[name]))
	}
	if ossv.CatalogVisibility.EffectiveRestrictions != "" || ossv.CatalogVisibility.LocalRestrictions != "" {
		if ossv.CatalogVisibility.EffectiveRestrictions != ossv.CatalogVisibility.LocalRestrictions {
			result.WriteString(fmt.Sprintf("-- Catalog Visibility: effective=%s/local=%s  Active=%v  UI.Hidden=%v  Disabled=%v\n",
				ossv.CatalogVisibility.EffectiveRestrictions, ossv.CatalogVisibility.LocalRestrictions, ossv.CatalogVisibility.Active, ossv.CatalogVisibility.Hidden, ossv.CatalogVisibility.Disabled))

		} else {
			result.WriteString(fmt.Sprintf("-- Catalog Visibility: %-10s  Active=%v  UI.Hidden=%v  Disabled=%v\n",
				ossv.CatalogVisibility.EffectiveRestrictions, ossv.CatalogVisibility.Active, ossv.CatalogVisibility.Hidden, ossv.CatalogVisibility.Disabled))
		}
	}
	return result.String()
}

// Details generates a string representation of all the issues within this OSS ValidationInfo object
// Assumes the Issues are already sorted
func (ossv *OSSValidation) Details() string {
	var result strings.Builder
	result.WriteString(ossv.Header())
	counts := ossv.CountIssues(nil)
	result.WriteString(fmt.Sprintf("-- Validation Issues:  TOTAL=%d", counts["TOTAL"]))
	for _, s := range AllSeverityList() {
		result.WriteString(fmt.Sprintf("  %s=%d", s, counts[s]))
	}
	result.WriteString("\n")
	result.WriteString(fmt.Sprintf("-- Last overall run timestamp: \"%s\"\n", ossv.LastRunTimestamp))
	if ossv.LastRunActions != nil {
		actionNames := make([]string, 0, len(ossv.LastRunActions))
		for n := range ossv.LastRunActions {
			actionNames = append(actionNames, n)
		}
		sort.Strings(actionNames)
		for _, n := range actionNames {
			result.WriteString(fmt.Sprintf("-- Last run of action \"%s\": %s\n", n, ossv.LastRunActions[n]))
		}
	}
	if ossv.StatusCategoryCount != 0 {
		result.WriteString(fmt.Sprintf("-- Number of entries with the same StatusPage CategoryID: %d \n", ossv.StatusCategoryCount))
	}
	for _, issue := range ossv.Issues {
		result.WriteString(issue.String())
	}
	return result.String()
}
