// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"fmt"

	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

type RawCreate struct {
	Info              sqlembedgo.Info
	Replace           bool
	Suffix            string
	Return            *Var
	Arg               *Var
	Fields            []*Var
	SupportsReturning bool
}

func RawCreateFromIR(ir_cre *ir.Create, dialect sql.Dialect) *RawCreate {
	insert_sql := sql.InsertSQL(ir_cre, dialect)
	ins := &RawCreate{
		Info:              sqlembedgo.Embed("__", insert_sql),
		Replace:           ir_cre.Replace,
		Suffix:            convertSuffix(ir_cre.Suffix),
		SupportsReturning: dialect.Features().Returning,
	}
	if !ir_cre.NoReturn {
		ins.Return = VarFromModel(ir_cre.Model)
	}

	// the model struct is the only arg.
	ins.Arg = VarFromModel(ir_cre.Model)
	ins.Arg.Name = "raw_" + ins.Arg.Name

	// each field in the model is initialized from the raw model struct.
	for _, field := range ir_cre.Fields() {
		f := ModelFieldFromIR(field)
		v := VarFromField(field)
		if field.Nullable {
			v.InitVal = fmt.Sprintf("%s_%s_Raw(%s.%s).value()",
				ins.Arg.Type, f.Name, ins.Arg.Name, f.Name)
		} else {
			v.InitVal = fmt.Sprintf("%s_%s(%s.%s).value()",
				ins.Arg.Type, f.Name, ins.Arg.Name, f.Name)
		}
		v.Name = fmt.Sprintf("__%s_val", v.Name)
		ins.Fields = append(ins.Fields, v)
	}

	return ins
}

type Create struct {
	Info              sqlembedgo.Info
	Replace           bool
	Suffix            string
	Struct            *ModelStruct
	Return            *Var
	Args              []*Var
	RequiredArgs      []*Var
	StaticFields      []*Var
	SupportsReturning bool
	NeedsNow          bool
}

func CreateFromIR(ir_cre *ir.Create, dialect sql.Dialect) *Create {
	insert_sql := sql.InsertSQL(ir_cre, dialect)
	ins := &Create{
		Info:              sqlembedgo.Embed("__", insert_sql),
		Replace:           ir_cre.Replace,
		Suffix:            convertSuffix(ir_cre.Suffix),
		Struct:            ModelStructFromIR(ir_cre.Model),
		SupportsReturning: dialect.Features().Returning,
	}
	if !ir_cre.NoReturn {
		ins.Return = VarFromModel(ir_cre.Model)
	}

	for _, field := range ir_cre.InsertableStaticFields() {
		arg := ArgFromField(field)
		v := VarFromField(field)
		v.Name = fmt.Sprintf("__%s_val", v.Name)

		if field.InsertableRequired() {
			ins.Args = append(ins.Args, arg)
			ins.RequiredArgs = append(ins.RequiredArgs, arg)
			v.InitVal = fmt.Sprintf("%s.value()", arg.Name)
		} else if field.InsertableOptional() {
			v.InitVal = fmt.Sprintf("optional.%s.value()", ModelFieldFromIR(field).Name)
		} else { // must be autoinsert
			ins.NeedsNow = ins.NeedsNow || field.IsTime()
		}

		ins.StaticFields = append(ins.StaticFields, v)
	}

	if len(ir_cre.InsertableOptionalFields()) > 0 {
		ins.Args = append(ins.Args, &Var{
			Name: "optional",
			Type: ModelStructFromIR(ir_cre.Model).CreateStructName(),
		})
	}

	return ins
}
