package ossrecord

import "strings"

// Taxonomy represents a product entry in the Unified Taxonomy
type Taxonomy struct {
	MajorUnitUTL10 string `json:"major_unit_utl10"`
	MinorUnitUTL15 string `json:"minor_unit_utl15"`
	MarketUTL17    string `json:"market_utl17"`
	PortfolioUTL20 string `json:"portfolio_utl20"`
	OfferingUTL30  string `json:"offering_utl30"`
}

// IsValid checks if a given Taxonomy object is valid (not empty and not "unknown")
func (t *Taxonomy) IsValid() bool {
	const unknown = "Unknown CH UTL"
	if (t.MajorUnitUTL10 == "" || strings.HasPrefix(t.MajorUnitUTL10, unknown)) &&
		(t.MinorUnitUTL15 == "" || strings.HasPrefix(t.MinorUnitUTL15, unknown)) &&
		(t.MarketUTL17 == "" || strings.HasPrefix(t.MarketUTL17, unknown)) &&
		(t.PortfolioUTL20 == "" || strings.HasPrefix(t.PortfolioUTL20, unknown)) &&
		(t.OfferingUTL30 == "" || strings.HasPrefix(t.OfferingUTL30, unknown)) {
		return false
	}
	return true
}
