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

	clauses, err := transformClauses(lookup, models, ast_where.Clauses)
	if err != nil {
		return nil, err
	}

	// recursively construct a chain of or clauses
	var xform func(clauses []*ir.Clause) *ir.Where
	xform = func(clauses []*ir.Clause) *ir.Where {
		if len(clauses) == 0 {
			return where
		} else if len(clauses) == 1 {
			return &ir.Where{Clause: clauses[0]}
		} else {
			return &ir.Where{
				Or: &[2]*ir.Where{
					{Clause: clauses[0]},
					xform(clauses[1:]),
				},
			}
		}
	}

	return xform(clauses), nil
}
