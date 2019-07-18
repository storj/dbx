// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

type lookup struct {
	models map[string]*modelEntry
}

type modelEntry struct {
	model  *ir.Model
	ast    *ast.Model
	fields map[string]*fieldEntry
}

type fieldEntry struct {
	field *ir.Field
	ast   *ast.Field
}

func newLookup() *lookup {
	return &lookup{
		models: make(map[string]*modelEntry),
	}
}

func (l *lookup) AddModel(ast_model *ast.Model) (link *modelEntry, err error) {
	if existing, ok := l.models[ast_model.Name.Value]; ok {
		return nil, previouslyDefined(ast_model.Pos, "model", existing.ast.Pos)
	}

	link = newModelEntry(ast_model)
	l.models[ast_model.Name.Value] = link
	return link, nil
}

func (l *lookup) GetModel(name string) *modelEntry {
	return l.models[name]
}

func (l *lookup) FindModel(ref *ast.ModelRef) (*ir.Model, error) {
	link := l.models[ref.Model.Value]
	if link != nil {
		return link.model, nil
	}
	return nil, errutil.New(ref.Pos, "no model %q defined",
		ref.Model.Value)
}

func (l *lookup) FindField(ref *ast.FieldRef) (*ir.Field, error) {
	model_link := l.models[ref.Model.Value]
	if model_link == nil {
		return nil, errutil.New(ref.Pos, "no model %q defined",
			ref.Model.Value)
	}
	return model_link.FindField(ref.Relative())
}

func newModelEntry(ast_model *ast.Model) *modelEntry {
	return &modelEntry{
		model: &ir.Model{
			Name: ast_model.Name.Value,
		},
		ast:    ast_model,
		fields: make(map[string]*fieldEntry),
	}
}

func (m *modelEntry) newFieldEntry(ast_field *ast.Field) *fieldEntry {
	field := &ir.Field{
		Name:  ast_field.Name.Value,
		Model: m.model,
	}
	if ast_field.Type != nil {
		field.Type = ast_field.Type.Value
	}
	m.model.Fields = append(m.model.Fields, field)

	return &fieldEntry{
		field: field,
		ast:   ast_field,
	}
}

func (m *modelEntry) AddField(ast_field *ast.Field) (err error) {
	if existing, ok := m.fields[ast_field.Name.Value]; ok {
		return previouslyDefined(ast_field.Pos, "field", existing.ast.Pos)
	}
	m.fields[ast_field.Name.Value] = m.newFieldEntry(ast_field)
	return nil
}

func (m *modelEntry) GetField(name string) *fieldEntry {
	return m.fields[name]
}

func (m *modelEntry) FindField(ref *ast.RelativeFieldRef) (*ir.Field, error) {
	field_link := m.fields[ref.Field.Value]
	if field_link == nil {
		return nil, errutil.New(ref.Pos, "no field %q defined on model %q",
			ref.Field.Value, m.model.Name)
	}
	return field_link.field, nil
}
