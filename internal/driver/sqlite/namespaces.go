package sqlite

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (d *Driver) ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error) {
	items := []model.NamespaceInfo{{
		Name: "main",
		Kind: model.NamespaceKindDatabase,
	}}
	attached, err := d.ListAttached(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range attached {
		fp := a.FilePath
		items = append(items, model.NamespaceInfo{
			Name:     a.Alias,
			Kind:     model.NamespaceKindAttach,
			FilePath: &fp,
		})
	}
	return items, nil
}

func (d *Driver) ListSchemas(context.Context, string) ([]model.SchemaInfo, error) {
	return nil, nil
}

func sqliteNamespaceFromOpts(opts model.ListTablesOptions) string {
	if opts.Database != "" {
		return opts.Database
	}
	return opts.Schema
}
