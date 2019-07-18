// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"strconv"

	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
)

func parseExpr(node *tupleNode) (*ast.Expr, error) {
	// expressions are one of the following:
	//  placeholder      : ?
	//  string literal   : "foo"
	//  int literal      : 9
	//  float literal    : 9.9
	//  dotted field ref : model.field
	//  function         : foo(<expr>)

	expr := &ast.Expr{
		Pos: node.getPos(),
	}

	first, err := node.consumeToken(Question, String, Int, Float, Ident)
	if err != nil {
		return nil, err
	}

	switch first.tok {
	case Question:
		expr.Placeholder = placeholderFromToken(first)
		return expr, nil
	case String:
		unquoted, err := strconv.Unquote(first.text)
		if err != nil {
			return nil, errutil.New(first.getPos(),
				"(internal) unable to unquote string token text: %s", err)
		}
		expr.StringLit = stringFromValue(first, unquoted)
		return expr, nil
	case Int, Float:
		expr.NumberLit = stringFromValue(first, first.text)
		return expr, nil
	}

	first.debugAssertToken(Ident)

	if node.consumeIfToken(Dot) == nil {
		switch first.text {
		case "null":
			expr.Null = nullFromToken(first)
			return expr, nil
		case "true", "false":
			expr.BoolLit = boolFromToken(first)
			return expr, nil
		}

		list, err := node.consumeList()
		if err != nil {
			return nil, err
		}
		exprs, err := parseExprs(list)
		if err != nil {
			return nil, err
		}
		expr.FuncCall = funcCallFromTokenAndArgs(first, exprs)
		return expr, nil
	}

	second, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}

	expr.FieldRef = fieldRefFromTokens(first, second)
	return expr, nil
}

func parseExprs(list *listNode) (exprs []*ast.Expr, err error) {

	for {
		tuple, err := list.consumeTupleOrEmpty()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			return exprs, nil
		}
		expr, err := parseExpr(tuple)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
}
