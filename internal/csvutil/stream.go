package csvutil

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type crlfWriter struct {
	w io.Writer
}

func (c *crlfWriter) Write(p []byte) (int, error) {
	var buf []byte
	for _, b := range p {
		if b == '\n' {
			buf = append(buf, '\r', '\n')
		} else {
			buf = append(buf, b)
		}
	}
	if _, err := c.w.Write(buf); err != nil {
		return 0, err
	}
	return len(p), nil
}

// RowWriter streams CSV rows to w without buffering the entire table in memory.
type RowWriter struct {
	csv     *csv.Writer
	columns []string
}

func NewRowWriter(w io.Writer, columns []string, opts model.CSVFormatOptions) (*RowWriter, error) {
	o := opts.Normalized()
	delimiter, size := utf8.DecodeRuneInString(o.Delimiter)
	if size == 0 {
		delimiter = ','
	}
	writer := csv.NewWriter(w)
	writer.Comma = delimiter
	rw := &RowWriter{csv: writer, columns: columns}
	if o.HasHeader && len(columns) > 0 {
		if err := writer.Write(columns); err != nil {
			return nil, err
		}
	}
	return rw, nil
}

func (rw *RowWriter) WriteRows(rows []map[string]any) error {
	for _, row := range rows {
		if err := rw.writeRow(row); err != nil {
			return err
		}
	}
	return nil
}

func (rw *RowWriter) writeRow(row map[string]any) error {
	record := make([]string, len(rw.columns))
	for i, col := range rw.columns {
		record[i] = formatCell(row[col])
	}
	return rw.csv.Write(record)
}

func (rw *RowWriter) Close() error {
	rw.csv.Flush()
	return rw.csv.Error()
}

func OpenExportFile(path string, opts model.CSVFormatOptions) (*os.File, io.Writer, error) {
	clean, err := validateExportPath(path)
	if err != nil {
		return nil, nil, err
	}
	o := opts.Normalized()
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return nil, nil, model.ErrInternal(err.Error())
	}
	f, err := os.Create(clean)
	if err != nil {
		return nil, nil, model.ErrInternal(err.Error())
	}
	var w io.Writer = f
	switch strings.ToLower(o.Encoding) {
	case "", "utf-8":
	case "utf-8-bom":
		if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			_ = f.Close()
			_ = os.Remove(clean)
			return nil, nil, model.ErrInternal(err.Error())
		}
	default:
		_ = f.Close()
		_ = os.Remove(clean)
		return nil, nil, model.ErrInvalidRequest("unsupported encoding: " + o.Encoding)
	}
	if strings.ToLower(o.LineEnding) != "lf" {
		w = &crlfWriter{w: f}
	}
	return f, w, nil
}

func validateExportPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", model.ErrInvalidPath("path is required")
	}
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return "", model.ErrInvalidPath("path must not contain ..")
	}
	ext := strings.ToLower(filepath.Ext(clean))
	if ext != ".csv" && ext != ".tsv" && ext != ".txt" {
		return "", model.ErrInvalidPath("only .csv, .tsv, .txt files are allowed")
	}
	return clean, nil
}
