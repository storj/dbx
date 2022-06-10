// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"fmt"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
)

// Struct is used for generating go structures
type ModelStruct struct {
	Name   string
	Table  string
	Fields []*ModelField
}

func ModelStructFromIR(model *ir.Model) *ModelStruct {
	name := structName(model)

	return &ModelStruct{
		Name:   name,
		Table:  model.Table,
		Fields: ModelFieldsFromIR(model.Fields),
	}
}

func ModelStructsFromIR(models []*ir.Model) (out []*ModelStruct) {
	for _, model := range models {
		out = append(out, ModelStructFromIR(model))
	}
	return out
}

func (s *ModelStruct) UpdatableFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Updatable && !field.AutoUpdate {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) InsertableStaticFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Insertable && field.InsertableStatic() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) InsertableDynamicFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Insertable && field.InsertableDynamic() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) InsertableRequiredFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Insertable && field.InsertableRequired() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) InsertableOptionalFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Insertable && field.InsertableOptional() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) UpdateStructName() string {
	return s.Name + "_Update_Fields"
}

func (s *ModelStruct) CreateStructName() string {
	return s.Name + "_Create_Fields"
}

type ModelField struct {
	Name       string
	ModelName  string
	Type       string
	CtorValue  string
	MutateFn   string
	Column     string
	Nullable   bool
	Default    string
	Insertable bool
	AutoInsert bool
	Updatable  bool
	AutoUpdate bool
	TakeAddr   bool
}

func ModelFieldFromIR(field *ir.Field) *ModelField {
	return &ModelField{
		Name:       fieldName(field),
		ModelName:  structName(field.Model),
		Type:       valueType(field.Type, field.Nullable),
		CtorValue:  valueType(field.Type, false),
		MutateFn:   mutateFn(field.Type),
		Column:     field.Column,
		Nullable:   field.Nullable,
		Default:    field.Default,
		Insertable: true,
		AutoInsert: field.AutoInsert,
		Updatable:  field.Updatable,
		AutoUpdate: field.AutoUpdate,
		TakeAddr:   field.Nullable && field.Type != consts.BlobField && field.Type != consts.JsonField,
	}
}

func (f *ModelField) InsertableStatic() bool {
	return f.AutoInsert || f.Default == ""
}

func (f *ModelField) InsertableDynamic() bool {
	return !f.InsertableStatic()
}

func (f *ModelField) InsertableRequired() bool {
	return !f.Nullable && f.Default == "" && !f.AutoInsert
}

func (f *ModelField) InsertableOptional() bool {
	return !f.InsertableRequired() && !f.AutoInsert
}

func ModelFieldsFromIR(fields []*ir.Field) (out []*ModelField) {
	for _, field := range fields {
		out = append(out, ModelFieldFromIR(field))
	}
	return out
}

func (f *ModelField) StructName() string {
	return fmt.Sprintf("%s_%s_Field", f.ModelName, f.Name)
}

func (f *ModelField) ArgType() string {
	if f.Nullable {
		return "*" + f.StructName()
	}
	return f.StructName()
}

func valueType(t consts.FieldType, nullable bool) (value_type string) {
	switch t {
	case consts.TextField:
		value_type = "string"
	case consts.IntField, consts.SerialField:
		value_type = "int"
	case consts.UintField:
		value_type = "uint"
	case consts.Int64Field, consts.Serial64Field:
		value_type = "int64"
	case consts.Uint64Field:
		value_type = "uint64"
	case consts.BlobField:
		value_type = "[]byte"
	case consts.TimestampField:
		value_type = "time.Time"
	case consts.TimestampUTCField:
		value_type = "time.Time"
	case consts.BoolField:
		value_type = "bool"
	case consts.FloatField:
		value_type = "float32"
	case consts.Float64Field:
		value_type = "float64"
	case consts.DateField:
		value_type = "time.Time"
	case consts.JsonField:
		value_type = "[]byte"
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}

	if nullable && t != consts.BlobField && t != consts.JsonField {
		return "*" + value_type
	}
	return value_type
}

func zeroVal(t consts.FieldType, nullable bool) string {
	if nullable {
		return "nil"
	}
	switch t {
	case consts.TextField:
		return `""`
	case consts.IntField, consts.SerialField:
		return `int(0)`
	case consts.UintField:
		return `uint(0)`
	case consts.Int64Field, consts.Serial64Field:
		return `int64(0)`
	case consts.Uint64Field:
		return `uint64(0)`
	case consts.BlobField:
		return `nil`
	case consts.TimestampField:
		return `time.Time{}`
	case consts.TimestampUTCField:
		return `time.Time{}`
	case consts.BoolField:
		return `false`
	case consts.FloatField:
		return `float32(0)`
	case consts.Float64Field:
		return `float64(0)`
	case consts.DateField:
		return `time.Time{}`
	case consts.JsonField:
		return `nil`
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}
}

func initVal(t consts.FieldType, nullable bool) string {
	switch t {
	case consts.TextField:
		if nullable {
			return `(*string)(nil)`
		}
		return `""`
	case consts.IntField, consts.SerialField:
		if nullable {
			return `(*int)(nil)`
		}
		return `int(0)`
	case consts.UintField:
		if nullable {
			return `(*uint)(nil)`
		}
		return `uint(0)`
	case consts.Int64Field, consts.Serial64Field:
		if nullable {
			return `(*int64)(nil)`
		}
		return `int64(0)`
	case consts.Uint64Field:
		if nullable {
			return `(*uint64)(nil)`
		}
		return `uint64(0)`
	case consts.BlobField:
		if nullable {
			return `[]byte(nil)`
		}
		return `nil`
	case consts.TimestampField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `__now`
	case consts.TimestampUTCField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `__now.UTC()`
	case consts.BoolField:
		if nullable {
			return `(*bool)(nil)`
		}
		return `false`
	case consts.FloatField:
		if nullable {
			return `(*float32)(nil)`
		}
		return `float32(0)`
	case consts.Float64Field:
		if nullable {
			return `(*float64)(nil)`
		}
		return `float64(0)`
	case consts.DateField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `toDate(__now)`
	case consts.JsonField:
		if nullable {
			return `[]byte(nil)`
		}
		return `nil`
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}
}

func mutateFn(field_type consts.FieldType) string {
	switch field_type {
	case consts.TimestampUTCField:
		return "toUTC"
	case consts.DateField:
		return "toDate"
	default:
		return ""
	}
}
