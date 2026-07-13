package questions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseCSVImportFile(t *testing.T) {
	input := "题目,题型,选项,答案,科目\n下列哪项正确？,single,\"A. 正确\nB. 错误\",正确,测试\n缺少答案,single,,,测试\n"
	items, report, err := parseImportReader(".csv", strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(items) != 1 || report.Total != 2 || report.Valid != 1 || report.Invalid != 1 {
		t.Fatalf("unexpected report: %#v, items=%d", report, len(items))
	}
	if len(items[0].Options) != 2 || items[0].Options[0].Text != "正确" || items[0].Answer != "正确" {
		t.Fatalf("unexpected parsed question: %#v", items[0])
	}
}

func newXLSXReader(rows [][]string) (*bytes.Buffer, error) {
	file := excelize.NewFile()
	defer file.Close()
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return nil, err
			}
			if err := file.SetCellValue("Sheet1", cell, value); err != nil {
				return nil, err
			}
		}
	}
	buffer := bytes.NewBuffer(nil)
	if _, err := file.WriteTo(buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

func TestParseXLSXImportFile(t *testing.T) {
	workbook, err := newXLSXReader([][]string{{"question", "type", "options", "answer", "collectedAt"}, {"1+1= ?", "single", "A. 2|B. 3", "2", "2026-07-14"}})
	if err != nil {
		t.Fatalf("create xlsx: %v", err)
	}
	items, report, err := parseImportReader(".xlsx", workbook)
	if err != nil {
		t.Fatalf("parse xlsx: %v", err)
	}
	if len(items) != 1 || report.Valid != 1 || items[0].CollectedAt == nil || len(items[0].Options) != 2 {
		t.Fatalf("unexpected xlsx result: %#v, report=%#v", items, report)
	}
}
