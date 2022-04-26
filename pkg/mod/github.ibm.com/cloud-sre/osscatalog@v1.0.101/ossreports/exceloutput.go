package ossreports

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/stats"

	"github.com/tealeg/xlsx"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// ExcelReport is the type for manipulating Excel reports
type ExcelReport struct {
	w          io.Writer
	xlFile     *xlsx.File
	reportName string
}

// ExcelLink is a special type to create a cell that contains a hyperlink
type ExcelLink struct {
	URL  string
	Text string
}

// NonZero is a special type to create cell that displays a non-zero integer but stays empty if zero
type NonZero int

// NonNegative is a special type to create cell that displays a non-negtive integer but stays empty if negative
type NonNegative int

// CreateExcel create a new (empty) Excel report (in memory, no yet written to a file until it is populated with data)
func CreateExcel(w io.Writer, reportName string) *ExcelReport {
	var result = &ExcelReport{
		w:          w,
		xlFile:     xlsx.NewFile(),
		reportName: reportName,
	}
	return result
}

// AddEmptySheet add a new empty sheet to this Excel report
func (xl *ExcelReport) AddEmptySheet(headers []string, titleHeight float64, sheetName string) (*xlsx.Sheet, error) {
	sheet, err := xl.xlFile.AddSheet(sheetName)
	if err != nil {
		return nil, debug.WrapError(err, "%s/%s: error adding empty sheet in xlsx file", xl.reportName, sheetName)
	}

	xl.initializeTitleRowSlice(sheet, headers, titleHeight, sheetName)

	return sheet, nil
}

// AddSheet add a new sheet to this Excel report and populates it with the supplied data
func (xl *ExcelReport) AddSheet(data []interface{}, titleHeight float64, sheetName string) error {
	sheet, err := xl.xlFile.AddSheet(sheetName)
	if err != nil {
		return debug.WrapError(err, "%s/%s: error adding sheet in xlsx file", xl.reportName, sheetName)
	}

	if len(data) == 0 {
		topRow := sheet.AddRow()
		cell := topRow.AddCell()
		cell.SetString("*NO DATA*")
		return nil
	}

	columns := xl.initializeTitleRow(sheet, data[0], titleHeight, sheetName)

	var i = 1
	for _, r := range data {
		debug.Debug(debug.Reports, "%s/%s entry(%d): %v", xl.reportName, sheetName, i, r)
		err := xl.appendRow(sheet, r, columns)
		if err != nil {
			return err
		}
		i++
	}

	return nil
}

type reportMetaData struct {
	Name  string `column:"Attribute,35"`
	Value string `column:"Value,60"`
}

// Finalize writes this Excel report (and all its populated data/sheets) to a file and closes it
func (xl *ExcelReport) Finalize() error {
	// Add a special sheet with report meta-data
	metadata := make([]interface{}, 0, 10)
	metadata = append(metadata, &reportMetaData{"Report Name:", xl.reportName})
	metadata = append(metadata, &reportMetaData{"Generated:", options.GlobalOptions().LogTimeStamp})
	metadata = append(metadata, &reportMetaData{"Optional run actions ENABLED:", fmt.Sprintf("%v", ossrunactions.ListEnabledRunActionNames())})
	metadata = append(metadata, &reportMetaData{"Optional run actions DISABLED:", fmt.Sprintf("%v", ossrunactions.ListDisabledRunActionNames())})
	stats := stats.GetGlobalActualStats()
	metadata = append(metadata, &reportMetaData{"Final number of valid OSS records:", fmt.Sprintf("services/components:%-4d  segments:%-4d  tribes:%-4d  environments:%-4d  schema:%-4d\n", stats.NumServicesActual, stats.NumSegmentsActual, stats.NumTribesActual, stats.NumEnvironmentsActual, stats.NumSchemaActual)})
	metadata = append(metadata, &reportMetaData{"Total errors:", fmt.Sprintf("%d", debug.CountErrors())})
	metadata = append(metadata, &reportMetaData{"Total critical issues:", fmt.Sprintf("%d", debug.CountCriticals())})
	err := xl.AddSheet(metadata, 30.0, "Report Info")
	if err != nil {
		return debug.WrapError(err, "%s: error generating report info sheet", xl.reportName)
	}

	err = xl.xlFile.Write(xl.w)
	if err != nil {
		return debug.WrapError(err, "%s: error writing xlsx file", xl.reportName)
	}

	// clear the attribute to prevent accidental repeated calls to Finalize
	xl.xlFile = nil

	return nil
}

// GenerateExcel generates a Excel file with a single sheet showing the report data
func GenerateExcel(w io.Writer, data []interface{}, titleHeight float64, reportName string) error {
	var err error
	xl := CreateExcel(w, reportName)

	err = xl.AddSheet(data, titleHeight, reportName)
	if err != nil {
		return err
	}

	err = xl.Finalize()
	if err != nil {
		return err
	}

	return err
}

func (xl *ExcelReport) appendRow(sheet *xlsx.Sheet, r interface{}, columns []int) error {
	slice := make([]interface{}, 0, len(columns))
	val := reflect.ValueOf(r)
	val = reflect.Indirect(val)
	for _, i := range columns {
		field := val.Field(i)
		v := field.Interface()
		slice = append(slice, v)
	}
	return xl.AppendRowSlice(sheet, slice)
}

// AppendRowSlice appends one row to a sheet of the report, from data supplied in a slice
// The order of entries in the data slice MUST be the same as the headers specificed in AddEmptySheet
func (xl *ExcelReport) AppendRowSlice(sheet *xlsx.Sheet, data []interface{}) error {
	row := sheet.AddRow()
	for _, cell := range data {
		switch v := cell.(type) {
		case string:
			row.AddCell().SetString(v)
		case int:
			row.AddCell().SetInt64(int64(v))
		case int8:
			row.AddCell().SetInt64(int64(v))
		case int16:
			row.AddCell().SetInt64(int64(v))
		case int32:
			row.AddCell().SetInt64(int64(v))
		case int64:
			row.AddCell().SetInt64(int64(v))
		case uint:
			row.AddCell().SetInt64(int64(v))
		case uint8:
			row.AddCell().SetInt64(int64(v))
		case uint16:
			row.AddCell().SetInt64(int64(v))
		case uint32:
			row.AddCell().SetInt64(int64(v))
		case uint64:
			row.AddCell().SetInt64(int64(v))
		case bool:
			row.AddCell().SetBool(v)
		case float32:
			row.AddCell().SetFloat(float64(v))
		case float64:
			row.AddCell().SetFloat(v)
		case ExcelLink:
			if v.URL != "" {
				cell := row.AddCell()
				cell.SetFormula(fmt.Sprintf(`HYPERLINK("%s","%s")`, v.URL, v.Text))
				style := cell.GetStyle()
				style.Font.Underline = true
				style.Font.Color = "0x1100E4" // blue
				//	style.Font.Color = "0x4A7FFF" // slightly lighter blue
				//	style.Font.Color = "16711680" // purple
				style.ApplyFont = true
				cell.SetStyle(style)
			} else {
				row.AddCell().SetString(v.Text)
			}
		case NonZero:
			if v != 0 {
				row.AddCell().SetInt64(int64(v))
			} else {
				row.AddCell().SetString("")
			}
		case NonNegative:
			if v >= 0 {
				row.AddCell().SetInt64(int64(v))
			} else {
				row.AddCell().SetString("")
			}
		default:
			panic(fmt.Sprintf("ossreport ExcelReport does not support cell type=%T", cell))
		}
	}
	return nil
}

var xlTagPattern = regexp.MustCompile(`([^",]+)(?:,([0-9\.]+))?(,center)?`)

func (xl *ExcelReport) initializeTitleRow(sheet *xlsx.Sheet, data interface{}, titleHeight float64, sheetName string) (columns []int) {
	typ := reflect.TypeOf(data)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	numField := typ.NumField()
	headers := make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("column")
		if tag == "" {
			tag = field.Name
		}
		headers = append(headers, tag)
	}
	return xl.initializeTitleRowSlice(sheet, headers, titleHeight, sheetName)
}

func (xl *ExcelReport) initializeTitleRowSlice(sheet *xlsx.Sheet, headers []string, titleHeight float64, sheetName string) (columns []int) {
	columns = make([]int, 0, len(headers))
	topRow := sheet.AddRow()
	colIndex := 0
	for i, tag := range headers {
		var width = 10.0
		var center bool
		var err error
		parsed := xlTagPattern.FindStringSubmatch(tag)
		if parsed == nil || parsed[1] == "" {
			panic(fmt.Sprintf(`%s/%s report - found invalid struct tag "%s" at index %d`, xl.reportName, sheetName, tag, i))
		}
		title := parsed[1]
		if parsed[2] != "" {
			width, err = strconv.ParseFloat(parsed[2], 64)
			if err != nil {
				panic(fmt.Sprintf(`%s/%s report - cannot parse width in struct tag "%s" at index %d: %v`, xl.reportName, sheetName, tag, i, err))
			}
		}
		if parsed[3] == ",center" {
			center = true
		}
		if title == "-" {
			debug.Debug(debug.Reports, `ossreports.xlInitializeTitleRow() - ignoring column at index %d  tag=%q  title=%q  width=%v`, i, tag, title, width)
			continue
		} else {
			debug.Debug(debug.Reports, `ossreports.xlInitializeTitleRow() - adding column index %d  tag=%q  title=%q  width=%v`, i, tag, title, width)
		}
		columns = append(columns, i)
		cell := topRow.AddCell()
		cell.SetString(title)
		// SetColWidth no longer returns an error:
		//err = sheet.SetColWidth(colIndex, colIndex, width)
		//if err != nil {
		//	panic(fmt.Sprintf(`%s/%s report - cannot set column width %v for column %d: %v`, xl.reportName, sheetName, width, colIndex, err))
		//}
		sheet.SetColWidth(colIndex, colIndex, width)
		if center {
			col := sheet.Col(colIndex)
			colStyle := col.GetStyle()
			colStyle.Alignment.Horizontal = "center"
			colStyle.ApplyAlignment = true
			col.SetStyle(colStyle)
			if debug.IsDebugEnabled(debug.Reports) {
				colStyle := col.GetStyle()
				debug.Debug(debug.Reports, "Column style at index %d (center): %+v", colIndex, colStyle)
			}
		}
		style := cell.GetStyle()
		//		debug.Debug(debug.Reports, "Cell Style: %#v", style)
		style.Font.Bold = true
		style.ApplyFont = true
		style.Alignment.Vertical = "top"
		style.Alignment.WrapText = true
		style.ApplyAlignment = true
		cell.SetStyle(style)
		colIndex++
		if debug.IsDebugEnabled(debug.Reports) {
			style := cell.GetStyle()
			debug.Debug(debug.Reports, "Cell style at index %d: %+v", colIndex, style)
		}
	}
	topRow.SetHeight(titleHeight)
	return columns
}
