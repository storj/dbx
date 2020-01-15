// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/consts"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformField(lookup *lookup, field_entry *fieldEntry) (err error) {
	field := field_entry.field
	ast_field := field_entry.ast

	field.Name = ast_field.Name.Value
	field.Column = ast_field.Column.Get()
	field.Nullable = ast_field.Nullable.Get()
	field.Updatable = ast_field.Updatable.Get()
	field.AutoInsert = ast_field.AutoInsert.Get()
	field.AutoUpdate = ast_field.AutoUpdate.Get()
	field.Length = ast_field.Length.Get()
	field.Default = ast_field.Default.Get()

	if field.AutoUpdate {
		field.Updatable = true
	}

	if ast_field.Relation != nil {
		related, err := lookup.FindField(ast_field.Relation)
		if err != nil {
			return err
		}
		relation_kind := ast_field.RelationKind.Value

		if relation_kind == consts.SetNull && !field.Nullable {
			return errutil.New(ast_field.Pos,
				"setnull relationships must be nullable")
		}

		field.Relation = &ir.Relation{
			Field: related,
			Kind:  relation_kind,
		}
		field.Type = related.Type.AsLink()
	} else {
		field.Type = ast_field.Type.Value
	}

	if ast_field.AutoUpdate != nil && !podFields[field.Type] {
		return errutil.New(ast_field.AutoInsert.Pos,
			"autoinsert must be on plain data type")
	}
	if ast_field.AutoUpdate != nil && !podFields[field.Type] {
		return errutil.New(ast_field.AutoUpdate.Pos,
			"autoupdate must be on plain data type")
	}
	if ast_field.Length != nil && field.Type != consts.TextField {
		return errutil.New(ast_field.Length.Pos,
			"length must be on a text field")
	}

	if field.Column == "" {
		field.Column = field.Name
	}

	return nil
}

var podFields = map[consts.FieldType]bool{
	consts.IntField:          true,
	consts.Int64Field:        true,
	consts.UintField:         true,
	consts.Uint64Field:       true,
	consts.BoolField:         true,
	consts.TextField:         true,
	consts.TimestampField:    true,
	consts.TimestampUTCField: true,
	consts.FloatField:        true,
	consts.Float64Field:      true,
	consts.DateField:         true,
}
