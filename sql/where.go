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
	// we put all the condition wheres at the end for ease of template
	// generation later.

	for _, where := range wheres {
		if where.NeedsCondition() {
			continue
		}
		out = append(out,
			J(" ", ExprSQL(where.Left, dialect),
				opSQL(where.Op, where.Left, where.Right),
				ExprSQL(where.Right, dialect)))
	}

	conditions := 0
	for _, where := range wheres {
		if !where.NeedsCondition() {
			continue
		}
		out = append(out, &sqlgen.Condition{
			Name:  fmt.Sprintf("cond_%d", conditions),
			Left:  ExprSQL(where.Left, dialect).Render(),
			Equal: where.Op == "=",
			Right: ExprSQL(where.Right, dialect).Render(),
		})
		conditions++
	}

	return out
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
