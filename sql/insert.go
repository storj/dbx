// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func InsertSQL(ir_cre *ir.Create, dialect Dialect) sqlgen.SQL {
	return SQLFromInsert(InsertFromIRCreate(ir_cre, dialect))
}

type Insert struct {
	Table        string
	Columns      []string
	Returning    []string
	ReplaceStyle *ReplaceStyle
}

func InsertFromIRCreate(ir_cre *ir.Create, dialect Dialect) *Insert {
	features := dialect.Features()
	ins := &Insert{
		Table: ir_cre.Model.Table,
	}
	if ir_cre.Replace {
		ins.ReplaceStyle = &features.ReplaceStyle
	}
	if features.Returning && !ir_cre.NoReturn {
		ins.Returning = ir_cre.Model.SelectRefs()
	}
	for _, field := range ir_cre.Fields() {
		if field == ir_cre.Model.BasicPrimaryKey() && !ir_cre.Raw {
			continue
		}
		ins.Columns = append(ins.Columns, field.Column)
	}

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
		}
	}

	stmt := Build(Lf("%s INTO %s", verb, insert.Table))
	cols := insert.Columns

	if len(cols) > 0 {
		stmt.Add(L("("), J(", ", Strings(cols)...), L(")"))
		stmt.Add(L("VALUES ("), J(", ", Placeholders(len(cols))...), L(")"))
	} else {
		stmt.Add(L("DEFAULT VALUES"))
	}

	if insert.ReplaceStyle != nil && *insert.ReplaceStyle == ReplaceStyle_OnConflictUpdate {
		stmt.Add(L("ON CONFLICT"))
		if len(cols) > 0 {
			stmt.Add(L("DO UPDATE SET"))
			var excluded []sqlgen.SQL
			for _, col := range cols {
				excluded = append(excluded, Lf("%s = EXCLUDED.%s", col, col))
			}
			stmt.Add(J(", ", excluded...))
		} else {
			stmt.Add(L("DO NOTHING"))
		}
	}

	if rets := insert.Returning; len(rets) > 0 {
		stmt.Add(L("RETURNING"), J(", ", Strings(rets)...))
	}

	return sqlcompile.Compile(stmt.SQL())
}
