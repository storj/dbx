// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xform

import (
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformJoins(lookup *lookup, in_scope []*ir.Model, ast_joins []*ast.Join) (models map[string]scanner.Position, joins []*ir.Join, err error) {
	models = make(map[string]scanner.Position)

	in_scope_set := make(map[*ir.Model]bool)
	for _, model := range in_scope {
		in_scope_set[model] = true
	}

	for _, ast_join := range ast_joins {
		left, err := lookup.FindField(ast_join.Left)
		if err != nil {
			return nil, nil, err
		}
		if !in_scope_set[left.Model] {
			return nil, nil, errutil.New(ast_join.Left.Pos,
				"model %q not in scope to join on", left.Model.Name)
		}

		right, err := lookup.FindField(ast_join.Right)
		if err != nil {
			return nil, nil, err
		}
		in_scope_set[right.Model] = true

		joins = append(joins, &ir.Join{
			Type:  ast_join.Type.Get(),
			Left:  left,
			Right: right,
		})

		models[ast_join.Left.Model.Value] = ast_join.Left.Pos
		models[ast_join.Right.Model.Value] = ast_join.Right.Pos
	}

	return models, joins, nil
}
