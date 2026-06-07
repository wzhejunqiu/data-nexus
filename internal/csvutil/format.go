package csvutil

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func NewReader(r io.Reader, opts model.CSVFormatOptions) (*csv.Reader, error) {
	o := opts.Normalized()
	if o.Delimiter == "" {
		return nil, fmt.Errorf("delimiter is required")
	}
	delimiter, size := utf8.DecodeRuneInString(o.Delimiter)
	if size == 0 {
		return nil, fmt.Errorf("invalid delimiter")
	}
	if o.QuoteChar != `"` && o.QuoteChar != `'` {
		return nil, fmt.Errorf("quote char must be \" or '")
	}
	reader := csv.NewReader(r)
	reader.Comma = delimiter
	reader.LazyQuotes = o.LazyQuotes
	reader.TrimLeadingSpace = o.TrimLeadingSpace
	if o.CommentChar != "" {
		comment, size := utf8.DecodeRuneInString(o.CommentChar)
		if size == 0 {
			return nil, fmt.Errorf("invalid comment char")
		}
		reader.Comment = comment
	}
	return reader, nil
}

func ReadPreview(path string, maxRows int, opts model.CSVFormatOptions) (*model.CSVPreview, error) {
	data, err := readFileBytes(path, opts.Encoding)
	if err != nil {
		return nil, err
	}
	reader, err := NewReader(bytes.NewReader(data), opts)
	if err != nil {
		return nil, model.ErrInvalidRequest(err.Error())
	}
	if maxRows <= 0 {
		maxRows = 20
	}

	var headers []string
	var rows []map[string]any
	rowNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, model.ErrInvalidRequest("csv parse error: " + err.Error())
		}
		if rowNum == 0 && opts.HasHeader {
			headers = append([]string{}, record...)
			rowNum++
			continue
		}
		if len(headers) == 0 {
			headers = make([]string, len(record))
			for i := range record {
				headers[i] = fmt.Sprintf("column_%d", i+1)
			}
		}
		row := make(map[string]any, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
		rowNum++
		if len(rows) >= maxRows {
			break
		}
	}
	if headers == nil {
		headers = []string{}
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	return &model.CSVPreview{
		Headers:   headers,
		Rows:      rows,
		RowCount:  len(rows),
		HasHeader: opts.HasHeader,
	}, nil
}

func ReadFileBytes(path, encoding string) ([]byte, error) {
	data, err := readRawFile(path)
	if err != nil {
		return nil, model.ErrInvalidPath(err.Error())
	}
	switch strings.ToLower(encoding) {
	case "", "utf-8", "utf-8-bom":
		return stripBOM(data), nil
	default:
		return nil, model.ErrInvalidRequest("unsupported encoding: " + encoding)
	}
}

func readFileBytes(path, encoding string) ([]byte, error) {
	return ReadFileBytes(path, encoding)
}

func stripBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

func formatCell(val any, nullValue string) string {
	if val == nil {
		return nullValue
	}
	switch v := val.(type) {
	case string:
		return v
	case map[string]any:
		if t, ok := v["type"].(string); ok && t == "blob" {
			if size, ok := v["size"].(float64); ok {
				return fmt.Sprintf("[BLOB %d bytes]", int(size))
			}
			if size, ok := v["size"].(int); ok {
				return fmt.Sprintf("[BLOB %d bytes]", size)
			}
			return "[BLOB]"
		}
		return fmt.Sprint(v)
	default:
		return fmt.Sprint(v)
	}
}

func WriteRows(w io.Writer, columns []string, rows []map[string]any, opts model.CSVFormatOptions) error {
	o := opts.Normalized()
	delimiter, size := utf8.DecodeRuneInString(o.Delimiter)
	if size == 0 {
		delimiter = ','
	}
	writer := csv.NewWriter(w)
	writer.Comma = delimiter
	if o.HasHeader && len(columns) > 0 {
		if err := writer.Write(columns); err != nil {
			return err
		}
	}
	for _, row := range rows {
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = formatCell(row[col], o.NullValue)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func LineEnding(opts model.CSVFormatOptions) string {
	if strings.ToLower(opts.LineEnding) == "lf" {
		return "\n"
	}
	return "\r\n"
}

func EncodeContent(body string, opts model.CSVFormatOptions) string {
	le := LineEnding(opts)
	if le == "\n" {
		return strings.ReplaceAll(body, "\r\n", "\n")
	}
	return strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
}
