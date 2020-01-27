// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

type Expr struct {
	Null        bool
	Placeholder int
	StringLit   *string
	NumberLit   *string
	BoolLit     *bool
	Field       *Field
	FuncCall    *FuncCall
	Row         []*Field
}

func (e *Expr) Nullable() bool {
	switch {
	case e.Null, e.Placeholder > 0:
		return true
	case e.Field != nil:
		return e.Field.Nullable
	case e.FuncCall != nil:
		return e.FuncCall.Nullable()
	default:
		return false
	}
}

func (e *Expr) HasPlaceholder() bool {
	if e.Placeholder > 0 {
		return true
	}
	if e.FuncCall != nil {
		return e.FuncCall.HasPlaceholder()
	}
	return false
}

func (e *Expr) Unique() bool {
	return e.Field != nil && e.Field.Unique()
}

type FuncCall struct {
	Name string
	Args []*Expr
}

func (fc *FuncCall) Nullable() bool {
	for _, arg := range fc.Args {
		if arg.Nullable() {
			return true
		}
	}
	return false
}

func (fc *FuncCall) HasPlaceholder() bool {
	for _, arg := range fc.Args {
		if arg.HasPlaceholder() {
			return true
		}
	}
	return false
}
