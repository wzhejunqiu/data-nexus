package wails

import (
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

func call[T any](log *zap.Logger, method string, fn func() (T, error)) (T, error) {
	start := time.Now()
	result, err := fn()
	fields := []zap.Field{
		zap.String("method", method),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	}
	if err != nil {
		if appErr, ok := err.(*model.AppError); ok {
			fields = append(fields, zap.String("error_code", appErr.Code))
		}
		log.Warn("service call failed", append(fields, zap.Error(err))...)
		return result, err
	}
	log.Debug("service call", fields...)
	return result, nil
}

func callVoid(log *zap.Logger, method string, fn func() error) error {
	_, err := call(log, method, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}
