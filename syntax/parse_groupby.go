// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseGroupBy(node *tupleNode) (*ast.GroupBy, error) {
	group_by := new(ast.GroupBy)
	group_by.Pos = node.getPos()

	field_refs, err := parseFieldRefs(node, true)
	if err != nil {
		return nil, err
	}
	group_by.Fields = field_refs

	return group_by, nil
}
