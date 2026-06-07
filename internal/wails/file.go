package wails

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type FileService struct {
	log *zap.Logger
}

func NewFileService(log *zap.Logger) *FileService {
	return &FileService{log: log}
}

func (s *FileService) WriteTextFile(path, content, encoding string) error {
	return callVoid(s.log, "FileService.WriteTextFile", func() error {
		if strings.TrimSpace(path) == "" {
			return model.ErrInvalidPath("path is required")
		}
		clean := filepath.Clean(path)
		if strings.Contains(clean, "..") {
			return model.ErrInvalidPath("path must not contain ..")
		}
		ext := strings.ToLower(filepath.Ext(clean))
		if ext != ".csv" && ext != ".tsv" && ext != ".txt" {
			return model.ErrInvalidPath("only .csv, .tsv, .txt files are allowed")
		}
		data := []byte(content)
		switch strings.ToLower(encoding) {
		case "", "utf-8":
			// no BOM
		case "utf-8-bom":
			data = append([]byte{0xEF, 0xBB, 0xBF}, data...)
		default:
			return model.ErrInvalidRequest("unsupported encoding: " + encoding)
		}
		if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			return model.ErrInternal(err.Error())
		}
		if err := os.WriteFile(clean, data, 0o644); err != nil {
			return model.ErrInternal(err.Error())
		}
		return nil
	})
}
