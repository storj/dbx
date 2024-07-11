// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

type Update struct {
	Args       []ConditionArg
	Info       sqlembedgo.Info
	Suffix     string
	Struct     *ModelStruct
	Return     *Var
	AutoFields []*Var
	NeedsNow   bool
}

func UpdateFromIR(ir_upd *ir.Update, dialect sql.Dialect) *Update {
	update_sql := sql.UpdateSQL(ir_upd, dialect)
	upd := &Update{
		Args:   ConditionArgsFromWheres(ir_upd.Where),
		Info:   sqlembedgo.Embed("__", update_sql),
		Suffix: convertSuffix(ir_upd.Suffix),
		Struct: ModelStructFromIR(ir_upd.Model),
	}
	if !ir_upd.NoReturn {
		upd.Return = VarFromModel(ir_upd.Model)
	}

	for _, field := range ir_upd.AutoUpdatableFields() {
		upd.NeedsNow = upd.NeedsNow || field.IsTime()
		upd.AutoFields = append(upd.AutoFields, VarFromField(field))
	}

	return upd
}
