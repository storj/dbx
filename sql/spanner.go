// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"
	"slices"
	"strings"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

type spanner struct {
	DefaultDialect
}

func (s *spanner) Name() string {
	return "spanner"
}

func (s *spanner) Features() Features {
	return Features{
		Returning:           true,
		PositionalArguments: true,
		NoLimitToken:        "ALL",
		ReplaceStyle:        ReplaceStyle_OnConflictUpdate,
	}
}

func (s *spanner) RowId() string {
	return ""
}

func (s *spanner) ColumnType(field *ir.Field) string {
	switch field.Type {
	case consts.SerialField, consts.Serial64Field:
		return "INT64_SEQ"
	case consts.IntField, consts.Int64Field,
		consts.UintField, consts.Uint64Field:
		return "INT64"
	case consts.FloatField, consts.Float64Field:
		return "REAL"
	case consts.TextField:
		return "STRING"
	case consts.JsonField:
		return "JSON"
	case consts.BoolField:
		return "BOOL"
	case consts.TimestampField, consts.TimestampUTCField:
		return "TIMESTAMP"
	case consts.BlobField:
		return "BYTES"
	case consts.DateField:
		return "TIMESTAMP"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (s *spanner) Rebind(sql string) string {
	return sql
}

func (s *spanner) EscapeString(e string) string {
	return e
}

func (s *spanner) BoolLit(v bool) string {
	return fmt.Sprintf("%v", v)
}

func (s *spanner) DefaultLit(v string) string {
	return v
}

func (s *spanner) ReturningLit() string {
	return "THEN RETURN"
}

func Spanner() Dialect {
	return &spanner{}
}

func (s *spanner) CreateSchema(schema *Schema) []sqlgen.SQL {
	var stmts []sqlgen.SQL

	for _, table := range schema.Tables {
		var dirs []sqlgen.SQL

		for _, column := range table.Columns {
			sizeSpec := ""
			if column.Type == "STRING" || column.Type == "BYTES" {
				sizeSpec = "(MAX)"
			}
			dir := Build(Lf("%s %s%s", column.Name, strings.Split(column.Type, "_")[0], sizeSpec))
			if column.NotNull {
				dir.Add(L("NOT NULL"))
			}
			if column.Default != "" {
				dir.Add(Lf("DEFAULT (%s)", column.Default))
			}
			if column.Type == "INT64_SEQ" {
				seqName := fmt.Sprintf("%s_%s", table.Name, column.Name)
				stmts = append(stmts, Build(Lf("CREATE SEQUENCE %s OPTIONS (sequence_kind='bit_reversed_positive')", seqName)).SQL())
				dir.Add(Lf("DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE %s))", seqName))
			}
			if ref := column.Reference; ref != nil {
				dir.Add(Lf("REFERENCES %s( %s )", ref.Table, ref.Column))
				if ref.OnDelete != "" {
					dir.Add(Lf("ON DELETE %s", ref.OnDelete))
				}
				if ref.OnUpdate != "" {
					dir.Add(Lf("ON UPDATE %s", ref.OnUpdate))
				}
			}
			dirs = append(dirs, dir.SQL())
		}

		for _, unique := range table.Unique {
			dir := Build(L("UNIQUE ("))
			dir.Add(J(", ", Strings(unique)...))
			dir.Add(L(")"))
			dirs = append(dirs, dir.SQL())
		}

		directives := J(",\n\t", dirs...)

		var pkeyDir sqlgen.SQL
		if pkey := table.PrimaryKey; len(pkey) > 0 {
			dir := Build(L("PRIMARY KEY ("))
			dir.Add(J(", ", Strings(pkey)...))
			dir.Add(L(")"))
			pkeyDir = dir.SQL()
		}

		stmt := J("",
			Lf("CREATE TABLE %s (\n\t", table.Name),
			directives,
			Lf("\n) "),
			pkeyDir,
		)

		stmts = append(stmts, sqlcompile.Compile(stmt))
	}

	for _, index := range schema.Indexes {
		stmt := Build(L("CREATE"))
		if index.Unique {
			stmt.Add(L("UNIQUE"))
		}
		stmt.Add(Lf("INDEX %s ON %s (", index.Name, index.Table))
		stmt.Add(J(", ", Strings(index.Columns)...), L(")"))
		if len(index.Storing) > 0 {
			stmt.Add(L("STORING ("), J(", ", Strings(index.Storing)...), L(")"))
		}
		if len(index.Where) > 0 {
			stmt.Add(L("WHERE"), J(" AND ", index.Where...))
		}
		stmt.Add(L(";"))

		stmts = append(stmts, sqlcompile.Compile(stmt.SQL()))
	}

	return stmts
}

// DropSchema implements Dialect.
func (s *spanner) DropSchema(schema *Schema) (res []sqlgen.SQL) {
	var stmts []sqlgen.SQL

	tables := schema.Tables
	slices.Reverse(tables)
	for _, table := range tables {
		for _, pk := range table.PrimaryKey {
			stmts = append(stmts, sqlcompile.Compile(Lf("ALTER TABLE  %s ALTER %s SET DEFAULT (null)", table.Name, pk)))
			stmts = append(stmts, sqlcompile.Compile(Lf("DROP SEQUENCE IF EXISTS %s_%s", table.Name, pk)))
		}
		stmts = append(stmts, sqlcompile.Compile(Lf("DROP TABLE IF EXISTS %s", table.Name)))
	}

	return stmts
}
