// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"fmt"
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformExpr(lookup *lookup, models map[string]scanner.Position, ast_expr *ast.Expr, is_left bool) (expr *ir.Expr, err error) {
	switch {
	case ast_expr.Null != nil:
		if is_left {
			return nil, errutil.New(ast_expr.Pos,
				"null is not valid on the left side of a where clause")
		}
		return &ir.Expr{
			Null: true,
		}, nil
	case ast_expr.StringLit != nil:
		if is_left {
			return nil, errutil.New(ast_expr.Pos,
				"literals are not valid on the left side of a where clause")
		}
		return &ir.Expr{
			StringLit: &ast_expr.StringLit.Value,
		}, nil
	case ast_expr.NumberLit != nil:
		if is_left {
			return nil, errutil.New(ast_expr.Pos,
				"literals are not valid on the left side of a where clause")
		}
		return &ir.Expr{
			NumberLit: &ast_expr.NumberLit.Value,
		}, nil
	case ast_expr.BoolLit != nil:
		if is_left {
			return nil, errutil.New(ast_expr.Pos,
				"literals are not valid on the left side of a where clause")
		}
		return &ir.Expr{
			BoolLit: &ast_expr.BoolLit.Value,
		}, nil
	case ast_expr.Placeholder != nil:
		if is_left {
			return nil, errutil.New(ast_expr.Pos,
				"placeholders are not valid on the left side of a where clause")
		}
		return &ir.Expr{
			Placeholder: 1,
		}, nil
	case ast_expr.FieldRef != nil:
		if _, ok := models[ast_expr.FieldRef.Model.Value]; !ok {
			return nil, errutil.New(ast_expr.Pos,
				"invalid where condition %q; model %q is not joined",
				ast_expr, ast_expr.FieldRef.Model.Value)
		}
		field, err := lookup.FindField(ast_expr.FieldRef)
		if err != nil {
			return nil, err
		}
		return &ir.Expr{
			Field: field,
		}, nil
	case ast_expr.FuncCall != nil:
		func_call, err := transformFuncCall(lookup, models, ast_expr.FuncCall,
			is_left)
		if err != nil {
			return nil, err
		}
		return &ir.Expr{
			FuncCall: func_call,
		}, nil
	default:
		panic(fmt.Sprintf("unhandled expression: %+v", ast_expr))
	}
}

func transformFuncCall(lookup *lookup, models map[string]scanner.Position, ast_func_call *ast.FuncCall, is_left bool) (*ir.FuncCall, error) {
	var args []*ir.Expr
	for _, ast_expr := range ast_func_call.Args {
		arg, err := transformExpr(lookup, models, ast_expr, is_left)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	// TODO: better arg validation (type checks, etc.)
	name := ast_func_call.Name.Value
	switch name {
	case "lower":
		if err := checkArgs(ast_func_call, args, 1); err != nil {
			return nil, err
		}
	default:
		return nil, errutil.New(ast_func_call.Name.Pos,
			"unknown function %q", name)
	}

	return &ir.FuncCall{
		Name: name,
		Args: args,
	}, nil
}

func checkArgs(ast_func_call *ast.FuncCall, args []*ir.Expr, expected_count int) (err error) {
	if len(args) != expected_count {
		return errutil.New(ast_func_call.Pos,
			"expected %d argument, got %d", expected_count, len(args))
	}
	return nil
}
