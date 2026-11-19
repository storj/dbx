// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/consts"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformClauses(lookup *lookup, models map[string]scanner.Position, ast_clauses []*ast.Clause) (clauses []*ir.Clause, err error) {
	for _, ast_clause := range ast_clauses {
		clause, err := transformClause(lookup, models, ast_clause)
		if err != nil {
			return nil, err
		}

		clauses = append(clauses, clause)
	}
	return clauses, nil
}

func transformClause(lookup *lookup, models map[string]scanner.Position, ast_clause *ast.Clause) (clause *ir.Clause, err error) {
	lexpr, err := transformExpr(lookup, models, ast_clause.Left, true)
	if err != nil {
		return nil, err
	}

	rexpr, err := transformExpr(lookup, models, ast_clause.Right, false)
	if err != nil {
		return nil, err
	}

	if ast_clause.Op.Value == consts.In {
		// "in" takes a list of values at runtime, so it only supports a
		// plain field on the left and a placeholder on the right.
		if lexpr.Field == nil {
			return nil, errutil.New(ast_clause.Pos,
				"the left side of an \"in\" clause must be a field")
		}
		if rexpr.Placeholder == 0 {
			return nil, errutil.New(ast_clause.Pos,
				"the right side of an \"in\" clause must be a placeholder")
		}
		if lexpr.Field.Type == consts.JsonField {
			return nil, errutil.New(ast_clause.Pos,
				"\"in\" clauses are not supported on json fields")
		}
	}

	return &ir.Clause{
		Left:  lexpr,
		Op:    ast_clause.Op.Value,
		Right: rexpr,
	}, nil
}
