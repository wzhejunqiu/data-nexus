package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func (d *Driver) GetTableProfile(ctx context.Context, tableName string) (*model.TableProfile, error) {
	ref, err := d.parseTableRef(tableName)
	if err != nil {
		return nil, err
	}
	schema, err := d.GetTableSchema(ctx, ref.Qualified)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", ref.FromRef)).Scan(&total); err != nil {
		return nil, model.ErrSQL(err.Error())
	}

	profile := &model.TableProfile{
		TableName:   ref.Qualified,
		TotalRows:   &total,
		SampledRows: total,
		IsSampled:   false,
		Columns:     make([]model.ColumnProfile, 0, len(schema.Columns)),
	}

	if total == 0 {
		for _, col := range schema.Columns {
			profile.Columns = append(profile.Columns, model.ColumnProfile{Name: col.Name})
		}
		return profile, nil
	}

	fromClause := ref.FromRef
	if total > model.ProfileSampleLimit {
		profile.IsSampled = true
		profile.SampledRows = model.ProfileSampleLimit
		fromClause = fmt.Sprintf("(SELECT * FROM %s LIMIT %d)", ref.FromRef, model.ProfileSampleLimit)
	}

	for _, col := range schema.Columns {
		cp, err := d.profileColumn(ctx, fromClause, col)
		if err != nil {
			return nil, err
		}
		profile.Columns = append(profile.Columns, cp)
	}
	return profile, nil
}

func (d *Driver) profileColumn(ctx context.Context, fromClause string, col model.ColumnInfo) (model.ColumnProfile, error) {
	if !sqlutil.IsSafeQuotedIdentifier(col.Name) {
		return model.ColumnProfile{}, model.ErrInvalidRequest("invalid column name")
	}
	cp := model.ColumnProfile{Name: col.Name}
	colRef := quoteIdent(col.Name)

	var distinct int64
	q := fmt.Sprintf("SELECT COUNT(DISTINCT %s) FROM %s", colRef, fromClause)
	if err := d.db.QueryRowContext(ctx, q).Scan(&distinct); err != nil {
		return cp, model.ErrSQL(err.Error())
	}
	cp.DistinctCount = &distinct

	var nullPct float64
	q = fmt.Sprintf("SELECT 100.0 * SUM(CASE WHEN %s IS NULL THEN 1 ELSE 0 END) / COUNT(*) FROM %s", colRef, fromClause)
	if err := d.db.QueryRowContext(ctx, q).Scan(&nullPct); err != nil {
		return cp, model.ErrSQL(err.Error())
	}
	cp.NullPercent = &nullPct

	if isNumericOrDateType(col.DataType) {
		var minVal, maxVal sql.NullString
		q = fmt.Sprintf("SELECT MIN(%s)::text, MAX(%s)::text FROM %s WHERE %s IS NOT NULL", colRef, colRef, fromClause, colRef)
		if err := d.db.QueryRowContext(ctx, q).Scan(&minVal, &maxVal); err != nil {
			return cp, model.ErrSQL(err.Error())
		}
		if minVal.Valid {
			cp.MinValue = &minVal.String
		}
		if maxVal.Valid {
			cp.MaxValue = &maxVal.String
		}
	}

	if distinct > 0 && distinct <= model.LowCardinalityThreshold {
		cp.IsLowCardinality = true
		top, err := d.profileTopValues(ctx, fromClause, col.Name)
		if err != nil {
			return cp, err
		}
		cp.TopValues = top
	}
	return cp, nil
}

func (d *Driver) profileTopValues(ctx context.Context, fromClause, colName string) ([]model.ValueCount, error) {
	colRef := quoteIdent(colName)
	q := fmt.Sprintf(
		`SELECT %s::text, COUNT(*) AS cnt FROM %s WHERE %s IS NOT NULL GROUP BY %s ORDER BY cnt DESC LIMIT %d`,
		colRef, fromClause, colRef, colRef, model.ProfileTopValuesLimit,
	)
	rows, err := d.db.QueryContext(ctx, q)
	if err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var top []model.ValueCount
	for rows.Next() {
		var val string
		var cnt int64
		if err := rows.Scan(&val, &cnt); err != nil {
			return nil, model.ErrSQL(err.Error())
		}
		top = append(top, model.ValueCount{Value: val, Count: cnt})
	}
	return top, rows.Err()
}

func isNumericOrDateType(dataType string) bool {
	upper := strings.ToUpper(dataType)
	keywords := []string{
		"INT", "REAL", "FLOAT", "DOUBLE", "NUMERIC", "DECIMAL",
		"DATE", "TIME", "TIMESTAMP", "MONEY", "SERIAL",
	}
	for _, k := range keywords {
		if strings.Contains(upper, k) {
			return true
		}
	}
	return false
}
