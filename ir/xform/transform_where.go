// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/ir"
)

func transformWheres(lookup *lookup, models map[string]scanner.Position,
	ast_wheres []*ast.Where) (wheres []*ir.Where, err error) {
	for _, ast_where := range ast_wheres {
		where, err := transformWhere(lookup, models, ast_where)
		if err != nil {
			return nil, err
		}

		wheres = append(wheres, where)
	}
	return wheres, nil
}

func transformWhere(lookup *lookup, models map[string]scanner.Position,
	ast_where *ast.Where) (where *ir.Where, err error) {

	lexpr, err := transformExpr(lookup, models, ast_where.Left, true)
	if err != nil {
		return nil, err
	}

	rexpr, err := transformExpr(lookup, models, ast_where.Right, false)
	if err != nil {
		return nil, err
	}

	// TODO: it's easier to support `or` in the grammar now

	return &ir.Where{Clause: &ir.Clause{
		Left:  lexpr,
		Op:    ast_where.Op.Value,
		Right: rexpr,
	}}, nil
}
