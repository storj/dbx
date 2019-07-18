// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseSuffix(node *tupleNode) (*ast.Suffix, error) {
	suf := new(ast.Suffix)
	suf.Pos = node.getPos()

	tokens, err := node.consumeTokens(Ident)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		suf.Parts = append(suf.Parts, stringFromToken(token))
	}

	return suf, nil
}
