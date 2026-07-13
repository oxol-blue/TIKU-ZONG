package questions

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	maxImportFileSize = 10 << 20
	maxImportRows     = 1000
)

// ImportRowError describes a rejected row without leaking uploaded file contents.
type ImportRowError struct {
	Row      int    `json:"row"`
	Question string `json:"question"`
	Message  string `json:"message"`
}

// FileImportReport is returned to the admin UI after a CSV/XLSX import.
type FileImportReport struct {
	Total      int              `json:"total"`
	Valid      int              `json:"valid"`
	Created    int              `json:"created"`
	Duplicates int              `json:"duplicates"`
	Invalid    int              `json:"invalid"`
	Preview    []ImportInput    `json:"preview"`
	Errors     []ImportRowError `json:"errors"`
}

var optionPrefix = regexp.MustCompile(`^\s*([A-Za-z0-9]+)\s*[\.、:：]\s*(.+)$`)

func parseImportFile(fileHeader *multipart.FileHeader) ([]ImportInput, FileImportReport, error) {
	if fileHeader == nil {
		return nil, FileImportReport{}, errors.New("file is required")
	}
	if fileHeader.Size > maxImportFileSize {
		return nil, FileImportReport{}, fmt.Errorf("file must not exceed %d MB", maxImportFileSize>>20)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, FileImportReport{}, fmt.Errorf("open import file: %w", err)
	}
	defer file.Close()
	return parseImportReader(filepath.Ext(fileHeader.Filename), io.LimitReader(file, maxImportFileSize+1))
}

func parseImportReader(extension string, reader io.Reader) ([]ImportInput, FileImportReport, error) {
	var rows [][]string
	var err error
	switch strings.ToLower(strings.TrimSpace(extension)) {
	case ".csv":
		csvReader := csv.NewReader(reader)
		csvReader.FieldsPerRecord = -1
		rows, err = csvReader.ReadAll()
	case ".xlsx":
		workbook, openErr := excelize.OpenReader(reader)
		if openErr != nil {
			return nil, FileImportReport{}, fmt.Errorf("invalid XLSX file: %w", openErr)
		}
		defer workbook.Close()
		sheets := workbook.GetSheetList()
		if len(sheets) == 0 {
			return nil, FileImportReport{}, errors.New("XLSX file has no worksheet")
		}
		rows, err = workbook.GetRows(sheets[0])
	default:
		return nil, FileImportReport{}, errors.New("only .csv and .xlsx files are supported")
	}
	if err != nil {
		return nil, FileImportReport{}, fmt.Errorf("read import file: %w", err)
	}
	return parseImportRows(rows)
}

func parseImportRows(rows [][]string) ([]ImportInput, FileImportReport, error) {
	if len(rows) < 2 {
		return nil, FileImportReport{}, errors.New("file must contain a header row and at least one question")
	}
	headers := make(map[string]int, len(rows[0]))
	for index, value := range rows[0] {
		key := normalizeHeader(value)
		if key != "" {
			headers[key] = index
		}
	}
	if _, ok := headerIndex(headers, "question", "题目"); !ok {
		return nil, FileImportReport{}, errors.New("missing required column: question/题目")
	}
	if _, ok := headerIndex(headers, "answer", "答案"); !ok {
		return nil, FileImportReport{}, errors.New("missing required column: answer/答案")
	}

	report := FileImportReport{Preview: make([]ImportInput, 0, 20), Errors: make([]ImportRowError, 0)}
	items := make([]ImportInput, 0, minImportRows(len(rows)-1))
	for rowIndex, row := range rows[1:] {
		if report.Total >= maxImportRows {
			return nil, report, fmt.Errorf("file contains more than %d data rows", maxImportRows)
		}
		if emptyRow(row) {
			continue
		}
		report.Total++
		item, rowErr := parseImportRow(row, headers)
		if rowErr != nil {
			report.Invalid++
			if len(report.Errors) < 100 {
				report.Errors = append(report.Errors, ImportRowError{Row: rowIndex + 2, Question: valueAt(row, headers, "question", "题目"), Message: rowErr.Error()})
			}
			continue
		}
		report.Valid++
		if len(report.Preview) < 20 {
			report.Preview = append(report.Preview, item)
		}
		items = append(items, item)
	}
	if report.Valid == 0 {
		return nil, report, errors.New("no valid question rows found")
	}
	return items, report, nil
}

func parseImportRow(row []string, headers map[string]int) (ImportInput, error) {
	item := ImportInput{
		Question:  strings.TrimSpace(valueAt(row, headers, "question", "题目")),
		Type:      strings.TrimSpace(valueAt(row, headers, "type", "题型")),
		Answer:    strings.TrimSpace(valueAt(row, headers, "answer", "答案")),
		AnswerRaw: strings.TrimSpace(valueAt(row, headers, "answerraw", "原始答案")),
		Platform:  strings.TrimSpace(valueAt(row, headers, "platform", "平台")),
		Subject:   strings.TrimSpace(valueAt(row, headers, "subject", "科目")),
		Source:    strings.TrimSpace(valueAt(row, headers, "source", "来源")),
	}
	if item.Question == "" || item.Answer == "" {
		return ImportInput{}, errors.New("question and answer are required")
	}
	options, err := parseOptions(valueAt(row, headers, "options", "选项"))
	if err != nil {
		return ImportInput{}, err
	}
	item.Options = options
	if rawTime := strings.TrimSpace(valueAt(row, headers, "collectedat", "采集时间")); rawTime != "" {
		parsed, timeErr := parseCollectedAt(rawTime)
		if timeErr != nil {
			return ImportInput{}, errors.New("invalid collectedAt/采集时间")
		}
		item.CollectedAt = &parsed
	}
	return item, nil
}

func parseOptions(value string) ([]OptionInput, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if strings.HasPrefix(value, "[") {
		var items []OptionInput
		if err := json.Unmarshal([]byte(value), &items); err != nil {
			return nil, errors.New("options JSON is invalid")
		}
		for index := range items {
			items[index].Key = strings.TrimSpace(items[index].Key)
			items[index].Text = strings.TrimSpace(items[index].Text)
			if items[index].Key == "" || items[index].Text == "" {
				return nil, errors.New("each option requires key and text")
			}
		}
		return items, nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == '\r' || r == '|' })
	items := make([]OptionInput, 0, len(parts))
	for index, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key := string(rune('A' + index))
		text := part
		if match := optionPrefix.FindStringSubmatch(part); len(match) == 3 {
			key, text = match[1], match[2]
		}
		items = append(items, OptionInput{Key: key, Text: strings.TrimSpace(text)})
	}
	return items, nil
}

func parseCollectedAt(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, errors.New("invalid time")
}

func normalizeHeader(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")))
}

func headerIndex(headers map[string]int, aliases ...string) (int, bool) {
	for _, alias := range aliases {
		if index, ok := headers[normalizeHeader(alias)]; ok {
			return index, true
		}
	}
	return 0, false
}

func valueAt(row []string, headers map[string]int, aliases ...string) string {
	index, ok := headerIndex(headers, aliases...)
	if !ok || index >= len(row) {
		return ""
	}
	return row[index]
}

func emptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func minImportRows(value int) int {
	if value < maxImportRows {
		return value
	}
	return maxImportRows
}
