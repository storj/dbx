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
	Entries []OrderByEntry
}

type OrderByEntry struct {
	Field      string
	Descending bool
}

func OrderByFromIROrderBy(ir_order_by *ir.OrderBy) (order_by *OrderBy) {
	order_by = new(OrderBy)
	for _, entry := range ir_order_by.Entries {
		order_by.Entries = append(order_by.Entries, OrderByEntry{
			Field:      entry.Field.ColumnRef(),
			Descending: entry.Descending,
		})
	}
	return order_by
}

func SQLFromOrderBy(order_by *OrderBy) sqlgen.SQL {
	stmt := Build(L("ORDER BY"))
	var entries []sqlgen.SQL
	for _, entry := range order_by.Entries {
		clause := Build(L(entry.Field))
		if entry.Descending {
			clause.Add(L("DESC"))
		}
		entries = append(entries, clause.SQL())
	}
	stmt.Add(J(", ", entries...))
	return sqlcompile.Compile(stmt.SQL())
}
