// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseCreate(node *tupleNode) (*ast.Create, error) {
	cre := new(ast.Create)
	cre.Pos = node.getPos()

	model_ref_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	cre.Model = modelRefFromToken(model_ref_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"raw":      tupleFlagField("create", "raw", &cre.Raw),
		"noreturn": tupleFlagField("create", "noreturn", &cre.NoReturn),
		"replace":  tupleFlagField("create", "replace", &cre.Replace),
		"suffix": func(node *tupleNode) error {
			if cre.Suffix != nil {
				return previouslyDefined(node.getPos(), "create", "suffix",
					cre.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			cre.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return cre, nil
}
