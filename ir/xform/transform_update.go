// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformUpdate(lookup *lookup, ast_upd *ast.Update) (
	upd *ir.Update, err error) {

	model, err := lookup.FindModel(ast_upd.Model)
	if err != nil {
		return nil, err
	}
	if !model.HasUpdatableFields() {
		return nil, errutil.New(ast_upd.Pos,
			"update on model with no updatable fields")
	}

	if len(model.PrimaryKey) > 1 && len(ast_upd.Joins) > 0 {
		return nil, errutil.New(ast_upd.Joins[0].Pos,
			"update with joins unsupported on multicolumn primary key:")
	}

	upd = &ir.Update{
		Model:    model,
		NoReturn: ast_upd.NoReturn.Get(),
		Suffix:   transformSuffix(ast_upd.Suffix),
	}

	models, joins, err := transformJoins(
		lookup, []*ir.Model{model}, ast_upd.Joins)
	if err != nil {
		return nil, err
	}
	models[model.Name] = ast_upd.Model.Pos

	upd.Joins = joins

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	upd.Where, err = transformWheres(lookup, models, ast_upd.Where)
	if err != nil {
		return nil, err
	}

	if !upd.One() {
		return nil, errutil.New(ast_upd.Pos,
			"updates for more than one row are unsupported")
	}

	if upd.Suffix == nil {
		upd.Suffix = DefaultUpdateSuffix(upd)
	}

	return upd, nil
}
