// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseView(node *tupleNode) (*ast.View, error) {
	view := new(ast.View)
	view.Pos = node.getPos()

	err := node.consumeTokensNamedUntilList(tokenCases{
		{Ident, "all"}:   tokenFlagField("view", "all", &view.All),
		{Ident, "paged"}: tokenFlagField("view", "paged", &view.Paged),
		{Ident, "count"}: tokenFlagField("view", "count", &view.Count),
		{Ident, "has"}:   tokenFlagField("view", "has", &view.Has),
		{Ident, "limitoffset"}: tokenFlagField("view", "limitoffset",
			&view.LimitOffset),
		{Ident, "scalar"}: tokenFlagField("view", "scalar", &view.Scalar),
		{Ident, "one"}:    tokenFlagField("view", "one", &view.One),
		{Ident, "first"}:  tokenFlagField("view", "first", &view.First),
	})
	if err != nil {
		return nil, err
	}

	return view, nil
}
