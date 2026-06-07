package mysql

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func expectExportTableLookup(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT TABLE_TYPE FROM information_schema.TABLES`).
		WithArgs("testdb", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_TYPE"}).AddRow("BASE TABLE"))
}

func TestStreamingExportCursor(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectExportTableLookup(mock)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `post_tags` LIMIT 0")).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `post_tags`")).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}).
			AddRow(int64(1), int64(10)).
			AddRow(int64(1), int64(20)).
			AddRow(int64(2), int64(5)))

	cursor, err := drv.OpenTableExport(context.Background(), "post_tags", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 2 || !batch.HasMore {
		t.Fatalf("unexpected first batch: %+v", batch)
	}

	batch, err = cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 1 || batch.HasMore {
		t.Fatalf("unexpected second batch: %+v", batch)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
