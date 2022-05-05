// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import (
	"fmt"

	"storj.io/dbx/consts"
)

type Selectable interface {
	SelectRefs() []string
	ModelOf() *Model
	selectable()
}

type Read struct {
	Suffix      []string
	Selectables []Selectable
	From        *Model
	Joins       []*Join
	Where       []*Where
	OrderBy     *OrderBy
	GroupBy     *GroupBy
	View        View
}

func (r *Read) Signature() string {
	return fmt.Sprintf("READ(%q,%q)", r.Suffix, r.View)
}

func (r *Read) Distinct() bool {
	var targets []*Model
	for _, selectable := range r.Selectables {
		targets = append(targets, selectable.ModelOf())
	}
	return queryUnique(distinctModels(targets), r.Joins, r.Where)
}

// SelectedModel returns the single model being selected or nil if there are
// more than one selectable or the selectable is a field.
func (r *Read) SelectedModel() *Model {
	if len(r.Selectables) == 1 {
		if model, ok := r.Selectables[0].(*Model); ok {
			return model
		}
	}
	return nil
}

type View string

const (
	All         View = "all"
	LimitOffset View = "limitoffset"
	Paged       View = "paged"
	Count       View = "count"
	Has         View = "has"
	Scalar      View = "scalar"
	One         View = "one"
	First       View = "first"
)

type Join struct {
	Type  consts.JoinType
	Left  *Field
	Right *Field
}

type OrderBy struct {
	Entries []*OrderByEntry
}

type OrderByEntry struct {
	Field      *Field
	Descending bool
}

type GroupBy struct {
	Fields []*Field
}
