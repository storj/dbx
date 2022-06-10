// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"
	"strings"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
)

type sqlite3 struct {
}

func SQLite3() Dialect {
	return &sqlite3{}
}

func (s *sqlite3) Name() string {
	return "sqlite3"
}

func (s *sqlite3) Features() Features {
	return Features{
		Returning:    false,
		NoLimitToken: "-1",
		ReplaceStyle: ReplaceStyle_Replace,
	}
}

func (s *sqlite3) RowId() string {
	return "_rowid_"
}

func (s *sqlite3) ColumnType(field *ir.Field) string {
	switch field.Type {
	case consts.SerialField, consts.Serial64Field,
		consts.IntField, consts.Int64Field,
		consts.UintField, consts.Uint64Field:
		return "INTEGER"
	case consts.FloatField, consts.Float64Field:
		return "REAL"
	case consts.TextField, consts.JsonField:
		return "TEXT"
	case consts.BoolField:
		return "INTEGER"
	case consts.TimestampField, consts.TimestampUTCField:
		return "TIMESTAMP"
	case consts.BlobField:
		return "BLOB"
	case consts.DateField:
		return "DATE"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (s *sqlite3) Rebind(sql string) string {
	return sql
}

var sqlite3Replacer = strings.NewReplacer(
	`'`, `''`,
)

func (p *sqlite3) EscapeString(s string) string {
	return sqlite3Replacer.Replace(s)
}

func (p *sqlite3) BoolLit(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func (p *sqlite3) DefaultLit(v string) string {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return `'` + p.EscapeString(v[1:len(v)-1]) + `'`
	}
	return v
}
