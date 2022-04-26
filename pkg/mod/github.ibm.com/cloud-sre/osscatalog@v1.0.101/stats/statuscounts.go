package stats

import (
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// StatusCounts captures a count of green/yellow/red status items
type StatusCounts struct {
	CRN struct {
		Green   int
		Yellow  int
		Red     int
		Unknown int
	}
	Overall struct {
		Green   int
		Yellow  int
		Red     int
		Unknown int
	}
}

// updateCRNStatus extracts an OSSEntry's OSSTags related to CRN status and updates the StatusCounts struct accordingly
func (sc *StatusCounts) updateStatus(e ossrecord.OSSEntry) {
	crn := e.GetOSSTags().GetCRNStatus()
	switch crn {
	case osstags.StatusCRNGreen:
		sc.CRN.Green++
	case osstags.StatusCRNYellow:
		sc.CRN.Yellow++
	case osstags.StatusCRNRed:
		sc.CRN.Red++
	default:
		sc.CRN.Unknown++
	}
	overall := e.GetOSSTags().GetOverallStatus()
	switch overall {
	case osstags.StatusGreen:
		sc.Overall.Green++
	case osstags.StatusYellow:
		sc.Overall.Yellow++
	case osstags.StatusRed:
		sc.Overall.Red++
	default:
		sc.Overall.Unknown++
	}
}

// add adds the counts from one StatusCounts record into another
func (sc *StatusCounts) add(sc2 *StatusCounts) {
	sc.CRN.Green += sc2.CRN.Green
	sc.CRN.Yellow += sc2.CRN.Yellow
	sc.CRN.Red += sc2.CRN.Red
	sc.CRN.Unknown += sc2.CRN.Unknown
	sc.Overall.Green += sc2.Overall.Green
	sc.Overall.Yellow += sc2.Overall.Yellow
	sc.Overall.Red += sc2.Overall.Red
	sc.Overall.Unknown += sc2.Overall.Unknown
}
