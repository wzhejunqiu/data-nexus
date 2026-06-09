package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/driver/remote"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type Driver struct {
	db       *sql.DB
	readOnly bool
	database string
}

func New() *Driver { return &Driver{} }

func (d *Driver) Type() model.DriverType { return model.DriverTypeMySQL }

func (d *Driver) ReadOnly() bool { return d.readOnly }

func buildMySQLDSN(my *model.MySQLConfig, password string) string {
	params := []string{
		"parseTime=true",
		"allowNativePasswords=true",
		"charset=" + my.NormalizedCharset(),
		"collation=" + my.NormalizedCollation(),
	}
	if my.TLS {
		if my.TLSSkipVerify {
			params = append(params, "tls=skip-verify")
		} else {
			params = append(params, "tls=true")
		}
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		my.User, password, my.Host, my.NormalizedPort(), my.NormalizedDatabase(), strings.Join(params, "&"))
}

func (d *Driver) Connect(ctx context.Context, cfg model.DriverConfig) error {
	if cfg.MySQL == nil {
		return model.ErrInvalidRequest("mysql config required")
	}
	if d.db != nil {
		_ = d.Close()
	}
	my := cfg.MySQL
	d.readOnly = my.ReadOnly
	d.database = my.NormalizedDatabase()

	dsn := buildMySQLDSN(my, my.Password)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return model.ErrConnectionFailed(err.Error())
	}
	db.SetMaxOpenConns(10)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return model.ErrConnectionFailed(err.Error())
	}
	if engine := strings.TrimSpace(my.DefaultStorageEngine); engine != "" {
		if _, err := db.ExecContext(ctx, "SET SESSION default_storage_engine = ?", engine); err != nil {
			_ = db.Close()
			return model.ErrConnectionFailed(err.Error())
		}
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
	return model.ErrInvalidRequest("not supported for mysql")
}

func (d *Driver) Attach(context.Context, string, string) error { return d.unsupportedSQLiteFeature() }
func (d *Driver) Detach(context.Context, string) error         { return d.unsupportedSQLiteFeature() }
func (d *Driver) ListAttached(context.Context) ([]model.AttachedDatabase, error) {
	return nil, d.unsupportedSQLiteFeature()
}
func (d *Driver) DetectFTSTable(context.Context, string) (*model.FTSInfo, error) {
	return &model.FTSInfo{Enabled: false}, nil
}

func (d *Driver) OpenTableExport(ctx context.Context, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	return newExportCursor(d, ctx, tableName, opts)
}

func quoteIdent(name string) string {
	return remote.QuoteBacktick(name)
}

func tableFromRef(database, table string) string {
	return quoteIdent(database) + "." + quoteIdent(table)
}
