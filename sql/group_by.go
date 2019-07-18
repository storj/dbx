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

type GroupBy struct {
	Fields []string
}

func GroupByFromIRGroupBy(ir_group_by *ir.GroupBy) (group_by *GroupBy) {
	group_by = &GroupBy{}
	for _, ir_field := range ir_group_by.Fields {
		group_by.Fields = append(group_by.Fields, ir_field.ColumnRef())
	}
	return group_by
}

func SQLFromGroupBy(group_by *GroupBy) sqlgen.SQL {
	stmt := Build(L("GROUP BY"))
	stmt.Add(J(", ", Strings(group_by.Fields)...))
	return sqlcompile.Compile(stmt.SQL())
}
