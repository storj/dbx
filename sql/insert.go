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
	Table     string
	Columns   []string
	Returning []string
}

func InsertFromIRCreate(ir_cre *ir.Create, dialect Dialect) *Insert {
	ins := &Insert{
		Table: ir_cre.Model.Table,
	}
	if dialect.Features().Returning && !ir_cre.NoReturn {
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
	stmt := Build(Lf("INSERT INTO %s", insert.Table))

	if cols := insert.Columns; len(cols) > 0 {
		stmt.Add(L("("), J(", ", Strings(cols)...), L(")"))
		stmt.Add(L("VALUES ("), J(", ", Placeholders(len(cols))...), L(")"))
	} else {
		stmt.Add(L("DEFAULT VALUES"))
	}

	if rets := insert.Returning; len(rets) > 0 {
		stmt.Add(L("RETURNING"), J(", ", Strings(rets)...))
	}

	return sqlcompile.Compile(stmt.SQL())
}
