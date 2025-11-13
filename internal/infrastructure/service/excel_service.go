package service

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(data map[string][]any) (*bytes.Buffer, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data provided")
	}

	headers := make([]string, 0, len(data))
	rowCount := -1
	for header, values := range data {
		if rowCount == -1 {
			rowCount = len(values)
		} else if len(values) != rowCount {
			return nil, 0, fmt.Errorf("column %q length %d mismatched expected %d", header, len(values), rowCount)
		}
		headers = append(headers, header)
	}

	file := excelize.NewFile()
	sheetName := "Sheet1"
	index := file.GetActiveSheetIndex()
	file.SetSheetName(file.GetSheetName(index), sheetName)

	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return nil, 0, fmt.Errorf("set header %q: %w", header, err)
		}
	}

	for row := 0; row < rowCount; row++ {
		for colIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row+2)
			values := data[header]
			if row >= len(values) {
				continue
			}
			if err := file.SetCellValue(sheetName, cell, normalizeExcelValue(values[row])); err != nil {
				return nil, 0, fmt.Errorf("set cell %s: %w", cell, err)
			}
		}
	}

	// Configure sheet view (optional)
	_ = file.SetSheetView(sheetName, 0, &excelize.ViewOptions{})

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		return nil, 0, fmt.Errorf("write excel: %w", err)
	}
	return &buf, rowCount, nil
}

func normalizeExcelValue(value any) any {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

