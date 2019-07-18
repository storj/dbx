// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformDelete(lookup *lookup, ast_del *ast.Delete) (
	del *ir.Delete, err error) {

	model, err := lookup.FindModel(ast_del.Model)
	if err != nil {
		return nil, err
	}

	if len(model.PrimaryKey) > 1 && len(ast_del.Joins) > 0 {
		return nil, errutil.New(ast_del.Joins[0].Pos,
			"delete with joins unsupported on multicolumn primary key")
	}

	del = &ir.Delete{
		Model:  model,
		Suffix: transformSuffix(ast_del.Suffix),
	}

	models, joins, err := transformJoins(
		lookup, []*ir.Model{model}, ast_del.Joins)
	if err != nil {
		return nil, err
	}
	models[model.Name] = ast_del.Model.Pos

	del.Joins = joins

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	del.Where, err = transformWheres(lookup, models, ast_del.Where)
	if err != nil {
		return nil, err
	}

	if del.Suffix == nil {
		del.Suffix = DefaultDeleteSuffix(del)
	}

	return del, nil
}
