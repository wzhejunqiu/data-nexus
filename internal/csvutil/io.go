package csvutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func readRawFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return nil, model.ErrInvalidPath("invalid path")
	}
	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, model.ErrInvalidPath(err.Error())
	}
	return data, nil
}
