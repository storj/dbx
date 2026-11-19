// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInFilterRejects(t *testing.T) {
	const model = `model foo (
	key id
	field id   int
	field name text
	field data json ( nullable )
)
`
	for _, tc := range []struct {
		name   string
		source string
		errMsg string
	}{
		{
			name:   "literal on the right",
			source: model + `read all ( select foo, where foo.id in 5 )`,
			errMsg: `the right side of an "in" clause must be a placeholder`,
		},
		{
			name:   "function call on the left",
			source: model + `read all ( select foo, where lower(foo.name) in ? )`,
			errMsg: `the left side of an "in" clause must be a field`,
		},
		{
			name:   "json field",
			source: model + `read all ( select foo, where foo.data in ? )`,
			errMsg: `"in" clauses are not supported on json fields`,
		},
		{
			name: "index condition",
			source: `model foo (
	key id
	index ( fields name, where foo.id in ? )
	field id   int
	field name text
)`,
			errMsg: `"in" clauses are not supported in index conditions`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "test.dbx")
			require.NoError(t, os.WriteFile(file, []byte(tc.source), 0o644))

			err := newGlobal().golangCmd("", []string{"pgx"}, "", false, []string{file}, dir)
			require.ErrorContains(t, err, tc.errMsg)
		})
	}
}
