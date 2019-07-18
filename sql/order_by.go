// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

type OrderBy struct {
	Fields     []string
	Descending bool
}

func OrderByFromIROrderBy(ir_order_by *ir.OrderBy) (order_by *OrderBy) {
	order_by = &OrderBy{
		Descending: ir_order_by.Descending,
	}
	for _, ir_field := range ir_order_by.Fields {
		order_by.Fields = append(order_by.Fields, ir_field.ColumnRef())
	}
	return order_by
}

func SQLFromOrderBy(order_by *OrderBy) sqlgen.SQL {
	stmt := Build(L("ORDER BY"))
	stmt.Add(J(", ", Strings(order_by.Fields)...))
	if order_by.Descending {
		stmt.Add(L("DESC"))
	}
	return sqlcompile.Compile(stmt.SQL())
}
