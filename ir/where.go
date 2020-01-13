// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import "storj.io/dbx/consts"

type Where struct {
	Clause *Clause
	Or     *[2]*Where
	And    *[2]*Where
}

type Clause struct {
	Left  *Expr
	Op    consts.Operator
	Right *Expr
}

func (c *Clause) NeedsCondition() bool {
	// only EQ and NE need a condition to switch on "=" v.s. "is", etc.
	switch c.Op {
	case consts.EQ, consts.NE:
	default:
		return false
	}

	// null values are fixed and don't need a runtime condition to render
	// appropriately
	if c.Left.Null || c.Right.Null {
		return false
	}

	return c.Left.Nullable() && c.Right.Nullable()
}
