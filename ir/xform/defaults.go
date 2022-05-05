// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"fmt"
	"strings"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
)

func DefaultIndexName(i *ir.Index) string {
	parts := []string{i.Model.Table}
	for _, field := range i.Fields {
		parts = append(parts, field.Column)
	}
	if i.Unique {
		parts = append(parts, "unique")
	}
	parts = append(parts, "index")
	return strings.Join(parts, "_")
}

func DefaultCreateSuffix(cre *ir.Create) []string {
	var parts []string
	parts = append(parts, cre.Model.Name)
	return parts
}

func DefaultReadSuffix(read *ir.Read) []string {
	var parts []string
	for _, selectable := range read.Selectables {
		switch obj := selectable.(type) {
		case *ir.Model:
			parts = append(parts, obj.Name)
		case *ir.Field:
			parts = append(parts, obj.Model.Name)
			parts = append(parts, obj.Name)
		default:
			panic(fmt.Sprintf("unhandled selectable %T", selectable))
		}
	}
	full := len(read.Joins) > 0
	parts = append(parts, whereSuffix(read.Where, full)...)
	if read.OrderBy != nil {
		parts = append(parts, "order_by")
		for _, entry := range read.OrderBy.Entries {
			if entry.Descending {
				parts = append(parts, "desc")
			} else {
				parts = append(parts, "asc")
			}
			if full {
				parts = append(parts, entry.Field.Model.Name)
			}
			parts = append(parts, entry.Field.Name)
		}
	}
	if read.GroupBy != nil {
		parts = append(parts, "group_by")
		for _, field := range read.GroupBy.Fields {
			if full {
				parts = append(parts, field.Model.Name)
			}
			parts = append(parts, field.Name)
		}
	}

	return parts
}

func DefaultUpdateSuffix(upd *ir.Update) []string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, whereSuffix(upd.Where, len(upd.Joins) > 0)...)
	return parts
}

func DefaultDeleteSuffix(del *ir.Delete) []string {
	var parts []string
	parts = append(parts, del.Model.Name)
	parts = append(parts, whereSuffix(del.Where, len(del.Joins) > 0)...)
	return parts
}

func whereSuffix(wheres []*ir.Where, full bool) (parts []string) {
	if len(wheres) == 0 {
		return nil
	}
	parts = append(parts, "by")
	for i, where := range wheres {
		if i > 0 {
			parts = append(parts, "and")
		}
		parts = append(parts, singleWhereSuffix(where, full)...)
	}
	return parts
}

func singleWhereSuffix(where *ir.Where, full bool) (parts []string) {
	switch {
	case where.Clause != nil:
		return clauseSuffix(where.Clause, full)

	case where.And != nil:
		parts = append(parts, "")
		parts = append(parts, singleWhereSuffix(where.And[0], full)...)
		parts = append(parts, "and")
		parts = append(parts, singleWhereSuffix(where.And[1], full)...)
		return parts

	case where.Or != nil:
		parts = append(parts, "")
		parts = append(parts, singleWhereSuffix(where.Or[0], full)...)
		parts = append(parts, "or")
		parts = append(parts, singleWhereSuffix(where.Or[1], full)...)
		return parts

	default:
		panic("invalid where")
	}
}

func clauseSuffix(clause *ir.Clause, full bool) (parts []string) {
	left := exprSuffix(clause.Left, full)
	right := exprSuffix(clause.Right, full)

	parts = append(parts, left...)
	if len(right) > 0 || clause.Op != consts.EQ {
		op := clause.Op.Suffix()
		nulloperand := clause.Left.Null || clause.Right.Null
		switch clause.Op {
		case consts.EQ:
			if nulloperand {
				parts = append(parts, "is")
			} else {
				parts = append(parts, op)
			}
		case consts.NE:
			if nulloperand {
				parts = append(parts, "is not")
			} else {
				parts = append(parts, op)
			}
		default:
			parts = append(parts, op)
		}
	}
	if len(right) > 0 {
		parts = append(parts, right...)
	}

	return parts
}

func exprSuffix(expr *ir.Expr, full bool) (parts []string) {
	switch {
	case expr.Null:
		parts = []string{"null"}
	case expr.StringLit != nil:
		parts = []string{"string"}
	case expr.NumberLit != nil:
		parts = []string{"number"}
	case expr.BoolLit != nil:
		parts = []string{fmt.Sprint(*expr.BoolLit)}
	case expr.Placeholder > 0:
	case expr.Field != nil:
		if full {
			parts = append(parts, expr.Field.Model.Name)
		}
		parts = append(parts, expr.Field.Name)
	case expr.FuncCall != nil:
		parts = append(parts, expr.FuncCall.Name)
		for i, arg := range expr.FuncCall.Args {
			arg_suffix := exprSuffix(arg, full)
			if len(arg_suffix) == 0 {
				continue
			}
			if i != 0 {
				parts = append(parts, "and")
			}
			parts = append(parts, arg_suffix...)
		}
	default:
		panic(fmt.Sprintf("unhandled expr for suffix: %+v", expr))
	}
	return parts
}
