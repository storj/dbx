// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

func resolveFieldRefs(lookup *lookup, ast_refs []*ast.FieldRef) (
	fields []*ir.Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := lookup.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func resolveRelativeFieldRefs(model_entry *modelEntry,
	ast_refs []*ast.RelativeFieldRef) (fields []*ir.Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := model_entry.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func previouslyDefined(pos scanner.Position, kind string,
	where scanner.Position) error {

	return errutil.New(pos,
		"%s already defined. previous definition at %s",
		kind, where)
}

func duplicateQuery(pos scanner.Position, kind string,
	where scanner.Position) error {
	return errutil.New(pos,
		"%s: duplicate %s (first defined at %s)",
		pos, kind, where)
}
