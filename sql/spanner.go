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
		DefaultValues:       false,
		PositionalArguments: true,
		TupleComparsion:     false,
		NoLimitToken:        "ALL",
		ReplaceStyle:        ReplaceStyle_Upsert_Spanner,
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
	case consts.FloatField:
		return "FLOAT32"
	case consts.Float64Field:
		return "FLOAT64"
	case consts.TextField:
		if field.Length > 0 {
			return fmt.Sprintf("STRING(%d)", field.Length)
		}
		return "STRING"
	case consts.JsonField:
		return "JSON"
	case consts.BoolField:
		return "BOOL"
	case consts.TimestampField, consts.TimestampUTCField:
		return "TIMESTAMP"
	case consts.BlobField:
		if field.Length > 0 {
			return fmt.Sprintf("BYTES(%d)", field.Length)
		}
		return "BYTES"
	case consts.DateField:
		return "DATE"
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
				def := column.Default
				if def == `"epoch"` {
					def = "timestamp_seconds(0)"
				}
				//
				if column.Type == "JSON" && strings.HasPrefix(column.Default, `"`) {
					def = fmt.Sprintf("JSON %s", column.Default)
				}
				dir.Add(Lf("DEFAULT (%s)", def))
			}
			if column.Type == "INT64_SEQ" {
				seqName := fmt.Sprintf("%s_%s", table.Name, column.Name)
				stmts = append(stmts, Build(Lf("CREATE SEQUENCE %s OPTIONS (sequence_kind='bit_reversed_positive')", seqName)).SQL())
				dir.Add(Lf("DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE %s))", seqName))
			}
			dirs = append(dirs, dir.SQL())
		}

		for _, column := range table.Columns {
			if ref := column.Reference; ref != nil {
				action := ""
				if ref.OnDelete != "" {
					action += fmt.Sprintf(" ON DELETE %s ", ref.OnDelete)
				}
				if ref.OnUpdate != "" {
					action += fmt.Sprintf(" ON UPDATE %s ", ref.OnUpdate)
				}
				dir := Build(Lf("CONSTRAINT %s_%s_fkey FOREIGN KEY (%s) REFERENCES %s (%s)%s",
					table.Name, column.Name, column.Name, ref.Table, ref.Column, action))
				dirs = append(dirs, dir.SQL())
			}
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
		for _, unique := range table.Unique {
			indexName := fmt.Sprintf("index_%s_%s", table.Name, unique[0])
			stmts = append(stmts, Build(Lf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, table.Name, unique[0])).SQL())
		}
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

		stmts = append(stmts, sqlcompile.Compile(stmt.SQL()))
	}

	return stmts
}

// DropSchema implements Dialect.
func (s *spanner) DropSchema(schema *Schema) (res []sqlgen.SQL) {
	var stmts []sqlgen.SQL

	tables := schema.Tables
	slices.Reverse(tables)

	for _, table := range schema.Tables {
		for _, column := range table.Columns {
			if ref := column.Reference; ref != nil {
				dir := Build(Lf("ALTER TABLE %s DROP CONSTRAINT %s_%s_fkey", table.Name, table.Name, column.Name))
				stmts = append(stmts, sqlcompile.Compile(dir.SQL()))
			}
		}
		for _, unique := range table.Unique {
			indexName := fmt.Sprintf("index_%s_%s", table.Name, unique[0])
			stmts = append(stmts, Build(Lf("DROP INDEX IF EXISTS %s", indexName)).SQL())
		}
	}

	for _, index := range schema.Indexes {
		stmt := Build(L("DROP INDEX IF EXISTS"))
		stmt.Add(Lf(index.Name))
		stmts = append(stmts, sqlcompile.Compile(stmt.SQL()))
	}
	for _, table := range tables {
		for _, pk := range table.PrimaryKey {
			stmts = append(stmts, sqlcompile.Compile(Lf("ALTER TABLE  %s ALTER %s SET DEFAULT (null)", table.Name, pk)))
			stmts = append(stmts, sqlcompile.Compile(Lf("DROP SEQUENCE IF EXISTS %s_%s", table.Name, pk)))
		}
		stmts = append(stmts, sqlcompile.Compile(Lf("DROP TABLE IF EXISTS %s", table.Name)))
	}

	return stmts
}
