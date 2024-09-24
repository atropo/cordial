/*
Copyright © 2024 ITRS Group

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reporter

import (
	"fmt"
	"io"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"

	"github.com/itrs-group/cordial/pkg/config"
)

type XLSXReporter struct {
	x                 *excelize.File
	w                 io.Writer
	sheet             string // current sheet name
	summarySheet      string
	topHeading        int
	leftHeading       int
	rightAlign        int
	dateStyle         int
	intStyle          int
	percentStyle      int
	plainStyle        int
	scambleNames      bool
	password          *config.Plaintext
	scrambleColumns   []string
	conditionalFormat []ConditionalFormat
	freezeColumn      string
	cond              map[string]int
	minColWidth       float64
	maxColWidth       float64
}

// ensure that *Table is a Reporter
var _ Reporter = (*XLSXReporter)(nil)

type ConditionalFormat struct {
	Test ConditionalFormatTest  `mapstructure:"test,omitempty"`
	Set  []ConditionalFormatSet `mapstructure:"set,omitempty"`
	// Else ConditionalFormatSet   `mapstructure:"else,omitempty"`
}

type ConditionalFormatTest struct {
	Columns   []string `mapstructure:"columns,omitempty"`
	Logical   string   `mapstructure:"logical,omitempty"` // "and", "all" or "or", "any"
	Condition string   `mapstructure:"condition,omitempty"`
	Type      string   `mapstructure:"type,omitempty"`
	Value     string   `mapstructure:"value,omitempty"`
}

type ConditionalFormatSet struct {
	Rows    string   `mapstructure:"rows,omitempty"`
	NotRows string   `mapstructure:"not-rows,omitempty"`
	Columns []string `mapstructure:"columns,omitempty"`
	Format  string   `mapstructure:"format,omitempty"`
}

// NewTableReporter returns a new Table reporter
func NewXLSXReporter(w io.Writer, options ...XLSXReporterOptions) (x *XLSXReporter) {
	opts := evalXLSXReportOptions(options...)

	x = &XLSXReporter{
		x:            excelize.NewFile(),
		w:            w,
		scambleNames: opts.scramble,
		password:     opts.password,
	}

	x.topHeading, _ = x.x.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"cccccc"},
			Pattern: 1,
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})

	x.leftHeading, _ = x.x.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})

	x.rightAlign, _ = x.x.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})

	x.dateStyle, _ = x.x.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
		CustomNumFmt: &opts.dateFormat,
	})

	x.intStyle, _ = x.x.NewStyle(&excelize.Style{
		NumFmt: opts.intFormat,
	})

	x.percentStyle, _ = x.x.NewStyle(&excelize.Style{
		NumFmt: opts.percentFormat,
	})

	x.plainStyle, _ = x.x.NewStyle(&excelize.Style{
		// Alignment: &excelize.Alignment{
		// 	Horizontal: "fill",
		// },
	})

	// set conditional formats
	ok, _ := x.x.NewConditionalStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{opts.okColour},
			Pattern: 1,
		},
	})
	warning, _ := x.x.NewConditionalStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{opts.warningColour},
			Pattern: 1,
		},
	})
	critical, _ := x.x.NewConditionalStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{opts.criticalColour},
			Pattern: 1,
		},
	})
	undefined, _ := x.x.NewConditionalStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{opts.undefinedColour},
			Pattern: 1,
		},
	})

	x.cond = map[string]int{
		"ok":        ok,
		"warning":   warning,
		"critical":  critical,
		"undefined": undefined,
	}

	x.summarySheet = opts.summarySheetName
	x.x.SetSheetName("Sheet1", opts.summarySheetName)
	x.x.SetColStyle(opts.summarySheetName, "A", x.leftHeading)
	x.x.SetColStyle(opts.summarySheetName, "B", x.rightAlign)

	x.minColWidth = opts.minColWidth
	x.maxColWidth = opts.maxColWidth

	return
}

type XLSXReporterOptions func(*xlsxReportOptions)

type xlsxReportOptions struct {
	scramble         bool
	password         *config.Plaintext
	summarySheetName string
	dateFormat       string
	intFormat        int
	percentFormat    int
	undefinedColour  string
	okColour         string
	warningColour    string
	criticalColour   string
	minColWidth      float64
	maxColWidth      float64
}

func evalXLSXReportOptions(options ...XLSXReporterOptions) (xo *xlsxReportOptions) {
	xo = &xlsxReportOptions{
		summarySheetName: "Summary",
		dateFormat:       "yyyy-mm-ddThh:MM:ss",
		intFormat:        1,
		percentFormat:    9,
		undefinedColour:  "BFBFBF",
		okColour:         "5BB25C",
		warningColour:    "F9B057",
		criticalColour:   "FF5668",
		minColWidth:      10.0,
		maxColWidth:      30.0,
	}
	for _, opt := range options {
		opt(xo)
	}
	return
}

func XLSXScramble(scramble bool) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.scramble = scramble
	}
}

func XLSXPassword(password *config.Plaintext) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.password = password
	}
}

func SummarySheetName(name string) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.summarySheetName = name
	}
}

func DateFormat(dateFormat string) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.dateFormat = dateFormat
	}
}

func IntFormat(format int) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.intFormat = format
	}
}

func PercentFormat(format int) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.percentFormat = format
	}
}

func SeverityColours(undefined, ok, warning, critical string) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.undefinedColour = undefined
		xro.okColour = ok
		xro.warningColour = warning
		xro.criticalColour = critical
	}
}

func MinColumnWidth(n float64) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.minColWidth = n
	}
}

func MaxColumnWidth(n float64) XLSXReporterOptions {
	return func(xro *xlsxReportOptions) {
		xro.maxColWidth = n
	}
}

func (x *XLSXReporter) SetReport(report Report) (err error) {
	title := report.Name

	x.scrambleColumns = report.ScrambleColumns
	x.conditionalFormat = report.ConditionalFormat
	x.freezeColumn = report.FreezeColumn

	if len(title) > 31 {
		log.Debug().Msgf("report title '%s' exceeds sheet name limit of 31 chars, truncating", title)
		title = title[:31]
	}
	idx, _ := x.x.GetSheetIndex(title)
	if idx != -1 && title != x.summarySheet {
		log.Error().Msgf("a sheet with the same name already exists, data will clash: '%s'", title)
	}
	if _, err = x.x.NewSheet(title); err != nil {
		return
	}
	x.sheet = title
	return
}

var percentRE = regexp.MustCompile(`^\d+\s*%$`)
var numRE = regexp.MustCompile(`^\d+$`)
var validcond = []string{
	"=",
	">",
	"<",
	">=",
	"<=",
	"<>",
}

func (x *XLSXReporter) UpdateTable(data ...[]string) {
	if len(data) == 0 {
		return
	}
	if x.scambleNames {
		scrambleColumns(x.scrambleColumns, data)
	}
	columns := data[0]
	var err error
	if err = x.x.SetSheetRow(x.sheet, "A1", &columns); err != nil {
		return
	}

	colwidths := []float64{}
	for _, c := range columns {
		colwidths = append(colwidths, limitWidth(len(c), x.minColWidth, x.maxColWidth))
	}

	rownum := 1
	for _, rowStrings := range data[1:] {
		row := []any{}
		for _, cell := range rowStrings {
			// test for a date/time in either ISO or Go layouts
			if t, err := time.Parse(time.RFC3339, cell); err == nil {
				row = append(row, t)
			} else if t, err := time.Parse(time.Layout, cell); err == nil {
				row = append(row, t)
			} else if percentRE.MatchString(cell) {
				var f float64
				fmt.Sscan(cell, &f)
				row = append(row, f/100.0)
			} else if numRE.MatchString(cell) {
				var n int64
				fmt.Sscan(cell, &n)
				row = append(row, n)
			} else {
				row = append(row, cell)
			}
		}
		rownum++ // increment first, starts at A2
		if err = x.x.SetSheetRow(x.sheet, fmt.Sprintf("A%d", rownum), &row); err != nil {
			return
		}

		// update styles
		for i, cell := range row {
			cellname, _ := excelize.CoordinatesToCellName(i+1, rownum)
			switch cell.(type) {
			case time.Time:
				x.x.SetCellStyle(x.sheet, cellname, cellname, x.dateStyle)
			case int64:
				x.x.SetCellStyle(x.sheet, cellname, cellname, x.intStyle)
			case float64:
				x.x.SetCellStyle(x.sheet, cellname, cellname, x.percentStyle)
			default:
				x.x.SetCellStyle(x.sheet, cellname, cellname, x.plainStyle)
			}
		}
		for j, c := range rowStrings {
			colwidths[j] = limitWidth(len(fmt.Sprint(c)), colwidths[j], x.maxColWidth)
		}

		// apply condition formatting
		//
		// each condition can have a `test` section, but must have a `set` section

	CONDFORMAT:
		for _, c := range x.conditionalFormat {
			// validate conditions allowed
			if !slices.Contains(validcond, c.Test.Condition) {
				log.Error().Msgf("sheet %s: invalid condition %s, skipping test", x.sheet, c.Test.Condition)
				continue
			}

			rowname := fmt.Sprint(row[0])

			// match is true unless all "Rows/NotRows" fail. if no
			// tests, then succeed regardless
			match := true
			format := "undefined"
			cols := []string{}

			for _, s := range c.Set {
				if s.NotRows != "" {
					match = false
					if ok, _ := path.Match(s.NotRows, rowname); !ok {
						format = s.Format
						cols = s.Columns
						match = true
						break
					}
				} else if s.Rows != "" {
					match = false
					if ok, _ := path.Match(s.Rows, rowname); ok {
						format = s.Format
						cols = s.Columns
						match = true
						break
					}
				} else {
					format = s.Format
					cols = s.Columns
				}
			}

			if !match {
				continue
			}

			tc := []string{}

			// if no set columns set, then use test columns
			if len(cols) == 0 {
				cols = c.Test.Columns
			}

			for _, t := range c.Test.Columns {
				i := slices.Index(columns, t)
				if i == -1 {
					log.Warn().Msgf("unknown column name %q, skipping conditional formatting for sheet %s", t, x.sheet)
					break CONDFORMAT
				}
				cellname, _ := excelize.CoordinatesToCellName(i+1, rownum, true)
				switch c.Test.Type {
				case "number":
					tc = append(tc, fmt.Sprintf("TEXT(%s, \"0\")%s%q", cellname, c.Test.Condition, c.Test.Value))
				default:
					tc = append(tc, fmt.Sprintf("%s%s%q", cellname, c.Test.Condition, c.Test.Value))
				}
			}

			logic := logicalWrapper(c.Test.Logical)
			formula := logic + "(" + strings.Join(tc, ",") + ")"

			r := []string{}
			for _, col := range cols {
				i := slices.Index(columns, col)
				if i == -1 {
					log.Warn().Msgf("unknown column name %q, skipping conditional formatting for sheet %s", col, x.sheet)
					break CONDFORMAT
				}
				cellname, _ := excelize.CoordinatesToCellName(i+1, rownum, true)
				r = append(r, cellname)
			}

			if err = x.x.SetConditionalFormat(x.sheet, strings.Join(r, ","), []excelize.ConditionalFormatOptions{
				{
					Type:     "formula",
					Criteria: formula,
					Format:   x.cond[format],
				},
			}); err != nil {
				log.Fatal().Err(err).Msgf("formula %s on %s", formula, strings.Join(r, ","))
			}
		}
	}

	// // mark up no data
	// if rownum == 1 {
	// 	if err = x.x.SetSheetRow(x.sheet, "A2", &[]string{"[No data]"}); err != nil {
	// 		return
	// 	}
	// 	colwidths[1] = colWidth(len("[No data]")*2, colwidths[1], maxColWidth)
	// }

	// set column widths
	for i, c := range colwidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		if err = x.x.SetColWidth(x.sheet, col, col, c); err != nil {
			return
		}
	}

	// x.x.SetColStyle(x.sheet, "D", x.dataColumnStyle)
	x.x.SetRowStyle(x.sheet, 1, 1, x.topHeading)

	if x.freezeColumn == "" {
		if err = x.x.SetPanes(x.sheet, &excelize.Panes{
			Freeze:      true,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
			Selection: []excelize.Selection{
				{SQRef: "A2", ActiveCell: "A2", Pane: "bottomLeft"},
			},
		}); err != nil {
			log.Error().Err(err).Msg("freeze top row")
		}
	} else {
		i := slices.Index(columns, x.freezeColumn)
		if i == -1 {
			log.Warn().Msgf("unknown column name %q, skipping freeze left pane for sheet %s", x.freezeColumn, x.sheet)
			return
		}
		// cellname is the first unlocked cell (so +2)
		cellname, _ := excelize.CoordinatesToCellName(i+2, 2, true)
		if err = x.x.SetPanes(x.sheet, &excelize.Panes{
			Freeze:      true,
			XSplit:      i + 1,
			YSplit:      1,
			TopLeftCell: cellname,
			ActivePane:  "topLeft",
			Selection: []excelize.Selection{
				{SQRef: cellname, ActiveCell: cellname, Pane: "bottomRight"},
			},
		}); err != nil {
			log.Error().Err(err).Msg("freeze top row")
		}
	}
}

func logicalWrapper(logic string) string {
	switch strings.ToLower(logic) {
	case "or", "any":
		return "OR"
	default:
		return "AND"
	}
}

func (x *XLSXReporter) AddHeadline(name, value string) {
	// nothing
}

func (x *XLSXReporter) Flush() {
	x.x.Write(x.w, excelize.Options{
		Password: x.password.String(),
	})
}

func (x *XLSXReporter) Close() {
	x.x.Close()
}

// a scale factor for the column width versus string len
const colScale = 1.25

// minimum column width
// const minColWidth = 10.0

func limitWidth(chars int, minWidth, maxWidth float64) float64 {
	w := colScale * float64(chars)
	if w > 255 {
		return 255
	}
	// if w < minWidth {
	// 	return minWidth
	// }
	w = max(min(w, maxWidth), minWidth)
	return w
}
