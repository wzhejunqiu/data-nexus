package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type Driver struct {
	db       *sql.DB
	readOnly bool
	schema   string
}

func New() *Driver { return &Driver{} }

func (d *Driver) Type() model.DriverType { return model.DriverTypePostgres }

func (d *Driver) ReadOnly() bool { return d.readOnly }

func (d *Driver) Connect(ctx context.Context, cfg model.DriverConfig) error {
	if cfg.Postgres == nil {
		return model.ErrInvalidRequest("postgres config required")
	}
	if d.db != nil {
		_ = d.Close()
	}
	pg := cfg.Postgres
	d.readOnly = pg.ReadOnly
	d.schema = pg.NormalizedSchema()
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(pg.User, pg.Password),
		Host:   fmt.Sprintf("%s:%d", pg.Host, pg.NormalizedPort()),
		Path:   "/" + pg.Database,
	}
	q := u.Query()
	q.Set("sslmode", pg.NormalizedSSLMode())
	if pg.ReadOnly {
		q.Set("default_transaction_read_only", "on")
	}
	u.RawQuery = q.Encode()
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		return model.ErrConnectionFailed(err.Error())
	}
	db.SetMaxOpenConns(10)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return model.ErrConnectionFailed(err.Error())
	}
	d.db = db
	return nil
}

func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	err := d.db.Close()
	d.db = nil
	return err
}

func (d *Driver) Ping(ctx context.Context) error {
	if d.db == nil {
		return model.ErrConnectionNotFound("")
	}
	return d.db.PingContext(ctx)
}

func (d *Driver) unsupportedSQLiteFeature() error {
	return model.ErrInvalidRequest("not supported for postgres")
}

func (d *Driver) Attach(context.Context, string, string) error { return d.unsupportedSQLiteFeature() }
func (d *Driver) Detach(context.Context, string) error         { return d.unsupportedSQLiteFeature() }
func (d *Driver) ListAttached(context.Context) ([]model.AttachedDatabase, error) {
	return nil, d.unsupportedSQLiteFeature()
}
func (d *Driver) DetectFTSTable(context.Context, string) (*model.FTSInfo, error) {
	return &model.FTSInfo{Enabled: false}, nil
}

func (d *Driver) schemaRef() string { return d.schema }

func tableFromRef(schema, table string) string {
	return fmt.Sprintf("%s.%s", quoteIdent(schema), quoteIdent(table))
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (d *Driver) OpenTableExport(ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	return newExportCursor(d, ctx, tableName, opts)
}
