// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func transformModels(lookup *lookup, ast_models []*ast.Model) (
	models []*ir.Model, err error) {

	// step 1. create all the Model and Field instances and set their pointers
	// to point at each other appropriately.
	for _, ast_model := range ast_models {
		link, err := lookup.AddModel(ast_model)
		if err != nil {
			return nil, err
		}
		for _, ast_field := range ast_model.Fields {
			if err := link.AddField(ast_field); err != nil {
				return nil, err
			}
		}
	}

	// step 2. resolve all of the other fields on the models and Fields
	// including references between them. also check for duplicate table names.
	table_names := map[string]*ast.Model{}
	for _, ast_model := range ast_models {
		model_entry := lookup.GetModel(ast_model.Name.Value)
		if err := transformModel(lookup, model_entry); err != nil {
			return nil, err
		}

		model := model_entry.model

		if existing := table_names[model.Table]; existing != nil {
			return nil, errutil.New(ast_model.Pos,
				"table %q already used by model %q (%s)",
				model.Table, existing.Name.Get(), existing.Pos)
		}
		table_names[model.Table] = ast_model

		models = append(models, model_entry.model)
	}

	return models, nil
}
