// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlcompile

import (
	"testing"

	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqltest"
	"storj.io/dbx/testutil"
)

func TestEqual(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("fuzz identity", testEqualFuzzIdentity)
	tw.Runp("normal form", testEqualNormalForm)
}

func testEqualFuzzIdentity(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()

		if !sqlEqual(sql, sql) {
			tw.Logf("sql: %#v", sql)
			tw.Error()
		}
	}
}

func testEqualNormalForm(tw *testutil.T) {
	type normalFormTestCase struct {
		in     sqlgen.SQL
		normal bool
	}

	tests := []normalFormTestCase{
		{in: sqlgen.Literal(""), normal: true},
		{in: new(sqlgen.Condition), normal: true},
		{in: sqlgen.Literals{}, normal: true},
		{in: sqlgen.Literals{Join: "foo"}, normal: false},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("foo baz"),
				sqlgen.Literal("bif bar"),
			}},
			normal: false,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				new(sqlgen.Condition),
				sqlgen.Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("bif bar"),
				new(sqlgen.Condition),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("foo baz"),
				new(sqlgen.Condition),
				sqlgen.Literal("bif bar"),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				new(sqlgen.Condition),
				new(sqlgen.Condition),
				sqlgen.Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("bif bar"),
				new(sqlgen.Condition),
				new(sqlgen.Condition),
			}},
			normal: true,
		},
		{
			in: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("foo baz"),
				new(sqlgen.Condition),
				new(sqlgen.Condition),
				sqlgen.Literal("bif bar"),
			}},
			normal: true,
		},
		{
			in: &sqlgen.Hole{SQL: sqlgen.Literals{Join: "", SQLs: []sqlgen.SQL{
				sqlgen.Literal("foo baz"),
				sqlgen.Literal("bif bar"),
			}}},
			normal: false,
		},
	}
	for i, test := range tests {
		if got := sqlNormalForm(test.in); got != test.normal {
			tw.Errorf("%d: got:%v != exp:%v. sql:%#v",
				i, got, test.normal, test.in)
		}
	}
}
