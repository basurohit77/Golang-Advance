package catalogvisibilitygroups

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// CatalogVisibilityGroups generates a report for all known services/components that are currently hidden (not publicly visible) in the Global Catalog
// grouped by different types of visibility restrictions (based on the actual Visibility.Restrictions field but also based on various flags
// e.g Active, Disabled, Hidden
// TODO: print information to differentiate between effective Visibility (through parents) and local Visibility (each record for itself)
func CatalogVisibilityGroups(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	data := struct {
		TimeStamp string
		Pattern   string
		Groups    map[string][]*ossmerge.ServiceInfo
	}{
		TimeStamp: timeStamp,
		Pattern:   pattern.String(),
		Groups:    make(map[string][]*ossmerge.ServiceInfo),
	}

	// Build the list of groups
	var handler = func(si *ossmerge.ServiceInfo) {
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if si.OSSValidation.CatalogVisibility.EffectiveRestrictions != "" {
			var kind string
			var ignored string
			if si.IgnoredMainCatalog != nil {
				kind = si.IgnoredMainCatalog.Kind
				ignored = "   (ignored for OSS)"
			} else {
				kind = si.SourceMainCatalog.Kind
				ignored = ""
			}
			groupName := fmt.Sprintf("%+v  kind=%s%s", si.OSSValidation.CatalogVisibility, kind, ignored)
			data.Groups[groupName] = append(data.Groups[groupName], si)
		}
	}
	var err error
	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}

	// Sort all the buckets
	for _, m := range data.Groups {
		sort.Slice(m, func(i, j int) bool {
			name1 := m[i].SourceMainCatalog.Name
			if m[i].IgnoredMainCatalog != nil {
				name1 = m[i].IgnoredMainCatalog.Name
			}
			name2 := m[j].SourceMainCatalog.Name
			if m[j].IgnoredMainCatalog != nil {
				name2 = m[j].IgnoredMainCatalog.Name
			}
			return name1 < name2
		})
	}

	// Initialize the template
	templateSrc := catalogVisibilityGroupsTemplateText

	funcMap := make(map[string]interface{})
	funcMap["getName"] = getName
	funcMap["getTags"] = getTags
	t, err := template.New("CatalogVisibilityGroups").Funcs(funcMap).Parse(templateSrc)
	if err != nil {
		return err
	}

	// Generate output
	err = t.Execute(w, &data)
	return err
}

func getName(si *ossmerge.ServiceInfo) string {
	if si.IgnoredMainCatalog != nil {
		return si.IgnoredMainCatalog.Name
	}
	return si.SourceMainCatalog.Name
}

func getTags(si *ossmerge.ServiceInfo) string {
	var tags []string
	if si.IgnoredMainCatalog != nil {
		tags = si.IgnoredMainCatalog.Tags
	} else {
		tags = si.SourceMainCatalog.Tags
	}

	result := strings.Builder{}
	for _, t := range tags {
		switch t {
		case "ibm_third_party", "ibm_deprecated", "deprecated", "community":
			result.WriteString(" ")
			result.WriteString(t)
		}
	}
	return result.String()
}

var catalogVisibilityGroupsTemplateText = `=========== CatalogVisibilityGroups: Services/Components that are not fully public in the catalog, grouped by visibility class ==========================
  Generated: {{.TimeStamp}}        Pattern: {{.Pattern}}

Summary:
{{range $groupName, $groupMembers := .Groups -}}
* {{printf "%3d" (len $groupMembers)}} entries: Group {{printf "%-90s" $groupName}}
{{end}}


Lists of Entries in each Group:
{{range $groupName, $groupMembers := .Groups}}
* Group {{printf "%-90s" $groupName}} : {{len $groupMembers}} entries
{{- range $groupMembers}}
		{{printf "%-30s" (getName .) }}    {{getTags . -}}
{{end}}
{{end}}

----------- End of CatalogVisibilityGroups Report ---------------------------------------

`
