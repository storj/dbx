// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ast

import (
	"fmt"
	"strings"
	"text/scanner"

	"storj.io/dbx/consts"
)

type Root struct {
	Models  []*Model
	Creates []*Create
	Reads   []*Read
	Updates []*Update
	Deletes []*Delete
}

func (root *Root) Add(next *Root) {
	root.Creates = append(root.Creates, next.Creates...)
	root.Models = append(root.Models, next.Models...)
	root.Reads = append(root.Reads, next.Reads...)
	root.Updates = append(root.Updates, next.Updates...)
	root.Deletes = append(root.Deletes, next.Deletes...)
}

type String struct {
	Pos   scanner.Position
	Value string
}

func (s *String) Get() string {
	if s == nil {
		return ""
	}
	return s.Value
}

type Model struct {
	Pos        scanner.Position
	Name       *String
	Table      *String
	Fields     []*Field
	PrimaryKey *RelativeFieldRefs
	Unique     []*RelativeFieldRefs
	Indexes    []*Index
}

type Bool struct {
	Pos   scanner.Position
	Value bool
}

func (b *Bool) Get() bool {
	if b == nil {
		return false
	}
	return b.Value
}

func (b *Bool) String() string {
	return fmt.Sprint(b.Value)
}

type Int struct {
	Pos   scanner.Position
	Value int
}

func (i *Int) Get() int {
	if i == nil {
		return 0
	}
	return i.Value
}

type Suffix struct {
	Pos   scanner.Position
	Parts []*String
}

type Field struct {
	Pos  scanner.Position
	Name *String

	// Common to both regular and relation fields
	Column    *String
	Nullable  *Bool
	Updatable *Bool

	// Only make sense on a regular field
	Type       *FieldType
	AutoInsert *Bool
	AutoUpdate *Bool
	Length     *Int
	Default    *String

	// Only make sense on a relation
	Relation     *FieldRef
	RelationKind *RelationKind
}

type RelationKind struct {
	Pos   scanner.Position
	Value consts.RelationKind
}

type FieldType struct {
	Pos   scanner.Position
	Value consts.FieldType
}

type FieldRef struct {
	Pos   scanner.Position
	Model *String
	Field *String
}

func (r *FieldRef) String() string {
	if r.Field == nil {
		return r.Model.Value
	}
	if r.Model == nil {
		return r.Field.Value
	}
	return fmt.Sprintf("%s.%s", r.Model.Value, r.Field.Value)
}

func (f *FieldRef) Relative() *RelativeFieldRef {
	return &RelativeFieldRef{
		Pos:   f.Pos,
		Field: f.Field,
	}
}

func (f *FieldRef) ModelRef() *ModelRef {
	return &ModelRef{
		Pos:   f.Pos,
		Model: f.Model,
	}
}

type RelativeFieldRefs struct {
	Pos  scanner.Position
	Refs []*RelativeFieldRef
}

type RelativeFieldRef struct {
	Pos   scanner.Position
	Field *String
}

func (r *RelativeFieldRef) String() string {
	return r.Field.Value
}

type ModelRef struct {
	Pos   scanner.Position
	Model *String
}

func (m *ModelRef) String() string {
	return m.Model.Value
}

type Index struct {
	Pos     scanner.Position
	Name    *String
	Fields  *RelativeFieldRefs
	Unique  *Bool
	Where   []*Where
	Storing *RelativeFieldRefs
}

type Read struct {
	Pos     scanner.Position
	Select  *FieldRefs
	Joins   []*Join
	Where   []*Where
	OrderBy *OrderBy
	GroupBy *GroupBy
	View    *View
	Suffix  *Suffix
}

type Delete struct {
	Pos    scanner.Position
	Model  *ModelRef
	Joins  []*Join
	Where  []*Where
	Suffix *Suffix
}

type Update struct {
	Pos      scanner.Position
	Model    *ModelRef
	Joins    []*Join
	Where    []*Where
	NoReturn *Bool
	Suffix   *Suffix
}

type Create struct {
	Pos      scanner.Position
	Model    *ModelRef
	Raw      *Bool
	NoReturn *Bool
	Replace  *Bool
	Suffix   *Suffix
}

type View struct {
	Pos         scanner.Position
	All         *Bool
	LimitOffset *Bool
	Paged       *Bool
	Count       *Bool
	Has         *Bool
	Scalar      *Bool
	One         *Bool
	First       *Bool
}

type FieldRefs struct {
	Pos  scanner.Position
	Refs []*FieldRef
}

type Join struct {
	Pos   scanner.Position
	Left  *FieldRef
	Right *FieldRef
	Type  *JoinType
}

type JoinType struct {
	Pos   scanner.Position
	Value consts.JoinType
}

func (j *JoinType) Get() consts.JoinType {
	if j == nil {
		return consts.InnerJoin
	}
	return j.Value
}

type Where struct {
	Pos     scanner.Position
	Clauses []*Clause // 2 or more joined by or
}

func (w *Where) String() string {
	clauses := make([]string, 0, len(w.Clauses))
	for _, clause := range w.Clauses {
		clauses = append(clauses, clause.String())
	}
	if len(clauses) > 1 {
		return "(" + strings.Join(clauses, " or ") + ")"
	}
	return strings.Join(clauses, " or ")
}

type Clause struct {
	Pos   scanner.Position
	Left  *Expr
	Op    *Operator
	Right *Expr
}

func (c *Clause) String() string {
	return fmt.Sprintf("%s %s %s", c.Left, c.Op, c.Right)
}

type Expr struct {
	Pos scanner.Position
	// The following fields are mutually exclusive
	Null        *Null
	StringLit   *String
	NumberLit   *String
	BoolLit     *Bool
	Placeholder *Placeholder
	FieldRef    *FieldRef
	FuncCall    *FuncCall
}

func (e *Expr) String() string {
	switch {
	case e.Null != nil:
		return e.Null.String()
	case e.StringLit != nil:
		return fmt.Sprintf("%q", e.StringLit.Value)
	case e.NumberLit != nil:
		return e.NumberLit.Value
	case e.BoolLit != nil:
		return e.BoolLit.String()
	case e.Placeholder != nil:
		return e.Placeholder.String()
	case e.FieldRef != nil:
		return e.FieldRef.String()
	case e.FuncCall != nil:
		return e.FuncCall.String()
	default:
		return ""
	}
}

type Null struct {
	Pos scanner.Position
}

func (p *Null) String() string {
	return "null"
}

type Placeholder struct {
	Pos scanner.Position
}

func (p *Placeholder) String() string {
	return "?"
}

type FuncCall struct {
	Pos  scanner.Position
	Name *String
	Args []*Expr
}

func (f *FuncCall) String() string {
	var args []string
	for _, arg := range f.Args {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name.Value, strings.Join(args, ", "))
}

type Operator struct {
	Pos   scanner.Position
	Value consts.Operator
}

func (o *Operator) String() string { return string(o.Value) }

type OrderBy struct {
	Pos     scanner.Position
	Entries []*OrderByEntry
}

type OrderByEntry struct {
	Pos        scanner.Position
	Field      *FieldRef
	Descending *Bool
}

type GroupBy struct {
	Pos    scanner.Position
	Fields *FieldRefs
}
