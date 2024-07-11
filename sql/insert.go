// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func InsertSQL(ir_cre *ir.Create, dialect Dialect) sqlgen.SQL {
	return SQLFromInsert(InsertFromIRCreate(ir_cre, dialect))
}

type Insert struct {
	Table      string
	PrimaryKey []string

	// StaticColumns are column names without defaults, requested to be inserted
	StaticColumns []string

	// DynamicColumns are column names with defaults, requested to be inserted
	DynamicColumns []string

	// AllDefaults are all the available non-ref columns (even if they are not requested as insert parameters).
	AllDefaults []string

	Returning    []string
	ReplaceStyle *ReplaceStyle
	// SupportDefaultValues should be true, if `INSERT INTO .... DEFAULT VALUES` is supported
	SupportDefaultValues bool
	ReturningLit         string
}

func InsertFromIRCreate(ir_cre *ir.Create, dialect Dialect) *Insert {
	features := dialect.Features()
	ins := &Insert{
		Table: ir_cre.Model.Table,
	}
	for _, field := range ir_cre.Model.PrimaryKey {
		ins.PrimaryKey = append(ins.PrimaryKey, field.Column)
	}
	if ir_cre.Replace {
		ins.ReplaceStyle = &features.ReplaceStyle
	}
	if !ir_cre.NoReturn {
		ins.Returning = ir_cre.Model.SelectRefs()
	}
	for _, field := range ir_cre.InsertableStaticFields() {
		ins.StaticColumns = append(ins.StaticColumns, field.Column)
	}
	for _, field := range ir_cre.InsertableDynamicFields() {
		ins.DynamicColumns = append(ins.DynamicColumns, field.Column)
	}
	for _, field := range ir_cre.Model.Fields {
		if field.Default != "" || field.Type == consts.Serial64Field || field.Type == consts.SerialField {
			ins.AllDefaults = append(ins.AllDefaults, field.Column)
		}
	}
	ins.SupportDefaultValues = dialect.Features().DefaultValues
	ins.ReturningLit = dialect.ReturningLit()
	return ins
}

func SQLFromInsert(insert *Insert) sqlgen.SQL {
	verb := "INSERT"
	if insert.ReplaceStyle != nil {
		switch *insert.ReplaceStyle {
		case ReplaceStyle_OnConflictUpdate:
		case ReplaceStyle_Replace:
			verb = "REPLACE"
		case ReplaceStyle_Upsert:
			verb = "UPSERT"
		case ReplaceStyle_Upsert_Spanner:
			verb = "INSERT OR UPDATE"

		}
	}

	stmt := Build(Lf("%s INTO %s", verb, insert.Table))

	switch {
	case len(insert.StaticColumns)+len(insert.DynamicColumns) == 0:
		if insert.SupportDefaultValues {
			stmt.Add(L("DEFAULT VALUES"))
		} else {
			stmt.Add(L("("), J(", ", Strings(insert.AllDefaults)...), L(")"))
			stmt.Add(L("VALUES ("))
			for i := 0; i < len(insert.AllDefaults); i++ {
				if i > 0 {
					stmt.Add(L(","))
				}
				stmt.Add(L("DEFAULT"))
			}
			stmt.Add(L(")"))
		}
	case len(insert.DynamicColumns) == 0:
		stmt.Add(L("("), J(", ", Strings(insert.StaticColumns)...), L(")"))
		stmt.Add(L("VALUES"))
		stmt.Add(L("("), J(", ", Placeholders(len(insert.StaticColumns))...), L(")"))

	default:
		columns := Hole("columns")
		placeholders := Hole("placeholders")

		if len(insert.StaticColumns) > 0 {
			columns.SQL = J(", ", Strings(insert.StaticColumns)...)
			placeholders.SQL = J(", ", Placeholders(len(insert.StaticColumns))...)
		}

		clause := Hole("clause")
		clause.SQL = J("", L("("), columns, L(") VALUES ("), placeholders, L(")"))
		stmt.Add(clause)
	}

	if insert.ReplaceStyle != nil && *insert.ReplaceStyle == ReplaceStyle_OnConflictUpdate {
		stmt.Add(L("ON CONFLICT ("), J(", ", Strings(insert.PrimaryKey)...), L(")"))
		if len(insert.StaticColumns)+len(insert.DynamicColumns) > 0 {
			stmt.Add(L("DO UPDATE SET"))
			var excluded []sqlgen.SQL
			for _, col := range insert.StaticColumns {
				excluded = append(excluded, Lf("%s = EXCLUDED.%s", col, col))
			}
			for _, col := range insert.DynamicColumns {
				excluded = append(excluded, Lf("%s = EXCLUDED.%s", col, col))
			}
			stmt.Add(J(", ", excluded...))
		} else if insert.ReplaceStyle != nil && *insert.ReplaceStyle == ReplaceStyle_Upsert_Spanner {
		} else {
			stmt.Add(L("DO NOTHING"))
		}
	}

	if rets := insert.Returning; len(rets) > 0 {
		retLit := insert.ReturningLit
		if retLit == "" {
			retLit = "RETURNING"
		}
		stmt.Add(L(retLit), J(", ", Strings(rets)...))
	}

	return sqlcompile.Compile(stmt.SQL())
}
