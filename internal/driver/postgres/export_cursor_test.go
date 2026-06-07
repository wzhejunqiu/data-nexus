package postgres

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func expectExportTableLookup(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT table_type`).
		WithArgs("public", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{"table_type"}).AddRow("BASE TABLE"))
}

func TestExportCursorRollbackOnCreateFailure(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectExportTableLookup(mock)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags" LIMIT 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}))
	mock.ExpectExec(`BEGIN READ ONLY`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags"`)).
		WillReturnError(fmt.Errorf("select failed"))
	mock.ExpectExec(`ROLLBACK`).WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := drv.OpenTableExport(context.Background(), "post_tags", model.TableExportOptions{BatchSize: 2})
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamingExportCursor(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectExportTableLookup(mock)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags" LIMIT 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}))
	mock.ExpectExec(`BEGIN READ ONLY`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags"`)).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}).
			AddRow(int64(1), int64(10)).
			AddRow(int64(1), int64(20)).
			AddRow(int64(2), int64(5)))

	cursor, err := drv.OpenTableExport(context.Background(), "post_tags", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}

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
	mock.ExpectExec(`ROLLBACK`).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := cursor.Close(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
