// Copyright (C) 2020 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/consts"
)

func parseClause(node *tupleNode) (clause *ast.Clause, err error) {
	clause = new(ast.Clause)
	clause.Pos = node.getPos()

	clause.Left, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	err = node.consumeTokenNamed(tokenCases{
		Exclamation.tokenCase(): func(token *tokenNode) error {
			_, err := node.consumeToken(Equal)
			if err != nil {
				return err
			}
			clause.Op = operatorFromValue(token, consts.NE)
			return nil
		},
		{Ident, "like"}: func(token *tokenNode) error {
			clause.Op = operatorFromValue(token, consts.Like)
			return nil
		},
		Equal.tokenCase(): func(token *tokenNode) error {
			clause.Op = operatorFromValue(token, consts.EQ)
			return nil
		},
		LeftAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				clause.Op = operatorFromValue(token, consts.LE)
			} else {
				clause.Op = operatorFromValue(token, consts.LT)
			}
			return nil
		},
		RightAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				clause.Op = operatorFromValue(token, consts.GE)
			} else {
				clause.Op = operatorFromValue(token, consts.GT)
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	clause.Right, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	return clause, nil
}
