// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import "storj.io/dbx/ir"

type PartitionedArgs struct {
	AllArgs      []*Var
	StaticArgs   []*Var
	NullableArgs []*Var
}

func PartitionedArgsFromWheres(wheres []*ir.Where) (out PartitionedArgs) {
	for _, where := range wheres {
		if !where.Right.HasPlaceholder() {
			continue
		}

		arg := ArgFromWhere(where)
		out.AllArgs = append(out.AllArgs, arg)

		if where.NeedsCondition() {
			out.NullableArgs = append(out.NullableArgs, arg)
		} else {
			out.StaticArgs = append(out.StaticArgs, arg)
		}
	}
	return out
}
