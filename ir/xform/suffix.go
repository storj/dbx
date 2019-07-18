// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package xform

import "storj.io/dbx/ast"

func transformSuffix(suffix *ast.Suffix) []string {
	var parts []string
	if suffix == nil {
		return parts
	}
	for _, part := range suffix.Parts {
		parts = append(parts, part.Value)
	}
	return parts
}
