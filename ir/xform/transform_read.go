// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import (
	"storj.io/dbx/ast"
	"storj.io/dbx/consts"
	"storj.io/dbx/errutil"
	"storj.io/dbx/ir"
)

// transformAggregate converts an AST aggregate function to an IR aggregate
func transformAggregate(lookup *lookup, ast_agg *ast.AggregateFunc) (*ir.Aggregate, error) {
	agg := &ir.Aggregate{}

	// Set the function type
	switch ast_agg.Func.Value {
	case "sum":
		agg.Func = ir.AggSum
	case "count":
		agg.Func = ir.AggCount
	case "avg":
		agg.Func = ir.AggAvg
	case "min":
		agg.Func = ir.AggMin
	case "max":
		agg.Func = ir.AggMax
	default:
		return nil, errutil.New(ast_agg.Pos, "unknown aggregate function %q", ast_agg.Func.Value)
	}

	// Handle count(*) - no field reference
	if ast_agg.FieldRef == nil {
		if agg.Func != ir.AggCount {
			return nil, errutil.New(ast_agg.Pos, "only count() supports * argument")
		}
		agg.ResultType = consts.Int64Field
		agg.Nullable = false
		return agg, nil
	}

	// Resolve the field reference
	field, err := lookup.FindField(ast_agg.FieldRef)
	if err != nil {
		return nil, err
	}
	agg.Field = field

	// Determine result type based on aggregate function and field type
	switch agg.Func {
	case ir.AggCount:
		agg.ResultType = consts.Int64Field
		agg.Nullable = false
	case ir.AggSum:
		// SUM returns the same type as input
		agg.ResultType = field.Type
		agg.Nullable = true // SUM of empty set is NULL
	case ir.AggAvg:
		// AVG always returns float64
		agg.ResultType = consts.Float64Field
		agg.Nullable = true // AVG of empty set is NULL
	case ir.AggMin, ir.AggMax:
		// MIN/MAX return the same type as input
		agg.ResultType = field.Type
		agg.Nullable = true // MIN/MAX of empty set is NULL
	}

	return agg, nil
}

func transformRead(lookup *lookup, ast_read *ast.Read) (reads []*ir.Read, err error) {
	tmpl := &ir.Read{
		Suffix: transformSuffix(ast_read.Suffix),
	}

	if ast_read.Select == nil || len(ast_read.Select.Items) == 0 {
		return nil, errutil.New(ast_read.Pos, "no fields defined to select")
	}

	// Figure out which models are needed for the fields and that the field
	// references aren't repetetive. Also track if we have aggregates vs regular fields.
	selected := map[string]map[string]*ast.FieldRef{}
	in_scope := []*ir.Model{}
	hasAggregates := false
	hasRegularFields := false
	var regularFieldRefs []*ast.FieldRef // Track regular field refs for GROUP BY validation

	for _, item := range ast_read.Select.Items {
		if item.Aggregate != nil {
			hasAggregates = true
			// For aggregates with field refs, add the model to scope
			if item.Aggregate.FieldRef != nil {
				model, err := lookup.FindModel(item.Aggregate.FieldRef.ModelRef())
				if err != nil {
					return nil, err
				}
				in_scope = append(in_scope, model)

				// Track model for join validation
				modelName := item.Aggregate.FieldRef.Model.Value
				if selected[modelName] == nil {
					selected[modelName] = map[string]*ast.FieldRef{}
				}
			}

			// Transform the aggregate
			agg, err := transformAggregate(lookup, item.Aggregate)
			if err != nil {
				return nil, err
			}
			tmpl.Selectables = append(tmpl.Selectables, agg)
		} else {
			// Regular field reference
			hasRegularFields = true
			ast_fieldref := item.FieldRef

			model, err := lookup.FindModel(ast_fieldref.ModelRef())
			if err != nil {
				return nil, err
			}
			in_scope = append(in_scope, model)

			fields := selected[ast_fieldref.Model.Value]
			if fields == nil {
				fields = map[string]*ast.FieldRef{}
				selected[ast_fieldref.Model.Value] = fields
			}

			existing := fields[""]
			if existing == nil {
				existing = fields[ast_fieldref.Field.Get()]
			}
			if existing != nil {
				return nil, errutil.New(ast_fieldref.Pos,
					"field %q already selected by field %q",
					ast_fieldref, existing)
			}
			fields[ast_fieldref.Field.Get()] = ast_fieldref

			if ast_fieldref.Field.Get() == "" {
				model, err := lookup.FindModel(ast_fieldref.ModelRef())
				if err != nil {
					return nil, err
				}
				tmpl.Selectables = append(tmpl.Selectables, model)
			} else {
				field, err := lookup.FindField(ast_fieldref)
				if err != nil {
					return nil, err
				}
				tmpl.Selectables = append(tmpl.Selectables, field)
				regularFieldRefs = append(regularFieldRefs, ast_fieldref)
			}
		}
	}

	models, joins, err := transformJoins(lookup, in_scope, ast_read.Joins)
	if err != nil {
		return nil, err
	}

	tmpl.Joins = joins

	// Determine the From model
	if len(joins) > 0 {
		tmpl.From = joins[0].Left.Model
	} else if len(selected) == 1 {
		// Find the first item with a model reference
		var firstModelRef *ast.ModelRef
		for _, item := range ast_read.Select.Items {
			if item.FieldRef != nil {
				firstModelRef = item.FieldRef.ModelRef()
				break
			} else if item.Aggregate != nil && item.Aggregate.FieldRef != nil {
				firstModelRef = item.Aggregate.FieldRef.ModelRef()
				break
			}
		}
		if firstModelRef == nil {
			return nil, errutil.New(ast_read.Select.Pos,
				"cannot determine model for select")
		}
		from, err := lookup.FindModel(firstModelRef)
		if err != nil {
			return nil, err
		}
		tmpl.From = from
		models[firstModelRef.Model.Value] = firstModelRef.Pos
	} else if len(selected) == 0 {
		// All items are count(*) - need to figure out from somehow
		return nil, errutil.New(ast_read.Select.Pos,
			"cannot select only count(*) without specifying a model")
	} else {
		return nil, errutil.New(ast_read.Select.Pos,
			"cannot select from multiple models without a join")
	}

	// Make sure all of the field refs are accounted for in the set of models
	for _, item := range ast_read.Select.Items {
		var modelName string
		var pos = item.Pos
		if item.FieldRef != nil {
			modelName = item.FieldRef.Model.Value
		} else if item.Aggregate != nil && item.Aggregate.FieldRef != nil {
			modelName = item.Aggregate.FieldRef.Model.Value
			pos = item.Aggregate.FieldRef.Pos
		} else {
			continue // count(*) - no model to check
		}
		if _, ok := models[modelName]; !ok {
			return nil, errutil.New(pos,
				"cannot select from model %q; model is not joined",
				modelName)
		}
	}

	// Validate GROUP BY when mixing aggregates with regular fields
	if hasAggregates && hasRegularFields {
		if ast_read.GroupBy == nil {
			return nil, errutil.New(ast_read.Select.Pos,
				"when mixing aggregates with regular fields, GROUP BY is required")
		}

		// Build a set of GROUP BY fields for validation
		groupByFields := make(map[string]bool)
		for _, gbField := range ast_read.GroupBy.Fields.Refs {
			key := gbField.Model.Value + "." + gbField.Field.Get()
			groupByFields[key] = true
		}

		// All non-aggregated fields must appear in GROUP BY
		for _, fieldRef := range regularFieldRefs {
			key := fieldRef.Model.Value + "." + fieldRef.Field.Get()
			if !groupByFields[key] {
				return nil, errutil.New(fieldRef.Pos,
					"field %q must appear in GROUP BY clause when used with aggregates",
					fieldRef)
			}
		}
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	tmpl.Where, err = transformWheres(lookup, models, ast_read.Where)
	if err != nil {
		return nil, err
	}

	// Finalize GroupBy and make sure referenced fields are part of the select
	if ast_read.GroupBy != nil {
		fields, err := resolveFieldRefs(lookup, ast_read.GroupBy.Fields.Refs)
		if err != nil {
			return nil, err
		}
		for _, group_by_field := range ast_read.GroupBy.Fields.Refs {
			if _, ok := models[group_by_field.Model.Value]; !ok {
				return nil, errutil.New(group_by_field.Pos,
					"invalid groupby field %q; model %q is not joined",
					group_by_field, group_by_field.Model.Value)
			}
		}

		tmpl.GroupBy = &ir.GroupBy{
			Fields: fields,
		}
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if ast_read.OrderBy != nil {
		tmpl.OrderBy = new(ir.OrderBy)

		for _, entry := range ast_read.OrderBy.Entries {
			field, err := lookup.FindField(entry.Field)
			if err != nil {
				return nil, err
			}
			if _, ok := models[entry.Field.Model.Value]; !ok {
				return nil, errutil.New(entry.Field.Pos,
					"invalid orderby field %q; model %q is not joined",
					entry.Field, entry.Field.Model.Value)
			}
			tmpl.OrderBy.Entries = append(tmpl.OrderBy.Entries, &ir.OrderByEntry{
				Field:      field,
				Descending: entry.Descending.Get(),
			})
		}
	}

	// Now emit one select per view type (or one for all if unspecified)
	view := ast_read.View
	if view == nil {
		view = &ast.View{
			All: &ast.Bool{Value: true},
		}
	}

	addView := func(v ir.View) {
		read_copy := *tmpl
		read_copy.View = v
		if read_copy.Suffix == nil {
			read_copy.Suffix = DefaultReadSuffix(&read_copy)
		}
		reads = append(reads, &read_copy)
	}

	if view.All.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.All.Pos,
				"cannot limit/offset unique select")
		}
		addView(ir.All)
	}
	if view.Count.Get() {
		addView(ir.Count)
	}
	if view.Has.Get() {
		addView(ir.Has)
	}
	if view.LimitOffset.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.LimitOffset.Pos,
				"cannot use limitoffset view with distinct read")
		}
		addView(ir.LimitOffset)
	}
	if view.Paged.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.LimitOffset.Pos,
				"cannot use paged view with distinct read")
		}
		if tmpl.OrderBy != nil {
			return nil, errutil.New(view.Paged.Pos,
				"cannot page on model %q with order by",
				tmpl.From.Name)
		}
		if tmpl.GroupBy != nil {
			// Unless the primary key is part of the group by, then you can't
			// know which row the primary key would be chosen by. Not sure
			// this type of query would be useful, even if we could verify
			// that it was ok, so disabling for now.
			return nil, errutil.New(view.Paged.Pos,
				"cannot page on model %q with group by",
				tmpl.From.Name)
		}
		addView(ir.Paged)
	}
	if view.Scalar.Get() {
		addView(ir.Scalar)
	}
	if view.One.Get() {
		addView(ir.One)
	}
	if view.First.Get() {
		addView(ir.First)
	}

	return reads, nil
}
