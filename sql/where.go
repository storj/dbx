// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"
	"strings"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func WhereSQL(wheres []*ir.Where, dialect Dialect) (out []sqlgen.SQL) {
	gen := new(whereGenerator)
	for _, where := range wheres {
		out = append(out, gen.whereSQL(where, dialect))
	}
	return out
}

type whereGenerator struct {
	conditions int
}

func (g *whereGenerator) whereSQL(where *ir.Where, dialect Dialect) sqlgen.SQL {
	switch {
	case where.Clause != nil:
		return g.clauseSQL(where.Clause, dialect)

	case where.And != nil:
		left := g.whereSQL(where.And[0], dialect)
		right := g.whereSQL(where.And[1], dialect)
		return J("", L("("), left, L(" AND "), right, L(")"))

	case where.Or != nil:
		left := g.whereSQL(where.Or[0], dialect)
		right := g.whereSQL(where.Or[1], dialect)
		return J("", L("("), left, L(" OR "), right, L(")"))

	default:
		panic("exhaustive match")
	}
}

func (g *whereGenerator) clauseSQL(clause *ir.Clause, dialect Dialect) sqlgen.SQL {
	if clause.NeedsCondition() {
		g.conditions++
		return &sqlgen.Condition{
			Name:  fmt.Sprintf("cond_%d", g.conditions-1),
			Left:  ExprSQL(clause.Left, dialect).Render(),
			Equal: clause.Op == "=",
			Right: ExprSQL(clause.Right, dialect).Render(),
		}
	}
	return J(" ",
		ExprSQL(clause.Left, dialect),
		opSQL(clause.Op, clause.Left, clause.Right),
		ExprSQL(clause.Right, dialect))
}

func opSQL(op consts.Operator, left, right *ir.Expr) sqlgen.SQL {
	switch op {
	case consts.EQ:
		if left.Null || right.Null {
			return L("is")
		}
	case consts.NE:
		if left.Null || right.Null {
			return L("is not")
		}
	}
	return L(strings.ToUpper(string(op)))
}
