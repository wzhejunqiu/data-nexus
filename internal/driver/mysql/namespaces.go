package mysql

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error) {
	rows, err := d.db.QueryContext(ctx, "SHOW DATABASES")
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

func (d *Driver) ListSchemas(context.Context, string) ([]model.SchemaInfo, error) {
	return nil, nil
}

func (d *Driver) resolveDatabase(opts model.ListTablesOptions) string {
	if opts.Database != "" {
		return opts.Database
	}
	return d.database
}
