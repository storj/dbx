// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseUpdate(node *tupleNode) (*ast.Update, error) {
	upd := new(ast.Update)
	upd.Pos = node.getPos()

	model_ref_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	upd.Model = modelRefFromToken(model_ref_token)

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
			upd.Where = append(upd.Where, where)

			return nil
		},
		"join": func(node *tupleNode) error {
			join, err := parseJoin(node)
			if err != nil {
				return err
			}
			upd.Joins = append(upd.Joins, join)

			return nil
		},
		"noreturn": tupleFlagField("update", "noreturn", &upd.NoReturn),
		"suffix": func(node *tupleNode) error {
			if upd.Suffix != nil {
				return previouslyDefined(node.getPos(), "update", "suffix",
					upd.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			upd.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return upd, nil
}
