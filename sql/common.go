// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import "storj.io/dbx/ir"

type Features struct {
	// Supports the RETURNING syntax on INSERT/UPDATE
	Returning bool

	// Supports positional argument placeholders
	PositionalArguments bool

	// Token used with LIMIT to mean "no limit"
	NoLimitToken string

	// What style the database uses to handle replacement creates
	ReplaceStyle ReplaceStyle

	// Supports the STORING feature of indexes
	Storing bool
}

type ReplaceStyle int

const (
	ReplaceStyle_Replace ReplaceStyle = iota
	ReplaceStyle_OnConflictUpdate
	ReplaceStyle_Upsert
)

type Dialect interface {
	Name() string
	Features() Features
	RowId() string
	ColumnType(field *ir.Field) string
	// Scanner give the opportunity to wrap a Go type in a database/sql.Scanner.
	// This is only called on destination of types not supported by database/sql.Rows.Scan().
	// See https://pkg.go.dev/database/sql#Rows.Scan
	Scanner(dest interface{}) interface{}
	Rebind(sql string) string
	EscapeString(s string) string
	BoolLit(v bool) string
	DefaultLit(v string) string
}
