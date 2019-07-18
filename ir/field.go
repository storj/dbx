// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import (
	"fmt"

	"storj.io/dbx/consts"
)

type Relation struct {
	Field *Field
	Kind  consts.RelationKind
}

type Field struct {
	Name       string
	Column     string
	Model      *Model
	Type       consts.FieldType
	Relation   *Relation
	Nullable   bool
	AutoInsert bool
	AutoUpdate bool
	Updatable  bool
	Length     int // Text only
}

func (f *Field) Insertable() bool {
	if f.Relation != nil {
		return true
	}
	return f.Type != consts.SerialField && f.Type != consts.Serial64Field
}

func (f *Field) Unique() bool {
	return f.Model.FieldUnique(f)
}

func (f *Field) IsInt() bool {
	switch f.Type {
	case consts.SerialField, consts.Serial64Field,
		consts.IntField, consts.Int64Field,
		consts.UintField, consts.Uint64Field:
		return true
	default:
		return false
	}
}

func (f *Field) IsTime() bool {
	switch f.Type {
	case consts.TimestampField, consts.TimestampUTCField, consts.DateField:
		return true
	default:
		return false
	}
}

func (f *Field) ColumnRef() string {
	return fmt.Sprintf("%s.%s", f.Model.Table, f.Column)
}

func (f *Field) ModelOf() *Model {
	return f.Model
}

func (f *Field) UnderRef() string {
	return fmt.Sprintf("%s_%s", f.Model.Name, f.Name)
}

func (f *Field) SelectRefs() (refs []string) {
	return []string{f.ColumnRef()}
}

func (f *Field) selectable() {}
