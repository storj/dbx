// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"fmt"
	"text/scanner"

	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/internal/inflect"
	"storj.io/dbx/ir"
)

func transformModel(lookup *lookup, model_entry *modelEntry) (err error) {
	model := model_entry.model
	ast_model := model_entry.ast

	model.Name = ast_model.Name.Value
	model.Table = ast_model.Table.Get()
	if model.Table == "" {
		model.Table = inflect.Pluralize(model.Name)
	}

	column_names := map[string]*ast.Field{}
	for _, ast_field := range ast_model.Fields {
		field_entry := model_entry.GetField(ast_field.Name.Value)
		if err := transformField(lookup, field_entry); err != nil {
			return err
		}

		field := field_entry.field

		if existing := column_names[field.Column]; existing != nil {
			return errutil.New(ast_field.Pos,
				"column %q already used by field %q at %s",
				field.Column, existing.Name.Get(), existing.Pos)
		}
		column_names[field.Column] = ast_field
	}

	if ast_model.PrimaryKey == nil || len(ast_model.PrimaryKey.Refs) == 0 {
		return errutil.New(ast_model.Pos, "no primary key defined")
	}

	for _, ast_fieldref := range ast_model.PrimaryKey.Refs {
		field, err := model_entry.FindField(ast_fieldref)
		if err != nil {
			return err
		}
		if field.Nullable {
			return errutil.New(ast_fieldref.Pos,
				"nullable field %q cannot be a primary key",
				ast_fieldref)
		}
		if field.Updatable {
			return errutil.New(ast_fieldref.Pos,
				"updatable field %q cannot be a primary key",
				ast_fieldref)
		}
		model.PrimaryKey = append(model.PrimaryKey, field)
	}

	for _, ast_unique := range ast_model.Unique {
		fields, err := resolveRelativeFieldRefs(model_entry, ast_unique.Refs)
		if err != nil {
			return err
		}
		model.Unique = append(model.Unique, fields)
	}

	index_names := map[string]*ast.Index{}
	for _, ast_index := range ast_model.Indexes {
		// BUG(jeff): we can only have one index without a name specified when
		// really we want to pick a name for them that won't collide.
		if ast_index.Fields == nil || len(ast_index.Fields.Refs) < 1 {
			return errutil.New(ast_index.Pos,
				"index %q has no fields defined",
				ast_index.Name.Get())
		}

		fields, err := resolveRelativeFieldRefs(
			model_entry, ast_index.Fields.Refs)
		if err != nil {
			return err
		}

		var storing []*ir.Field
		if ast_index.Storing != nil {
			storing, err = resolveRelativeFieldRefs(
				model_entry, ast_index.Storing.Refs)
			if err != nil {
				return err
			}
		}

		models := map[string]scanner.Position{model.Name: ast_model.Pos}
		where, err := transformWheres(lookup, models, ast_index.Where)
		if err != nil {
			return err
		}

		index := &ir.Index{
			Name:    ast_index.Name.Get(),
			Model:   fields[0].Model,
			Fields:  fields,
			Unique:  ast_index.Unique.Get(),
			Where:   where,
			Storing: storing,
		}

		if index.Name == "" {
			index.Name = DefaultIndexName(index)
		}

		if existing, ok := index_names[index.Name]; ok {
			return previouslyDefined(ast_index.Pos,
				fmt.Sprintf("index (%s)", index.Name),
				existing.Pos)
		}
		index_names[index.Name] = ast_index

		model.Indexes = append(model.Indexes, index)
	}

	return nil
}
