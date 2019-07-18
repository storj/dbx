// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseDelete(node *tupleNode) (*ast.Delete, error) {
	del := new(ast.Delete)
	del.Pos = node.getPos()

	model_ref_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	del.Model = modelRefFromToken(model_ref_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"where": func(node *tupleNode) error {
			where, err := parseWhere(node)
			if err != nil {
				return err
			}
			del.Where = append(del.Where, where)

			return nil
		},
		"join": func(node *tupleNode) error {
			join, err := parseJoin(node)
			if err != nil {
				return err
			}
			del.Joins = append(del.Joins, join)

			return nil
		},
		"suffix": func(node *tupleNode) error {
			if del.Suffix != nil {
				return previouslyDefined(node.getPos(), "delete", "suffix",
					del.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			del.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return del, nil
}
