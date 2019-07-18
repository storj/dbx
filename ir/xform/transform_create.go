// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/ir"
)

func transformCreate(lookup *lookup, ast_cre *ast.Create) (
	cre *ir.Create, err error) {

	model, err := lookup.FindModel(ast_cre.Model)
	if err != nil {
		return nil, err
	}

	cre = &ir.Create{
		Model:    model,
		Raw:      ast_cre.Raw.Get(),
		NoReturn: ast_cre.NoReturn.Get(),
		Suffix:   transformSuffix(ast_cre.Suffix),
	}
	if cre.Suffix == nil {
		cre.Suffix = DefaultCreateSuffix(cre)
	}

	return cre, nil
}
