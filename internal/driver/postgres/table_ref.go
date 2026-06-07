package postgres

import (
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

type tableRef struct {
	Schema    string
	BareName  string
	FromRef   string
	Qualified string
}

func (d *Driver) parseTableRef(tableName string) (tableRef, error) {
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schema, bare := parts[0], parts[1]
		if !sqlutil.IsSafeQuotedIdentifier(schema) || !sqlutil.IsSafeQuotedIdentifier(bare) {
			return tableRef{}, model.ErrTableNotFound(tableName)
		}
		return tableRef{
			Schema:    schema,
			BareName:  bare,
			FromRef:   tableFromRef(schema, bare),
			Qualified: tableName,
		}, nil
	}
	if !sqlutil.IsSafeQuotedIdentifier(tableName) {
		return tableRef{}, model.ErrTableNotFound(tableName)
	}
	schema := d.schemaRef()
	return tableRef{
		Schema:    schema,
		BareName:  tableName,
		FromRef:   tableFromRef(schema, tableName),
		Qualified: tableName,
	}, nil
}

func schemaLabel(schema string) *string {
	s := schema
	return &s
}
