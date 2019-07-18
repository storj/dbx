// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/consts"
)

func parseWhere(node *tupleNode) (where *ast.Where, err error) {
	where = new(ast.Where)
	where.Pos = node.getPos()

	where.Left, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	err = node.consumeTokenNamed(tokenCases{
		Exclamation.tokenCase(): func(token *tokenNode) error {
			_, err := node.consumeToken(Equal)
			if err != nil {
				return err
			}
			where.Op = operatorFromValue(token, consts.NE)
			return nil
		},
		{Ident, "like"}: func(token *tokenNode) error {
			where.Op = operatorFromValue(token, consts.Like)
			return nil
		},
		Equal.tokenCase(): func(token *tokenNode) error {
			where.Op = operatorFromValue(token, consts.EQ)
			return nil
		},
		LeftAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				where.Op = operatorFromValue(token, consts.LE)
			} else {
				where.Op = operatorFromValue(token, consts.LT)
			}
			return nil
		},
		RightAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				where.Op = operatorFromValue(token, consts.GE)
			} else {
				where.Op = operatorFromValue(token, consts.GT)
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	where.Right, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	return where, nil
}
