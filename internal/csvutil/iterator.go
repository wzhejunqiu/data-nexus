package csvutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"encoding/csv"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

// Iterator streams CSV data rows without loading the entire file into memory.
type Iterator struct {
	file        *os.File
	reader      *csv.Reader
	format      model.CSVFormatOptions
	headers     []string
	pendingRow  []string
	hasPending  bool
}

func OpenIterator(path string, format model.CSVFormatOptions) (*Iterator, error) {
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return nil, model.ErrInvalidPath("invalid path")
	}
	f, err := os.Open(clean)
	if err != nil {
		return nil, model.ErrInvalidPath(err.Error())
	}

	reader, err := NewReader(f, format)
	if err != nil {
		_ = f.Close()
		return nil, model.ErrInvalidRequest(err.Error())
	}

	it := &Iterator{file: f, reader: reader, format: format.Normalized()}
	if err := it.readHeaders(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return it, nil
}

func (it *Iterator) Headers() []string {
	return append([]string{}, it.headers...)
}

func (it *Iterator) readHeaders() error {
	record, err := it.reader.Read()
	if err == io.EOF {
		it.headers = []string{}
		return nil
	}
	if err != nil {
		return model.ErrInvalidRequest("csv parse error: " + err.Error())
	}
	if it.format.HasHeader {
		it.headers = append([]string{}, record...)
		return nil
	}
	it.headers = make([]string, len(record))
	for i := range record {
		it.headers[i] = fmt.Sprintf("column_%d", i+1)
	}
	it.pendingRow = append([]string{}, record...)
	it.hasPending = true
	return nil
}

// ReadBatch returns up to maxRows data records. io.EOF when finished.
func (it *Iterator) ReadBatch(maxRows int) ([][]string, error) {
	if maxRows <= 0 {
		maxRows = 1
	}
	records := make([][]string, 0, maxRows)
	if it.hasPending {
		records = append(records, append([]string{}, it.pendingRow...))
		it.hasPending = false
		maxRows--
	}
	for maxRows > 0 {
		record, err := it.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return records, model.ErrInvalidRequest("csv parse error: " + err.Error())
		}
		records = append(records, append([]string{}, record...))
		maxRows--
	}
	if len(records) == 0 {
		return nil, io.EOF
	}
	return records, nil
}

func (it *Iterator) Close() error {
	if it.file == nil {
		return nil
	}
	err := it.file.Close()
	it.file = nil
	return err
}
