// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlembedgo

import (
	"go/parser"
	"testing"

	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	"storj.io/dbx/sqlgen/sqltest"
	"storj.io/dbx/testutil"
)

func TestGolang(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("basic types", testGolangBasicTypes)
	tw.Runp("fuzz", testGolangFuzz)
}

func testGolangBasicTypes(tw *testutil.T) {
	tests := []sqlgen.SQL{
		sqlgen.Literal(""),
		sqlgen.Literal("foo bar sql"),
		sqlgen.Literal("`"),
		sqlgen.Literal(`"`),

		&sqlgen.Condition{Name: "foo"},

		sqlgen.Literals{},
		sqlgen.Literals{Join: "foo"},
		sqlgen.Literals{Join: "`"},
		sqlgen.Literals{Join: `"`},
		sqlgen.Literals{Join: "bar", SQLs: []sqlgen.SQL{
			sqlgen.Literal("foo baz"),
			sqlgen.Literal("another"),
			&sqlgen.Condition{Name: "foo"},
		}},

		&sqlgen.Hole{Name: "foo", SQL: sqlgen.Literal("foo bar sql")},

		&sqlgen.Hole{Name: "foo", SQL: &sqlgen.Hole{Name: "bar"}},
	}
	for i, test := range tests {
		info := Embed("prefix_", test)
		if _, err := parser.ParseExpr(info.Expression); err != nil {
			tw.Errorf("%d: %+v but got error: %v", i, info, err)
		}
	}
}

func testGolangFuzz(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		compiled := sqlcompile.Compile(sql)
		info := Embed("prefix_", compiled)

		if _, err := parser.ParseExpr(info.Expression); err != nil {
			tw.Logf("sql:      %#v", sql)
			tw.Logf("compiled: %#v", compiled)
			tw.Logf("info:     %+v", info)
			tw.Logf("err:      %v", err)
			tw.Error()
		}
	}
}
