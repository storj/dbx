// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import "storj.io/dbx/consts"

type Where struct {
	Left  *Expr
	Op    consts.Operator
	Right *Expr
}

func (w *Where) NeedsCondition() bool {
	// only EQ and NE need a condition to switch on "=" v.s. "is", etc.
	switch w.Op {
	case consts.EQ, consts.NE:
	default:
		return false
	}

	// null values are fixed and don't need a runtime condition to render
	// appropriately
	if w.Left.Null || w.Right.Null {
		return false
	}

	return w.Left.Nullable() && w.Right.Nullable()
}
