package cataloghidden

import (
	"html/template"
	"io"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// CatalogHidden generates a report for all known services/components that are currently hidden (not publicly visible) in the Global Catalog and writes it to the specified Writer
// TODO: Revise/fix use the -visibility option
// TODO: print information to differentiate between effective Visibility (through parents) and local Visibility (each record for itself)
func CatalogHidden(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	data := struct {
		TimeStamp string
		Pattern   string
		Records   []*ossmerge.ServiceInfo
	}{
		TimeStamp: timeStamp,
		Pattern:   pattern.String(),
	}

	var handler = func(si *ossmerge.ServiceInfo) {
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if si.HasSourceMainCatalog() {
			//FIXME:		if si.GetSourceMainCatalog().EffectiveVisibility.Restrictions != string(catalog.VisibilityPublic) || !si.GetSourceMainCatalog().Active {
			if true { // XXX
				data.Records = append(data.Records, si)
			}
		}
	}

	var err error
	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}

	templateSrc := catalogHiddenTemplateText

	t, err := template.New("CatalogHidden").Parse(templateSrc)
	if err != nil {
		return err
	}
	err = t.Execute(w, &data)
	return err
}

var catalogHiddenTemplateText = `=========== CatalogHidden: Services/Components that are currently hidden in the catalog ==========================
  Generated: {{.TimeStamp}}        Pattern: {{.Pattern}}
{{range .Records}}
*** {{.OSSService.ReferenceResourceName}}  "{{.OSSService.ReferenceDisplayName}}"  Type={{.OSSService.GeneralInfo.EntryType}}  Status={{.OSSService.GeneralInfo.OperationalStatus}}
	OSS Tags:                 {{.OSSService.GeneralInfo.OSSTags}}
	Tags:                     {{.SourceMainCatalog.Tags}}
	Canonical Name:           {{printf "%-25s" .OSSValidation.CanonicalName}}  Found in: {{.OSSValidation.CanonicalNameSources}}
	{{- range $index, $element := .OSSValidation.OtherNamesSources}} 
	Other Name Variant:       {{printf "%-25s" $index}}  Found in: {{$element}}
	{{- end}}
	Active:                   {{.SourceMainCatalog.Active}}
	Disabled:                 {{.SourceMainCatalog.Disabled}}
	Visibility Restrictions:  {{.SourceMainCatalog.EffectiveVisibility.Restrictions}}
	Visibility Full Record:   {{.SourceMainCatalog.EffectiveVisibility}}

{{end}}
----------- End of CatalogHidden Report ---------------------------------------

`
