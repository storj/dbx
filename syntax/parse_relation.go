// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/consts"
)

func parseRelation(node *tupleNode, field *ast.Field) error {
	err := node.consumeTokenNamed(tokenCases{
		{Ident, "setnull"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.SetNull)
			return nil
		},
		{Ident, "cascade"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.Cascade)
			return nil
		},
		{Ident, "restrict"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.Restrict)
			return nil
		},
	})
	if err != nil {
		return err
	}

	attributes_list := node.consumeIfList()
	if attributes_list != nil {
		err := attributes_list.consumeAnyTuples(tupleCases{
			"column": func(node *tupleNode) error {
				if field.Column != nil {
					return previouslyDefined(node.getPos(), "relation", "column",
						field.Column.Pos)
				}

				name_token, err := node.consumeToken(Ident)
				if err != nil {
					return err
				}
				field.Column = stringFromToken(name_token)

				return nil
			},
			"nullable": tupleFlagField("relation", "nullable",
				&field.Nullable),
			"updatable": tupleFlagField("relation", "updatable",
				&field.Updatable),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
