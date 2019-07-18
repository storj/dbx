// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
)

func parseFieldRefs(node *tupleNode, needs_dot bool) (*ast.FieldRefs, error) {
	refs := new(ast.FieldRefs)
	refs.Pos = node.getPos()

	for {
		ref, err := parseFieldRefOrEmpty(node, needs_dot)
		if err != nil {
			return nil, err
		}
		if ref == nil {
			return refs, nil
		}
		refs.Refs = append(refs.Refs, ref)
	}
}

func parseFieldRefOrEmpty(node *tupleNode, needs_dot bool) (
	*ast.FieldRef, error) {

	first, second, err := node.consumeDottedIdentsOrEmpty()
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, nil
	}
	if second == nil && needs_dot {
		return nil, errutil.New(first.getPos(),
			"field ref must specify a model")
	}
	return fieldRefFromTokens(first, second), nil
}

func parseFieldRef(node *tupleNode, needs_dot bool) (*ast.FieldRef, error) {
	first, second, err := node.consumeDottedIdents()
	if err != nil {
		return nil, err
	}
	if second == nil && needs_dot {
		return nil, errutil.New(first.getPos(),
			"field ref must specify a model")
	}
	return fieldRefFromTokens(first, second), nil
}

func parseRelativeFieldRefs(node *tupleNode) (*ast.RelativeFieldRefs, error) {
	refs := new(ast.RelativeFieldRefs)
	refs.Pos = node.getPos()

	for {
		ref_token, err := node.consumeTokenOrEmpty(Ident)
		if err != nil {
			return nil, err
		}
		if ref_token == nil {
			if len(refs.Refs) == 0 {
				return nil, errutil.New(node.getPos(),
					"must specify some field references")
			}
			return refs, nil
		}
		refs.Refs = append(refs.Refs, relativeFieldRefFromToken(ref_token))
	}
}
