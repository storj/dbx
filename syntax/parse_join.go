// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseJoin(node *tupleNode) (*ast.Join, error) {
	join := new(ast.Join)
	join.Pos = node.getPos()

	left_field_ref, err := parseFieldRef(node, true)
	if err != nil {
		return nil, err
	}
	join.Left = left_field_ref

	_, err = node.consumeToken(Equal)
	if err != nil {
		return nil, err
	}

	right_field_ref, err := parseFieldRef(node, true)
	if err != nil {
		return nil, err
	}
	join.Right = right_field_ref

	return join, nil
}
