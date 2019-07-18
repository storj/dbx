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

func DeleteSQL(ir_del *ir.Delete, dialect Dialect) sqlgen.SQL {
	stmt := Build(Lf("DELETE FROM %s", ir_del.Model.Table))

	var wheres []sqlgen.SQL
	if len(ir_del.Joins) == 0 {
		wheres = WhereSQL(ir_del.Where, dialect)
	} else {
		pk_column := ir_del.Model.PrimaryKey[0].ColumnRef()
		sel := SelectSQL(&ir.Read{
			View:        ir.All,
			From:        ir_del.Model,
			Selectables: []ir.Selectable{ir_del.Model.PrimaryKey[0]},
			Joins:       ir_del.Joins,
			Where:       ir_del.Where,
		}, dialect)
		wheres = append(wheres, J("", L(pk_column), L(" IN ("), sel, L(")")))
	}

	if len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	return sqlcompile.Compile(stmt.SQL())

}
