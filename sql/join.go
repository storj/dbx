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

type Join struct {
	Type  string
	Table string
	Left  string
	Right string
}

func JoinFromIRJoin(ir_join *ir.Join) Join {
	join := Join{
		Table: ir_join.Right.Model.Table,
		Left:  ir_join.Left.ColumnRef(),
		Right: ir_join.Right.ColumnRef(),
	}
	switch ir_join.Type {
	case consts.InnerJoin:
	default:
		panic(fmt.Sprintf("unhandled join type %q", join.Type))
	}
	return join
}

func JoinsFromIRJoins(ir_joins []*ir.Join) (joins []Join) {
	for _, ir_join := range ir_joins {
		joins = append(joins, JoinFromIRJoin(ir_join))
	}
	return joins
}

func SQLFromJoin(join Join) sqlgen.SQL {
	clause := Build(Lf("%s JOIN %s ON %s =", join.Type, join.Table, join.Left))
	if join.Right != "" {
		clause.Add(L(join.Right))
	} else {
		clause.Add(Placeholder)
	}
	return sqlcompile.Compile(clause.SQL())
}

func SQLFromJoins(joins []Join) []sqlgen.SQL {
	var out []sqlgen.SQL
	for _, join := range joins {
		out = append(out, SQLFromJoin(join))
	}
	return out
}
