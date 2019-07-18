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

var (
	countFields = []string{"COUNT(*)"}
	hasFields   = []string{"1"}
)

func SelectSQL(ir_read *ir.Read, dialect Dialect) sqlgen.SQL {
	return SQLFromSelect(SelectFromIRRead(ir_read, dialect))
}

func GetLastSQL(ir_model *ir.Model, dialect Dialect) sqlgen.SQL {
	return SQLFromSelect(&Select{
		Fields: ir_model.SelectRefs(),
		From:   ir_model.Table,
		Where: []sqlgen.SQL{
			J(" ", L(dialect.RowId()), L("="), L("?")),
		},
	})
}

type Select struct {
	From    string
	Fields  []string
	Joins   []Join
	Where   []sqlgen.SQL
	OrderBy *OrderBy
	GroupBy *GroupBy
	Limit   string
	Offset  string
	Has     bool
}

func SelectFromIRRead(ir_read *ir.Read, dialect Dialect) *Select {
	sel := &Select{
		From:  ir_read.From.Table,
		Where: WhereSQL(ir_read.Where, dialect),
		Joins: JoinsFromIRJoins(ir_read.Joins),
	}

	for _, ir_selectable := range ir_read.Selectables {
		sel.Fields = append(sel.Fields, ir_selectable.SelectRefs()...)
	}

	if ir_read.OrderBy != nil {
		sel.OrderBy = OrderByFromIROrderBy(ir_read.OrderBy)
	}
	if ir_read.GroupBy != nil {
		sel.GroupBy = GroupByFromIRGroupBy(ir_read.GroupBy)
	}

	switch ir_read.View {
	case ir.All:
	case ir.One, ir.Scalar:
		if !ir_read.Distinct() {
			sel.Limit = "2"
		}
	case ir.LimitOffset:
		sel.Limit = "?"
		sel.Offset = "?"
	case ir.Paged:
		pk := ir_read.From.PagablePrimaryKey()
		sel.Where = append(sel.Where, WhereSQL([]*ir.Where{&ir.Where{
			Left: &ir.Expr{
				Field: pk,
			},
			Op: consts.GT,
			Right: &ir.Expr{
				Placeholder: true,
			},
		}}, dialect)...)
		sel.OrderBy = &OrderBy{
			Fields: []string{pk.ColumnRef()},
		}
		sel.Limit = "?"
		sel.Fields = append(sel.Fields, pk.SelectRefs()...)
	case ir.Has:
		sel.Has = true
		sel.Fields = hasFields
		sel.OrderBy = nil
	case ir.Count:
		sel.Fields = countFields
		sel.OrderBy = nil
	case ir.First:
		sel.Limit = "1"
		sel.Offset = "0"
	default:
		panic(fmt.Sprintf("unsupported select view %s", ir_read.View))
	}

	return sel
}

func SQLFromSelect(sel *Select) sqlgen.SQL {
	stmt := Build(nil)

	if sel.Has {
		stmt.Add(L("SELECT EXISTS("))
	}

	fields := J(", ", Strings(sel.Fields)...)
	stmt.Add(L("SELECT"), fields, Lf("FROM %s", sel.From))

	if joins := SQLFromJoins(sel.Joins); len(joins) > 0 {
		stmt.Add(joins...)
	}

	if len(sel.Where) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", sel.Where...))
	}

	if sel.OrderBy != nil {
		stmt.Add(SQLFromOrderBy(sel.OrderBy))
	}

	if sel.GroupBy != nil {
		stmt.Add(SQLFromGroupBy(sel.GroupBy))
	}

	if sel.Limit != "" {
		stmt.Add(Lf("LIMIT %s", sel.Limit))
	}

	if sel.Offset != "" {
		stmt.Add(Lf("OFFSET %s", sel.Offset))
	}

	if sel.Has {
		stmt.Add(L(")"))
	}

	return sqlcompile.Compile(stmt.SQL())
}
