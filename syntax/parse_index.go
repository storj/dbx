// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseIndex(node *tupleNode) (*ast.Index, error) {
	index := new(ast.Index)
	index.Pos = node.getPos()

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"name": func(node *tupleNode) error {
			if index.Name != nil {
				return previouslyDefined(node.getPos(), "index", "name",
					index.Name.Pos)
			}

			name_token, err := node.consumeToken(Ident)
			if err != nil {
				return err
			}
			index.Name = stringFromToken(name_token)

			return nil
		},
		"fields": func(node *tupleNode) error {
			if index.Fields != nil {
				return previouslyDefined(node.getPos(), "index", "fields",
					index.Fields.Pos)
			}

			fields, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			index.Fields = fields

			return nil
		},
		"unique": tupleFlagField("index", "unique", &index.Unique),
		"storing": func(node *tupleNode) error {
			if index.Storing != nil {
				return previouslyDefined(node.getPos(), "index", "storing",
					index.Storing.Pos)
			}

			storing, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			index.Storing = storing

			return nil
		},
		"where": func(node *tupleNode) error {
			where, err := parseWhere(node)
			if err != nil {
				return err
			}
			index.Where = append(index.Where, where)

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return index, nil
}
