package model_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestTableExportOptionsNormalizedBatchSize(t *testing.T) {
	if (model.TableExportOptions{}).NormalizedBatchSize() != 1000 {
		t.Fatal("expected default batch size 1000")
	}
	if (model.TableExportOptions{BatchSize: -1}).NormalizedBatchSize() != 1000 {
		t.Fatal("expected default for non-positive batch size")
	}
	if (model.TableExportOptions{BatchSize: 500}).NormalizedBatchSize() != 500 {
		t.Fatal("expected explicit batch size")
	}
}
