// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

import (
	"testing"

	"storj.io/dbx/testutil"
)

type rebindTestCase struct {
	in  string
	out string
}

func testDialectsRebind(tw *testutil.T, dialect Dialect,
	tests []rebindTestCase) {

	for i, test := range tests {
		if got := dialect.Rebind(test.in); got != test.out {
			tw.Errorf("%d: %q != %q", i, got, test.out)
		}
	}
}

func TestDialects(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("postgres", testDialectsPostgres)
	tw.Runp("sqlite3", testDialectsSQLite3)
	tw.Runp("cockroach", testDialectsCockroach)
}

func testDialectsPostgres(tw *testutil.T) {
	tw.Runp("rebind", testDialectsPostgresRebind)
}

func testDialectsPostgresRebind(tw *testutil.T) {
	testDialectsRebind(tw, postgres{}, []rebindTestCase{
		{in: "", out: ""},
		{in: "? foo bar ? baz", out: "$1 foo bar $2 baz"},
		{in: "? ? ?", out: "$1 $2 $3"},
		{in: "?,?,?,?,?,?,?,?,?,?", out: "$1,$2,$3,$4,$5,$6,$7,$8,$9,$10"},
		{
			in:  "select ?like?, 'question?''?' as \"quotedcolname?\"\"?\", -- comment?\nbar from table\nwhere x=-? and y='-?'",
			out: "select $1like$2, 'question?''?' as \"quotedcolname?\"\"?\", -- comment?\nbar from table\nwhere x=-$3 and y='-?'",
		},
	})
}

func testDialectsSQLite3(tw *testutil.T) {
	tw.Runp("rebind", testDialectsSQLite3Rebind)
}

func testDialectsSQLite3Rebind(tw *testutil.T) {
	testDialectsRebind(tw, sqlite3{}, []rebindTestCase{
		{in: "", out: ""},
		{in: "? foo bar ? baz", out: "? foo bar ? baz"},
		{in: "? ? ?", out: "? ? ?"},
	})
}

func testDialectsCockroach(tw *testutil.T) {
	tw.Runp("rebind", testDialectsCockroachRebind)
}

func testDialectsCockroachRebind(tw *testutil.T) {
	testDialectsRebind(tw, cockroach{}, []rebindTestCase{
		{in: "", out: ""},
		{in: "? foo bar ? baz", out: "$1 foo bar $2 baz"},
		{in: "? ? ?", out: "$1 $2 $3"},
		{in: "?,?,?,?,?,?,?,?,?,?", out: "$1,$2,$3,$4,$5,$6,$7,$8,$9,$10"},
		{
			in:  "select ?like?, 'question?''?' as \"quotedcolname?\"\"?\", -- comment?\nbar from table\nwhere x=-? and y='-?'",
			out: "select $1like$2, 'question?''?' as \"quotedcolname?\"\"?\", -- comment?\nbar from table\nwhere x=-$3 and y='-?'",
		},
	})
}
