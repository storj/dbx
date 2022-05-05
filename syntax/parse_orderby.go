// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func parseOrderBy(node *tupleNode) (*ast.OrderBy, error) {
	order_by := new(ast.OrderBy)
	order_by.Pos = node.getPos()

	entries := node.consumeIfList()
	if entries != nil {
		for len(entries.value) > 0 {
			entryTuple, err := entries.consumeTuple()
			if err != nil {
				return nil, err
			}
			entry, err := parseOrderByEntry(entryTuple)
			if err != nil {
				return nil, err
			}
			order_by.Entries = append(order_by.Entries, entry)
		}
	} else {
		entry, err := parseOrderByEntry(node)
		if err != nil {
			return nil, err
		}
		order_by.Entries = append(order_by.Entries, entry)
	}

	return order_by, nil
}

func parseOrderByEntry(node *tupleNode) (*ast.OrderByEntry, error) {
	entry := new(ast.OrderByEntry)
	entry.Pos = node.getPos()

	err := node.consumeTokenNamed(tokenCases{
		{Ident, "asc"}: func(token *tokenNode) error { return nil },
		{Ident, "desc"}: func(token *tokenNode) error {
			entry.Descending = boolFromValue(token, true)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	field, err := parseFieldRef(node, true)
	if err != nil {
		return nil, err
	}
	entry.Field = field

	return entry, nil
}
