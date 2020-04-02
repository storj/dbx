// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
)

func parseWhere(node *tupleNode) (where *ast.Where, err error) {
	where = new(ast.Where)
	where.Pos = node.getPos()

	clauses := node.consumeIfList()
	if clauses != nil {
		for len(clauses.value) > 0 {
			clauseTuple, err := clauses.consumeTuple()
			if err != nil {
				return nil, err
			}
			clause, err := parseClause(clauseTuple)
			if err != nil {
				return nil, err
			}
			where.Clauses = append(where.Clauses, clause)
		}
	} else {
		clause, err := parseClause(node)
		if err != nil {
			return nil, err
		}
		where.Clauses = append(where.Clauses, clause)
	}

	return where, nil
}
