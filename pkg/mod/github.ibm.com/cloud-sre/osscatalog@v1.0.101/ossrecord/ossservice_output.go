package ossrecord

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Header returns a one-line string representing this OSS record
func (oss *OSSService) Header() string {
	var result string
	var onboardingPhase string
	if oss.GeneralInfo.OSSOnboardingPhase != "" {
		onboardingPhase = fmt.Sprintf("OSSOnboardingPhase=%s", oss.GeneralInfo.OSSOnboardingPhase)
	}
	result = fmt.Sprintf("NAME=%s  TYPE=%s/%s %s OSSTAGS=%s  DISPLAY_NAME=\"%s\"\n",
		oss.ReferenceResourceName,
		oss.GeneralInfo.EntryType,
		oss.GeneralInfo.OperationalStatus,
		onboardingPhase,
		&oss.GeneralInfo.OSSTags,
		oss.ReferenceDisplayName)
	return result
}

// JSON returns a JSON-style string representation of the entire OSS record (with no header), multi-line
func (oss *OSSService) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_service": `)
	json, _ := json.MarshalIndent(oss, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	result.WriteString("\n")
	result.WriteString("  }")
	return result.String()
}

// String returns a short string representation of this OSS record
func (oss *OSSService) String() string {
	//	return fmt.Sprintf(`OSS(%s)`, oss.ReferenceResourceName)
	if oss.ReferenceResourceName == "" {
		return "<*** empty OSS record***>"
	}
	return string(oss.ReferenceResourceName)
}

// String returns a string representation of this Person record
func (p Person) String() string {
	var result strings.Builder
	if p.Name != "" {
		result.WriteString(p.Name)
		if p.W3ID != "" {
			result.WriteString("  ")
		}
	}
	if p.W3ID != "" {
		result.WriteString("<")
		result.WriteString(p.W3ID)
		result.WriteString(">")
	}
	return result.String()
}
