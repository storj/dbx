// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func SchemaSQL(ir_root *ir.Root, dialect Dialect) sqlgen.SQL {
	return SQLFromSchema(SchemaFromIRModels(ir_root.Models, dialect))
}

type Schema struct {
	Tables  []Table
	Indexes []Index
}

type Table struct {
	Name       string
	Columns    []Column
	PrimaryKey []string
	Unique     [][]string
}

type Column struct {
	Name      string
	Type      string
	NotNull   bool
	Default   string
	Reference *Reference
}

type Reference struct {
	Table    string
	Column   string
	OnDelete string
	OnUpdate string
}

type Index struct {
	Name    string
	Table   string
	Columns []string
	Unique  bool
	Where   []sqlgen.SQL
	Storing []string
}

func SchemaFromIRModels(ir_models []*ir.Model, dialect Dialect) *Schema {
	schema := &Schema{}
	for _, ir_model := range ir_models {
		table := Table{
			Name: ir_model.Table,
		}
		for _, ir_field := range ir_model.PrimaryKey {
			table.PrimaryKey = append(table.PrimaryKey, ir_field.Column)
		}
		for _, ir_unique := range ir_model.Unique {
			var unique []string
			for _, ir_field := range ir_unique {
				unique = append(unique, ir_field.Column)
			}
			table.Unique = append(table.Unique, unique)
		}
		for _, ir_field := range ir_model.Fields {
			column := Column{
				Name:    ir_field.Column,
				Type:    dialect.ColumnType(ir_field),
				NotNull: !ir_field.Nullable,
				Default: dialect.DefaultLit(ir_field.Default),
			}
			if ir_field.Relation != nil {
				column.Reference = &Reference{
					Table:  ir_field.Relation.Field.Model.Table,
					Column: ir_field.Relation.Field.Column,
				}
				switch ir_field.Relation.Kind {
				case consts.SetNull:
					column.Reference.OnDelete = "SET NULL"

				case consts.Cascade:
					column.Reference.OnDelete = "CASCADE"

				case consts.Restrict:
					column.Reference.OnDelete = ""

				default:
					panic(fmt.Sprintf("unhandled relation kind %q",
						ir_field.Relation.Kind))
				}
			}
			table.Columns = append(table.Columns, column)
		}
		schema.Tables = append(schema.Tables, table)
		for _, ir_index := range ir_model.Indexes {
			index := Index{
				Name:   ir_index.Name,
				Table:  ir_index.Model.Table,
				Unique: ir_index.Unique,
				Where:  WhereSQL(ir_index.Where, dialect),
			}
			if dialect.Features().Storing {
				for _, ir_storing := range ir_index.Storing {
					index.Storing = append(index.Storing, ir_storing.Column)
				}
			}
			for _, ir_field := range ir_index.Fields {
				index.Columns = append(index.Columns, ir_field.Column)
			}
			schema.Indexes = append(schema.Indexes, index)
		}
	}
	return schema
}

func SQLFromSchema(schema *Schema) sqlgen.SQL {
	var stmts []sqlgen.SQL

	for _, table := range schema.Tables {
		var dirs []sqlgen.SQL

		for _, column := range table.Columns {
			dir := Build(Lf("%s %s", column.Name, column.Type))
			if column.NotNull {
				dir.Add(L("NOT NULL"))
			}
			if column.Default != "" {
				dir.Add(Lf("DEFAULT %s", column.Default))
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

		if pkey := table.PrimaryKey; len(pkey) > 0 {
			dir := Build(L("PRIMARY KEY ("))
			dir.Add(J(", ", Strings(pkey)...))
			dir.Add(L(")"))
			dirs = append(dirs, dir.SQL())
		}

		for _, unique := range table.Unique {
			dir := Build(L("UNIQUE ("))
			dir.Add(J(", ", Strings(unique)...))
			dir.Add(L(")"))
			dirs = append(dirs, dir.SQL())
		}

		directives := J(",\n\t", dirs...)

		stmt := J("",
			Lf("CREATE TABLE %s (\n\t", table.Name),
			directives,
			Lf("\n);"),
		)

		stmts = append(stmts, stmt)
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

		stmts = append(stmts, stmt.SQL())
	}

	return sqlcompile.Compile(J("\n", stmts...))
}
