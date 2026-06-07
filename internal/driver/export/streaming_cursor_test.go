package export_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestStreamingCursorBatchesUntilEOF(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(int64(1), "a").
		AddRow(int64(2), "b").
		AddRow(int64(3), "c")
	mock.ExpectQuery(`SELECT \* FROM t`).WillReturnRows(rows)

	columns := []model.ColumnMeta{{Name: "id"}, {Name: "name"}}
	cursor, err := export.NewStreamingCursor(context.Background(), db, "SELECT * FROM t", columns, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 2 || !batch.HasMore {
		t.Fatalf("first batch: %+v", batch)
	}

	batch, err = cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 1 || batch.HasMore {
		t.Fatalf("second batch: %+v", batch)
	}

	batch, err = cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 0 || batch.HasMore {
		t.Fatalf("third batch: %+v", batch)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamingCursorExactBatchMultiple(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(1)).
		AddRow(int64(2))
	mock.ExpectQuery(`SELECT \* FROM t`).WillReturnRows(rows)

	cursor, err := export.NewStreamingCursor(
		context.Background(),
		db,
		"SELECT * FROM t",
		[]model.ColumnMeta{{Name: "id"}},
		2,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 2 || !batch.HasMore {
		t.Fatalf("expected first full batch with hasMore, got %+v", batch)
	}

	batch, err = cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 0 || batch.HasMore {
		t.Fatalf("expected empty terminal batch, got %+v", batch)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamingCursorRespectsContextCancel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT \* FROM t`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	cursor, err := export.NewStreamingCursor(
		context.Background(),
		db,
		"SELECT * FROM t",
		[]model.ColumnMeta{{Name: "id"}},
		10,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := cursor.NextBatch(ctx); err == nil {
		t.Fatal("expected context error")
	}
}
