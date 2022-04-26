// Code to read the "All parts with set and group code of BLUMX.xlsx" report from Cognos

package partsinput

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// minPartsRecordsCount minimum number of records expected - for sanity checks
var minPartsRecordsCount = 5000

// Entry represents the information from one entry in the parts input file
type Entry struct {
	PartNumber      string
	WWPC            string
	WWPCDescription string
	ProductID       string
	Division        string
}

const (
	colRowNumber = iota
	colBrand
	colSet
	colSetDscrLong
	colGroup
	colGroupDscrLong
	colTradeName
	colTradeNameDscrLong
	colWWPC
	colWWPCDscrLong
	colIBMProdID
	colIBMProdIDDscr
	colGBT30
	colCCId
	colCCIdDscr
	colSubID
	colSubIDDscrFull
	colPPT
	colPartNumber
	colPartDscr
	colPartDscrLong
	colCuID
	colCuDscr
	colCuQty
	colFCSDate
	colEOLDate
	colAddDate
	colCtrctProg
	colRevnStream
	colDistribtnCode
	colDivision
	colPrftCntr
	colPH1
	colPH2
	colPH3
	colPH4
	colSMG1
	colSMG2
	colSMG3
	colSMG4
	colSMG5
	colAudienceMask
	colSalesStatus
	colMatlType
	colMAtlGrp
	colItemCatGrp
	colAAG
	colERO
	colECCN
	colRenwlMdl
	colPricingTierMdl
	colProvisngHold
	colPriceDurtn
	colBillgUpfrnt
	colBillgAnl
	colBillgMthly
	colBillgQrtly
	colBillgEvent
	colSerNum
	colBNPM
	colPartDscrFull
)

var allPartNumbers map[string]*Entry
var allProductIDs map[string][]*Entry

func checkColTitle(row *xlsx.Row, index int, expected string, output *strings.Builder) {
	if row.Cells[index].Value != expected {
		output.WriteString(fmt.Sprintf(`ListPartsRecords(): unexpected column title(%2d): expected "%s"   found "%s"\n`, index, expected, row.Cells[index].Value))
	}
}

// ReadPartsInputFile reads all Parts records from the input spreadsheet and indexes them by part number
func ReadPartsInputFile(fname string) error {
	var countRecords = 0
	var countDups = 0
	partsMap := make(map[string]*Entry)
	pidsMap := make(map[string][]*Entry)

	// Use explicit row limit to get around a bug where the input sheet might contain an invalid DimensionRef
	file, err := xlsx.OpenFileWithRowLimit(fname, 10000 /* xlsx.NoRowLimit */)
	if err != nil {
		return debug.WrapError(err, "Cannot open the parts input file")
	}

	sheetName := "Page1"
	sheet, ok := file.Sheet[sheetName]
	if !ok {
		return fmt.Errorf(`Cannot find sheet "%s" in spreadsheet "%s"`, sheetName, fname)
	}

	for ix, row := range sheet.Rows {
		if len(row.Cells) < 14 {
			debug.PrintError(`ReadPartsInputFile(): ignoring row with not enough columns: %d: %v`, ix, row)
			continue
		}
		if ix == 0 {
			// Special checks for title row
			output := &strings.Builder{}
			checkColTitle(row, colRowNumber, "Row Number", output)
			checkColTitle(row, colWWPC, "WWPC", output)
			checkColTitle(row, colWWPCDscrLong, "WWPC Dscr Long", output)
			checkColTitle(row, colIBMProdID, "IBM Prod Id", output)
			checkColTitle(row, colPartNumber, "Part Number", output)
			checkColTitle(row, colDivision, "Division", output)
			str := output.String()
			if len(str) > 0 {
				return fmt.Errorf("ReadPartsInputFile(): cannot parse the input\n%s", str)
			}
		} else {
			cells := row.Cells

			e := Entry{}
			e.PartNumber = strings.TrimSpace(cells[colPartNumber].Value)
			e.WWPC = strings.TrimSpace(cells[colWWPC].Value)
			e.WWPCDescription = strings.TrimSpace(cells[colWWPCDscrLong].Value)
			e.ProductID = ossrecord.NormalizeProductID(cells[colIBMProdID].Value)
			e.Division = strings.TrimSpace(cells[colDivision].Value)

			countRecords++

			if e.PartNumber != "" {
				if e1, found := partsMap[e.PartNumber]; found {
					debug.PrintError(`ReadPartsInputFile(): found duplicate part number "%s": entry1=%+v    entry2=%+v`, e.PartNumber, e1, e)
					countDups++
				} else {
					partsMap[e.PartNumber] = &e
				}
			} else {
				debug.PrintError(`ReadPartsInputFile(): found entry with empty part number: %+v`, e)
			}

			if e.ProductID != "" {
				entries, _ := pidsMap[e.ProductID]
				entries = append(entries, &e)
				pidsMap[e.ProductID] = entries
			}
		}
	}

	if countRecords < minPartsRecordsCount {
		return fmt.Errorf(`Unexpectedly low number of records found in the Parts input spreadsheet "%s" :  %d records (expected at least %d)`, fname, countRecords, minPartsRecordsCount)
	}
	if countDups == 0 {
		debug.Info(`Completed reading the Parts input spreadsheet "%s" :  %d records (no duplicates found)`, fname, countRecords)
	} else {
		debug.Info(`Completed reading the Parts input spreadsheet "%s" :  %d records  -  *** with %d duplicates ***`, fname, countRecords, countDups)
	}

	allPartNumbers = partsMap
	allProductIDs = pidsMap
	return nil
}

// LookupPartNumber returns the Entry record associated with a given part number
func LookupPartNumber(part string) (*Entry, bool) {
	e, ok := allPartNumbers[part]
	return e, ok
}

// LookupProductID returns the Entry records associated with a given product ID
func LookupProductID(pid string) ([]*Entry, bool) {
	if entries, ok := allProductIDs[pid]; ok && len(entries) > 0 {
		return entries, ok
	}
	return nil, false
}

// HasPartNumbers returns true if the table of part numbers has been initialized and loaded
func HasPartNumbers() bool {
	return allPartNumbers != nil
}
