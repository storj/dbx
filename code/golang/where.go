// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"fmt"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
)

type ConditionArg struct {
	Var         *Var
	IsCondition bool
	Condition   int
}

func ConditionArgsFromWheres(wheres []*ir.Where) (out []ConditionArg) {
	gen := newConditionArgGenerator()
	for _, where := range wheres {
		out = append(out, gen.FromWhere(where)...)
	}
	return out
}

type conditionArgGenerator struct {
	condition int
	names     map[string]int
}

func newConditionArgGenerator() *conditionArgGenerator {
	return &conditionArgGenerator{names: make(map[string]int)}
}

func (c *conditionArgGenerator) FromClause(clause *ir.Clause) *ConditionArg {
	// Placeholders are normalized to always be on the right. If we don't
	// have a placeholder, we don't have an argument.
	if !clause.Right.HasPlaceholder() {
		return nil
	}

	// TODO: clean this up when we do full expression type evaluation.
	// assume for now that the left hand side evaluates eventually to a single
	// field wrapped in zero or more function calls since that is all that is
	// possible via the xform package.
	expr := clause.Left
	for expr.Field == nil {
		expr = expr.FuncCall.Args[0]
	}

	name := expr.Field.UnderRef()
	if clause.Op != consts.EQ {
		name += "_" + clause.Op.Suffix()
	}

	arg := new(ConditionArg)

	// we don't set ZeroVal or InitVal because these args should only be used
	// as incoming arguments to function calls.
	arg.Var = &Var{
		Name: name,
		Type: ModelFieldFromIR(expr.Field).StructName(),
	}

	if clause.NeedsCondition() {
		arg.IsCondition = true
		arg.Condition = c.condition
		c.condition++
	}

	c.names[arg.Var.Name]++
	if count := c.names[arg.Var.Name]; count > 1 {
		arg.Var.Name += fmt.Sprintf("_%d", count)
	}

	return arg
}

func (c *conditionArgGenerator) FromWhere(where *ir.Where) (out []ConditionArg) {
	switch {
	case where.Clause != nil:
		arg := c.FromClause(where.Clause)
		if arg != nil {
			out = append(out, *arg)
		}

	case where.And != nil:
		out = append(out, c.FromWhere(where.And[0])...)
		out = append(out, c.FromWhere(where.And[1])...)

	case where.Or != nil:
		out = append(out, c.FromWhere(where.Or[0])...)
		out = append(out, c.FromWhere(where.Or[1])...)
	}

	return out
}
