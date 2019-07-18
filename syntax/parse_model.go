// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
)

func parseModel(node *tupleNode) (*ast.Model, error) {
	model := new(ast.Model)
	model.Pos = node.getPos()

	name_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	model.Name = stringFromToken(name_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"table": func(node *tupleNode) error {
			if model.Table != nil {
				return previouslyDefined(node.getPos(), "model", "table",
					model.Table.Pos)
			}

			name_token, err := node.consumeToken(Ident)
			if err != nil {
				return err
			}
			model.Table = stringFromToken(name_token)

			return nil
		},
		"field": func(node *tupleNode) error {
			field, err := parseField(node)
			if err != nil {
				return err
			}
			model.Fields = append(model.Fields, field)

			return nil
		},
		"key": func(node *tupleNode) error {
			if model.PrimaryKey != nil {
				return previouslyDefined(node.getPos(), "model", "key",
					model.PrimaryKey.Pos)
			}
			primary_key, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			model.PrimaryKey = primary_key
			return nil
		},
		"unique": func(node *tupleNode) error {
			unique, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			model.Unique = append(model.Unique, unique)
			return nil
		},
		"index": func(node *tupleNode) error {
			index, err := parseIndex(node)
			if err != nil {
				return err
			}
			model.Indexes = append(model.Indexes, index)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}
