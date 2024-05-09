// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"slices"

	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

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
	Rebind(sql string) string
	EscapeString(s string) string
	BoolLit(v bool) string
	DefaultLit(v string) string
	// ReturningLit returns with literal to be used for returning values from an insert.
	ReturningLit() string

	// CreateSchema generates SQL from abstract schema to create all the required tables/indexes.
	CreateSchema(schema *Schema) []sqlgen.SQL

	//DropSchema drops all the resources from the schema. Useful for testing.
	DropSchema(schema *Schema) []sqlgen.SQL
}

// DefaultDialect can be used to make forward compatible Dialects with default implementations for new methods.
type DefaultDialect struct {
}

// CreateSchema implements Dialect.
func (d DefaultDialect) CreateSchema(schema *Schema) (res []sqlgen.SQL) {
	res = append(res, SQLFromSchema(schema))
	return res
}

// DropSchema implements Dialect.
func (d DefaultDialect) DropSchema(schema *Schema) (res []sqlgen.SQL) {
	var stmts []sqlgen.SQL

	tables := schema.Tables
	slices.Reverse(tables)
	for _, table := range tables {
		stmts = append(stmts, sqlcompile.Compile(Lf("DROP TABLE IF EXISTS %s", table.Name)))
	}

	return stmts
}

func (d DefaultDialect) ReturningLit() string {
	return "RETURNING"
}
