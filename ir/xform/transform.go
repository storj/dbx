// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/ir"
)

func Transform(ast_root *ast.Root) (root *ir.Root, err error) {
	lookup := newLookup()

	models, err := transformModels(lookup, ast_root.Models)
	if err != nil {
		return nil, err
	}
	models, err = ir.SortModels(models)
	if err != nil {
		return nil, err
	}

	root = &ir.Root{
		Models: models,
	}

	create_signatures := map[string]*ast.Create{}
	for _, ast_cre := range ast_root.Creates {
		cre, err := transformCreate(lookup, ast_cre)
		if err != nil {
			return nil, err
		}

		if existing := create_signatures[cre.Signature()]; existing != nil {
			return nil, duplicateQuery(ast_cre.Pos, "create", existing.Pos)
		}
		create_signatures[cre.Signature()] = ast_cre

		root.Creates = append(root.Creates, cre)
	}

	read_signatures := map[string]*ast.Read{}
	for _, ast_read := range ast_root.Reads {
		reads, err := transformRead(lookup, ast_read)
		if err != nil {
			return nil, err
		}

		for _, read := range reads {
			if existing := read_signatures[read.Signature()]; existing != nil {
				return nil, duplicateQuery(ast_read.Pos, "read", existing.Pos)
			}
			read_signatures[read.Signature()] = ast_read
		}

		root.Reads = append(root.Reads, reads...)
	}

	update_signatures := map[string]*ast.Update{}
	for _, ast_upd := range ast_root.Updates {
		upd, err := transformUpdate(lookup, ast_upd)
		if err != nil {
			return nil, err
		}

		if existing := update_signatures[upd.Signature()]; existing != nil {
			return nil, duplicateQuery(ast_upd.Pos, "update", existing.Pos)
		}
		update_signatures[upd.Signature()] = ast_upd

		root.Updates = append(root.Updates, upd)
	}

	delete_signatures := map[string]*ast.Delete{}
	for _, ast_del := range ast_root.Deletes {
		del, err := transformDelete(lookup, ast_del)
		if err != nil {
			return nil, err
		}

		if existing := delete_signatures[del.Signature()]; existing != nil {
			return nil, duplicateQuery(ast_del.Pos, "delete", existing.Pos)
		}
		delete_signatures[del.Signature()] = ast_del

		root.Deletes = append(root.Deletes, del)
	}

	return root, nil
}
