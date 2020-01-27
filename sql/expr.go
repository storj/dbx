// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"

	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func ExprSQL(expr *ir.Expr, dialect Dialect) sqlgen.SQL {
	switch {
	case expr.Null:
		return L("NULL")
	case expr.StringLit != nil:
		return J("", L("'"), L(dialect.EscapeString(*expr.StringLit)), L("'"))
	case expr.NumberLit != nil:
		return L(*expr.NumberLit)
	case expr.BoolLit != nil:
		return L(dialect.BoolLit(*expr.BoolLit))
	case expr.Placeholder == 1:
		return L("?")
	case expr.Placeholder > 0:
		return J("", L("("), J(", ", Placeholders(expr.Placeholder)...), L(")"))
	case expr.Field != nil:
		return L(expr.Field.ColumnRef())
	case expr.FuncCall != nil:
		var args []sqlgen.SQL
		for _, arg := range expr.FuncCall.Args {
			args = append(args, ExprSQL(arg, dialect))
		}
		return J("", L(expr.FuncCall.Name), L("("), J(", ", args...), L(")"))
	case len(expr.Row) > 0:
		refs := make([]string, 0, len(expr.Row))
		for _, field := range expr.Row {
			refs = append(refs, field.ColumnRef())
		}
		return J("", L("("), J(", ", Strings(refs)...), L(")"))
	default:
		panic(fmt.Sprintf("unhandled expression variant: %+v", expr))
	}
}
