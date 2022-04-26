package catalogga

import (
	"html/template"
	"io"
	"regexp"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
)

// CatalogGA generates a report for all services/components in the Catalog that are currently GA and Active (client-facing)
func CatalogGA(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	data := struct {
		TimeStamp string
		Pattern   string
		Records   map[ossrecord.EntryType][]*ossmerge.ServiceInfo
	}{
		TimeStamp: timeStamp,
		Pattern:   pattern.String(),
		Records:   make(map[ossrecord.EntryType][]*ossmerge.ServiceInfo),
	}

	var handler = func(si *ossmerge.ServiceInfo) {
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if si.HasSourceMainCatalog() {
			if si.GetSourceMainCatalog().Active && si.OSSService.GeneralInfo.OperationalStatus == ossrecord.GA {
				entryType := si.OSSService.GeneralInfo.EntryType
				bucket, ok := data.Records[entryType]
				if !ok {
					bucket = make([]*ossmerge.ServiceInfo, 0)
				}
				data.Records[entryType] = append(bucket, si)
			}
		}
	}

	var err error
	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}

	// Sort the results
	for _, b := range data.Records {
		sort.Slice(b, func(i, j int) bool {
			return b[i].OSSService.ReferenceResourceName < b[j].OSSService.ReferenceResourceName
		})
	}

	templateSrc := catalogGATemplateText

	t, err := template.New("CatalogGA").Parse(templateSrc)
	if err != nil {
		return err
	}
	err = t.Execute(w, &data)
	return err
}

var catalogGATemplateText = `=========== CatalogGA: all Active, GA Services/Components in the catalog ==========================
  Generated: {{.TimeStamp}}        Pattern: {{.Pattern}}
{{range $key, $value := .Records}}
------ Type={{$key}}      ({{len $value}} entries)
{{range $value}}
*** {{.OSSService.ReferenceResourceName}}  "{{.OSSService.ReferenceDisplayName}}"  Type={{.OSSService.GeneralInfo.EntryType}}  Status={{.OSSService.GeneralInfo.OperationalStatus}}
	OSS Tags:                 {{.OSSService.GeneralInfo.OSSTags}}
	Tags:                     {{.SourceMainCatalog.Tags}}
	Catalog Kind:             {{.SourceMainCatalog.Kind}}
	Canonical Name:           {{printf "%-25s" .OSSValidation.CanonicalName}}  Found in: {{.OSSValidation.CanonicalNameSources}}
	{{- range $index, $element := .OSSValidation.OtherNamesSources}} 
	Other Name Variant:       {{printf "%-25s" $index}}  Found in: {{$element}}
	{{- end}}
	Active:                   {{.SourceMainCatalog.Active}}

{{end}}
{{end}}
----------- End of CatalogGA Report ---------------------------------------

`
