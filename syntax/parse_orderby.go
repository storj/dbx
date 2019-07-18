// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseOrderBy(node *tupleNode) (*ast.OrderBy, error) {
	order_by := new(ast.OrderBy)
	order_by.Pos = node.getPos()

	err := node.consumeTokenNamed(tokenCases{
		{Ident, "asc"}: func(token *tokenNode) error { return nil },
		{Ident, "desc"}: func(token *tokenNode) error {
			order_by.Descending = boolFromValue(token, true)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	field_refs, err := parseFieldRefs(node, true)
	if err != nil {
		return nil, err
	}
	order_by.Fields = field_refs

	return order_by, nil
}
