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

func parseFieldRefOrEmpty(node *tupleNode, needs_dot bool) (*ast.FieldRef, error) {
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

// isAggregateFunc checks if the given identifier is an aggregate function
func isAggregateFunc(name string) bool {
	switch name {
	case "sum", "count", "avg", "min", "max":
		return true
	}
	return false
}

// parseSelectItems parses items in a SELECT clause that can be field refs or aggregate functions
func parseSelectItems(node *tupleNode) (*ast.SelectItems, error) {
	items := &ast.SelectItems{
		Pos: node.getPos(),
	}

	for {
		// Try to consume an identifier
		first, err := node.consumeTokenOrEmpty(Ident)
		if err != nil {
			return nil, err
		}
		if first == nil {
			if len(items.Items) == 0 {
				return nil, errutil.New(node.getPos(),
					"must specify some fields or aggregates to select")
			}
			return items, nil
		}

		item := &ast.SelectItem{
			Pos: first.getPos(),
		}

		// Check if this is an aggregate function
		if isAggregateFunc(first.text) {
			// Parse aggregate function: func(field) or func(*)
			list := node.consumeIfList()
			if list == nil {
				return nil, errutil.New(first.getPos(),
					"aggregate function %q requires parentheses", first.text)
			}

			agg := &ast.AggregateFunc{
				Pos:  first.getPos(),
				Func: stringFromValue(first, first.text),
			}

			// Parse the argument inside the parentheses
			tuple, err := list.consumeTupleOrEmpty()
			if err != nil {
				return nil, err
			}

			if tuple != nil {
				// Check for count(*)
				asterisk := tuple.consumeIfToken(Asterisk)
				if asterisk != nil {
					// count(*) - FieldRef stays nil
					if first.text != "count" {
						return nil, errutil.New(asterisk.getPos(),
							"only count() supports * argument")
					}
				} else {
					// Parse field reference: model.field
					fieldFirst, fieldSecond, err := tuple.consumeDottedIdents()
					if err != nil {
						return nil, err
					}
					agg.FieldRef = fieldRefFromTokens(fieldFirst, fieldSecond)
				}

				if err := tuple.assertEmpty(); err != nil {
					return nil, err
				}
			} else if first.text != "count" {
				return nil, errutil.New(first.getPos(),
					"aggregate function %q requires a field argument", first.text)
			}

			item.Aggregate = agg
		} else {
			// Regular field reference: model or model.field
			var second *tokenNode
			if node.consumeIfToken(Dot) != nil {
				second, err = node.consumeToken(Ident)
				if err != nil {
					return nil, err
				}
			}
			item.FieldRef = fieldRefFromTokens(first, second)
		}

		items.Items = append(items.Items, item)
	}
}
