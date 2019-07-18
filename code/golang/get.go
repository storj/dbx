// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"fmt"
	"strings"

	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

type Get struct {
	PartitionedArgs
	Info         sqlembedgo.Info
	Suffix       string
	Row          *Var
	FirstInfo    sqlembedgo.Info
	Continuation *Var
}

func GetFromIR(ir_read *ir.Read, dialect sql.Dialect) *Get {
	select_sql := sql.SelectSQL(ir_read, dialect)
	get := &Get{
		PartitionedArgs: PartitionedArgsFromWheres(ir_read.Where),
		Info:            sqlembedgo.Embed("__", select_sql),
		Suffix:          convertSuffix(ir_read.Suffix),
		Row:             GetRowFromIR(ir_read),
	}

	if ir_read.View == ir.Paged {
		// hack: remove the last where clause before generating the sql
		first_select := sql.SelectFromIRRead(ir_read, dialect)
		first_select.Where = first_select.Where[:len(first_select.Where)-1]
		first_sql := sql.SQLFromSelect(first_select)
		get.FirstInfo = sqlembedgo.Embed("__", first_sql)
		get.Continuation = MakeContinuationVar(ir_read)
	}

	return get
}

func GetRowFromIR(ir_read *ir.Read) *Var {
	if model := ir_read.SelectedModel(); model != nil {
		return VarFromModel(model)
	}
	return MakeResultVar(ir_read.Selectables)
}

func StructFromVar(v *Var) *Struct {
	s := &Struct{Name: v.Type}
	for _, field := range v.Fields {
		s.Fields = append(s.Fields, Field{
			Name: field.Name,
			Type: field.Type,
		})
	}
	return s
}

func MakeContinuationVar(ir_read *ir.Read) *Var {
	return StructVar(
		"__continuation",
		fmt.Sprintf("Paged_%s_Continuation", convertSuffix(ir_read.Suffix)),
		[]*Var{
			{Name: "_value", Type: VarFromField(ir_read.From.PagablePrimaryKey()).Type},
			{Name: "_set", Type: "bool"},
		})
}

func MakeResultVar(selectables []ir.Selectable) *Var {
	vars := VarsFromSelectables(selectables)

	// construct the aggregate struct name
	var parts []string
	for _, v := range vars {
		parts = append(parts, v.Name)
	}
	parts = append(parts, "Row")
	name := strings.Join(parts, "_")
	return StructVar("row", name, vars)
}

func ContinuationStructFromRead(ir_read *ir.Read) *Struct {
	if ir_read.View != ir.Paged {
		return nil
	}
	return StructFromVar(MakeContinuationVar(ir_read))
}

func ResultStructFromRead(ir_read *ir.Read) *Struct {
	// no result struct if there is just a single model selected
	if ir_read.SelectedModel() != nil {
		return nil
	}
	return StructFromVar(MakeResultVar(ir_read.Selectables))
}
