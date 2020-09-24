// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"fmt"

	"storj.io/dbx/internal/inflect"
	"storj.io/dbx/ir"
)

func VarFromSelectable(selectable ir.Selectable, full_name bool) (v *Var) {
	switch obj := selectable.(type) {
	case *ir.Model:
		v = VarFromModel(obj)
		v.Name = inflect.Camelize(v.Name)
	case *ir.Field:
		v = VarFromField(obj)
		if full_name {
			v.Name = inflect.Camelize(obj.Model.Name) + "_" +
				inflect.Camelize(obj.Name)
		} else {
			v.Name = inflect.Camelize(v.Name)
		}
	default:
		panic(fmt.Sprintf("unhandled selectable type %T", obj))
	}
	return v
}

func VarsFromSelectables(selectables []ir.Selectable) (vars []*Var) {
	// we use a full name unless:
	// 1. it is a single model as the selectable.
	// 2. every selectable is a field with the same model.

	full_name := false
	field_model := (*ir.Model)(nil)

selectables:
	for _, selectable := range selectables {
		switch selectable := selectable.(type) {
		case *ir.Model:
			full_name = len(selectables) != 1

		case *ir.Field:
			if field_model == nil {
				field_model = selectable.Model
			}
			if selectable.Model != field_model {
				full_name = true
				break selectables
			}

		default:
			full_name = true
			break selectables
		}
	}

	for _, selectable := range selectables {
		v := VarFromSelectable(selectable, full_name)
		vars = append(vars, v)
	}

	return vars
}

func VarFromModel(model *ir.Model) *Var {
	fields := VarsFromFields(model.Fields)
	for _, field := range fields {
		field.Name = inflect.Camelize(field.Name)
	}
	return StructVar(model.Name, structName(model), fields)
}

func VarFromField(field *ir.Field) *Var {
	return &Var{
		Name:    field.Name,
		Type:    valueType(field.Type, field.Nullable),
		ZeroVal: zeroVal(field.Type, field.Nullable),
		InitVal: initVal(field.Type, field.Nullable),
	}
}

func VarsFromFields(fields []*ir.Field) (vars []*Var) {
	for _, field := range fields {
		vars = append(vars, VarFromField(field))
	}
	return vars
}

func ArgFromField(field *ir.Field) *Var {
	// we don't set ZeroVal or InitVal because these args should only be used
	// as incoming arguments to function calls.
	return &Var{
		Name: field.UnderRef(),
		Type: ModelFieldFromIR(field).StructName(),
	}
}

func StructVar(name string, typ string, vars []*Var) *Var {
	return &Var{
		Name:    name,
		Type:    typ,
		Fields:  vars,
		InitVal: fmt.Sprintf("&%s{}", typ),
		ZeroVal: fmt.Sprintf("(*%s)(nil)", typ),
	}
}

type Var struct {
	Name    string
	Type    string
	ZeroVal string
	InitVal string
	Fields  []*Var
}

func (v *Var) Copy() *Var {
	out := &Var{
		Name:    v.Name,
		Type:    v.Type,
		ZeroVal: v.ZeroVal,
		InitVal: v.InitVal,
	}
	for _, field := range v.Fields {
		out.Fields = append(out.Fields, field.Copy())
	}
	return out
}

func (v *Var) Value() string {
	return v.Name
}

func (v *Var) Arg() string {
	return v.Name
}

func (v *Var) Init() string {
	return fmt.Sprintf("%s = %s", v.Name, v.InitVal)
}

func (v *Var) InitNew() string {
	return fmt.Sprintf("%s := %s", v.Name, v.InitVal)
}

func (v *Var) Declare() string {
	return fmt.Sprintf("var %s %s", v.Name, v.Type)
}

func (v *Var) Zero() string {
	return v.ZeroVal
}

func (v *Var) AddrOf() string {
	return fmt.Sprintf("&%s", v.Name)
}

func (v *Var) Param() string {
	if v.IsStruct() {
		return fmt.Sprintf("%s *%s", v.Name, v.Type)
	}
	return fmt.Sprintf("%s %s", v.Name, v.Type)
}

func (v *Var) SliceOf() string {
	if v.IsStruct() {
		return fmt.Sprintf("[]*%s", v.Type)
	}
	return fmt.Sprintf("[]%s", v.Type)
}

func (v *Var) IsStruct() bool {
	return len(v.Fields) > 0
}

func (v *Var) Flatten() (flattened []*Var) {
	if len(v.Fields) == 0 {
		// return a copy
		copy := *v
		return append(flattened, &copy)
	}

	for _, field := range v.Fields {
		field_vars := field.Flatten()
		for _, field_var := range field_vars {
			field_var.Name = v.Name + "." + field_var.Name
		}
		flattened = append(flattened, field_vars...)
	}
	return flattened
}
