// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

type Delete struct {
	Args   []ConditionArg
	Info   sqlembedgo.Info
	Suffix string
	Result *Var
}

func DeleteFromIR(ir_del *ir.Delete, dialect sql.Dialect) *Delete {
	delete_sql := sql.DeleteSQL(ir_del, dialect)
	del := &Delete{
		Args:   ConditionArgsFromWheres(ir_del.Where),
		Info:   sqlembedgo.Embed("__", delete_sql),
		Suffix: convertSuffix(ir_del.Suffix),
	}

	if ir_del.Distinct() {
		del.Result = &Var{
			Name: "deleted",
			Type: "bool",
		}
	} else {
		del.Result = &Var{
			Name: "count",
			Type: "int64",
		}
	}

	return del
}
