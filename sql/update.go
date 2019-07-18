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

func UpdateSQL(ir_upd *ir.Update, dialect Dialect) sqlgen.SQL {
	return SQLFromUpdate(UpdateFromIRUpdate(ir_upd, dialect))
}

type Update struct {
	Table     string
	Where     []sqlgen.SQL
	Returning []string
	In        sqlgen.SQL
}

func UpdateFromIRUpdate(ir_upd *ir.Update, dialect Dialect) *Update {
	var returning []string
	if dialect.Features().Returning && !ir_upd.NoReturn {
		returning = ir_upd.Model.SelectRefs()
	}

	if len(ir_upd.Joins) == 0 {
		return &Update{
			Table:     ir_upd.Model.Table,
			Where:     WhereSQL(ir_upd.Where, dialect),
			Returning: returning,
		}
	}

	pk_column := ir_upd.Model.PrimaryKey[0].Column
	sel := SQLFromSelect(&Select{
		From:   ir_upd.Model.Table,
		Fields: []string{pk_column},
		Joins:  JoinsFromIRJoins(ir_upd.Joins),
		Where:  WhereSQL(ir_upd.Where, dialect),
	})
	in := J("", L(pk_column), L(" IN ("), sel, L(")"))

	return &Update{
		Table:     ir_upd.Model.Table,
		Returning: returning,
		In:        in,
	}
}

func SQLFromUpdate(upd *Update) sqlgen.SQL {
	stmt := Build(Lf("UPDATE %s SET", upd.Table))

	stmt.Add(Hole("sets"))

	wheres := upd.Where
	if upd.In != nil {
		wheres = append(wheres, upd.In)
	}
	if len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	if len(upd.Returning) > 0 {
		stmt.Add(L("RETURNING"), J(", ", Strings(upd.Returning)...))
	}

	return sqlcompile.Compile(stmt.SQL())
}
