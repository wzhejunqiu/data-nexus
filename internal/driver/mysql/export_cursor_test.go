package mysql

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func expectCompositePKSchemaWithLookups(mock sqlmock.Sqlmock, lookups int) {
	for range lookups {
		mock.ExpectQuery(`SELECT TABLE_TYPE FROM information_schema.TABLES`).
			WithArgs("testdb", "post_tags").
			WillReturnRows(sqlmock.NewRows([]string{"TABLE_TYPE"}).AddRow("BASE TABLE"))
	}
	mock.ExpectQuery(`SELECT COLUMN_NAME, DATA_TYPE`).
		WithArgs("testdb", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "DATA_TYPE", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT", "ORDINAL_POSITION",
		}).AddRow("post_id", "int", "int(11)", "NO", "PRI", nil, 1).
			AddRow("tag_id", "int", "int(11)", "NO", "PRI", nil, 2))
	mock.ExpectQuery(`SELECT INDEX_NAME, NON_UNIQUE`).
		WithArgs("testdb", "post_tags").
		WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "SEQ_IN_INDEX", "COLUMN_NAME"}).
			AddRow("PRIMARY", 0, 1, "post_id").
			AddRow("PRIMARY", 0, 2, "tag_id"))
}

func TestBuildKeysetQueryCompositePK(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectCompositePKSchemaWithLookups(mock, 2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `post_tags` LIMIT 0")).
		WillReturnRows(sqlmock.NewRows([]string{"post_id", "tag_id"}))

	cursor, err := drv.OpenTableExport(context.Background(), "post_tags", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `post_tags` ORDER BY `post_id` ASC, `tag_id` ASC LIMIT ?")).
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

	mock.ExpectQuery("SELECT \\* FROM `post_tags` WHERE \\(`post_id` > \\?\\) OR \\(`post_id` = \\? AND `tag_id` > \\?\\) ORDER BY `post_id` ASC, `tag_id` ASC LIMIT \\?").
		WithArgs(int64(1), int64(1), int64(20), 2).
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
