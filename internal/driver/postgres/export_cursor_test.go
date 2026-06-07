package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func expectCompositePKSchemaWithLookups(mock sqlmock.Sqlmock, lookups int) {
	for range lookups {
		mock.ExpectQuery(`SELECT table_type`).
			WithArgs("public", "post_tags").
			WillReturnRows(sqlmock.NewRows([]string{"table_type"}).AddRow("BASE TABLE"))
	}
	mock.ExpectQuery(`SELECT column_name, data_type`).
		WithArgs("public", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{
			"column_name", "data_type", "udt_name", "is_nullable", "column_default", "ordinal_position",
		}).AddRow("post_id", "integer", "int4", "NO", nil, 1).
			AddRow("tag_id", "integer", "int4", "NO", nil, 2))
	mock.ExpectQuery(`SELECT kcu.column_name`).
		WithArgs("public", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{"column_name"}).
			AddRow("post_id").
			AddRow("tag_id"))
	mock.ExpectQuery(`SELECT\s+irel.relname`).
		WithArgs("public", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{"index_name", "indisunique", "indisprimary", "column_name", "col_position"}).
			AddRow("post_tags_pkey", true, true, "post_id", 1).
			AddRow("post_tags_pkey", true, true, "tag_id", 2))
}

func TestBuildKeysetQueryCompositePK(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectCompositePKSchemaWithLookups(mock, 2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags" LIMIT 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}))

	cursor, err := drv.OpenTableExport(context.Background(), "post_tags", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."post_tags" ORDER BY "post_id" ASC, "tag_id" ASC LIMIT $1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}).
			AddRow(int64(1), int64(10)).
			AddRow(int64(1), int64(20)))

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 2 || !batch.HasMore {
		t.Fatalf("unexpected first batch: %+v", batch)
	}

	mock.ExpectQuery(`SELECT \* FROM "public"\."post_tags" WHERE \("post_id", "tag_id"\) > \(\$1, \$2\) ORDER BY "post_id" ASC, "tag_id" ASC LIMIT \$3`).
		WithArgs(int64(1), int64(20), 2).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}).
			AddRow(int64(2), int64(5)))

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
