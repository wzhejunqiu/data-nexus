package postgres

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT datname
		FROM pg_database
		WHERE datallowconn AND NOT datistemplate
		ORDER BY datname`)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var items []model.NamespaceInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		items = append(items, model.NamespaceInfo{
			Name: name,
			Kind: model.NamespaceKindDatabase,
		})
	}
	return items, rows.Err()
}

func (d *Driver) ListSchemas(ctx context.Context, database string) ([]model.SchemaInfo, error) {
	if database == "" {
		database = d.database
	}
	if database != d.database {
		return nil, nil
	}
	rows, err := d.db.QueryContext(ctx, `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE catalog_name = current_database()
		ORDER BY schema_name`)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var items []model.SchemaInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		items = append(items, model.SchemaInfo{Name: name})
	}
	return items, rows.Err()
}

func (d *Driver) resolveSchema(opts model.ListTablesOptions) string {
	if opts.Schema != "" {
		return opts.Schema
	}
	return d.schemaRef()
}
