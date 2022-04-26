package ownership

import (
	"html/template"
	"io"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
)

// RunReport generates the "ownership" report for all known services/components and writes it to the specified Writer
func RunReport(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	data := struct {
		TimeStamp string
		Pattern   string
		Records   []*ossmerge.ServiceInfo
	}{
		TimeStamp: timeStamp,
		Pattern:   pattern.String(),
	}

	var err error
	data.Records, err = ossmerge.GetAllServices(pattern)
	if err != nil {
		return err
	}

	templateSrc := ownershipTemplateText

	t, err := template.New("Ownership").Parse(templateSrc)
	if err != nil {
		return err
	}
	err = t.Execute(w, &data)
	return err
}

var ownershipTemplateText = `=========== Services/Components Ownership Report ==========================
  Generated: {{.TimeStamp}}        Pattern: {{.Pattern}}
{{range .Records}}
*** {{.OSSService.ReferenceResourceName}}  "{{.OSSService.ReferenceDisplayName}}"  Type={{.OSSService.GeneralInfo.EntryType}}  Status={{.OSSService.GeneralInfo.OperationalStatus}}
	OSS Onboarding Phase:     {{.OSSService.GeneralInfo.OSSOnboardingPhase}}
	OSS Tags:                 {{.OSSService.GeneralInfo.OSSTags}}
	Canonical Name:           {{printf "%-25s" .OSSValidation.CanonicalName}}  Found in: {{.OSSValidation.CanonicalNameSources}}
	{{- range $index, $element := .OSSValidation.OtherNamesSources}} 
	Other Name Variant:       {{printf "%-25s" $index}}  Found in: {{$element}}
	{{- end}}
	Segment:                  {{printf "%-25s" .OSSService.Ownership.SegmentName}}  Segment Owner: {{.OSSService.Ownership.SegmentOwner}}
	Tribe:                    {{printf "%-25s" .OSSService.Ownership.TribeName}}  Tribe Owner: {{.OSSService.Ownership.TribeOwner}}
	Offering Manager:         {{.OSSService.Ownership.OfferingManager}}
	Onboarding Contact:       {{.OSSService.Compliance.OnboardingContact}}
	Development Manager:      {{.OSSService.Ownership.DevelopmentManager}}
	Architecture Focal:       {{.OSSService.Compliance.ArchitectureFocal}}
	Support Manager:          {{.OSSService.Support.Manager}}
	Operations Manager:       {{.OSSService.Operations.Manager}}
	Catalog Provider:         {{.OSSService.CatalogInfo.Provider}}
	Catalog Provider Contact: {{.OSSService.CatalogInfo.ProviderContact}}  {{.OSSService.CatalogInfo.ProviderSupportEmail}}  {{.OSSService.CatalogInfo.ProviderPhone}}

{{end}}
----------- End of Ownership Report ---------------------------------------

`
