package xlsx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/bububa/atomic-agents/components/document"
)

var (
	re_SHARED       = regexp.MustCompile(`xl/sharedStrings\.xml`)
	re_SHEET        = regexp.MustCompile(`xl/worksheets/sheet(\d+)\.xml`)
	re_SHEET_RELS   = regexp.MustCompile(`xl/worksheets/_rels/sheet(\d+)\.xml\.rels`)
	re_CHARTS       = regexp.MustCompile(`xl/charts/chart\d+\.xml`)
	re_IMAGES       = regexp.MustCompile(`xl/media/image\d+\.(?:png|gif|jpg|jpeg)`)
	re_DIAGRAMS     = regexp.MustCompile(`xl/diagrams/data\d+\.xml`)
	re_DRAWINGS     = regexp.MustCompile(`xl/drawings/drawing\d+\.xml`)
	re_DRAWING_RELS = regexp.MustCompile(`xl/drawings/_rels/drawing(\d+)\.xml\.rels`)
)

type Parser struct {
	password string
}

var _ document.Parser = (*Parser)(nil)

type Option func(*Parser)

func WithPassword(passwd string) Option {
	return func(p *Parser) {
		p.password = passwd
	}
}

type CellRange struct {
	minX int
	minY int
	maxX int
	maxY int
}

func (r CellRange) Insice(rowIdx int, cellIdx int) bool {
	if r.minX == 0 && r.minY == 0 && r.maxX == 0 && r.maxY == 0 {
		return false
	}
	if rowIdx < r.maxX || rowIdx > r.maxY {
		return false
	}
	return cellIdx >= r.minX && cellIdx <= r.maxX
}

// Parse try to parse a pdf content from a bytes.Reader and write to an io.Writer
func (p *Parser) Parse(ctx context.Context, reader *bytes.Reader, writer io.Writer) error {
	opts := make([]excelize.Options, 0, 1)
	if p.password != "" {
		opts = append(opts, excelize.Options{Password: p.password})
	}
	doc, err := excelize.OpenReader(reader, opts...)
	if err != nil {
		return err
	}
	defer doc.Close()
	sheets := doc.GetSheetList()
	for _, sheet := range sheets {
		rows, err := doc.Rows(sheet)
		if err != nil {
			return err
		}
		// mergeCells, err := doc.GetMergeCells(sheet)
		// if err != nil {
		// 	return err
		// }
		// merges := make([]CellRange, 0, len(mergeCells))
		// for _, cell := range mergeCells {
		// 	if len(cell) != 2 {
		// 		continue
		// 	}
		// 	arr := strings.Split(cell[0], ":")
		// 	if len(arr) != 2 {
		// 		continue
		// 	}
		// 	var cellRange CellRange
		// 	for i, v := range arr {
		// 		if rowIdx, cellIdx, err := excelize.CellNameToCoordinates(v); err != nil {
		// 			break
		// 		} else if i == 0 {
		// 			cellRange.minX = cellIdx
		// 			cellRange.minY = rowIdx
		// 		} else {
		// 			cellRange.maxX = cellIdx
		// 			cellRange.maxY = rowIdx
		// 		}
		// 	}
		// 	merges = append(merges, cellRange)
		// }
		var totalRows int
		for rowIdx := 0; rows.Next(); rowIdx++ {
			if rowIdx == 0 {
				fmt.Fprintf(writer, "# %s\n\n", sheet)
			}
			row, err := rows.Columns()
			if err != nil {
				return err
			}
			colCount := len(row)
			for colIdx, cellValue := range row {
				if colIdx == 0 {
					writer.Write([]byte("| "))
				}
				cellValue = strings.TrimSpace(document.EscapeMarkdown(document.StripUnprintable(cellValue)))
				if cell, err := excelize.CoordinatesToCellName(colIdx, rowIdx); err == nil {
					if styleID, err := doc.GetCellStyle(sheet, cell); err == nil {
						if style, err := doc.GetStyle(styleID); err == nil {
							if style.Font.Bold {
								cellValue = fmt.Sprintf("**%s**", cellValue)
							} else if style.Font.Strike {
								cellValue = fmt.Sprintf("~~%s~~", cellValue)
							} else if style.Font.Italic {
								cellValue = fmt.Sprintf("*%s*", cellValue)
							}
						}
					}
					if _, target, _ := doc.GetCellHyperLink(sheet, cell); target != "" {
						cellValue = fmt.Sprintf("[%s](%s)", cellValue, target)
					}
				}
				writer.Write([]byte(cellValue))
				if colIdx < colCount-1 {
					writer.Write([]byte(" | "))
				}
			}
			writer.Write([]byte(" |\n"))
			totalRows++
		}
		if totalRows > 0 {
			writer.Write(bytes.Repeat([]byte{'-'}, 100))
			writer.Write([]byte{'\n'})
		}
	}
	return nil
}
