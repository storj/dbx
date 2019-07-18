// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlcompile

import (
	"testing"

	"storj.io/dbx/sqlgen/sqltest"
	"storj.io/dbx/testutil"
)

func TestCompile(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("fuzz render", testCompileFuzzRender)
	tw.Runp("idempotent", testCompileIdempotent)
	tw.Runp("fuzz normal form", testCompileFuzzNormalForm)
}

func testCompileFuzzRender(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		compiled := Compile(sql)
		exp := sql.Render()
		got := compiled.Render()

		if exp != got {
			tw.Logf("sql:      %#v", sql)
			tw.Logf("compiled: %#v", compiled)
			tw.Logf("exp:      %q", exp)
			tw.Logf("got:      %q", got)
			tw.Error()
		}
	}
}

func testCompileIdempotent(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		first := Compile(sql)
		second := Compile(first)

		if !sqlEqual(first, second) {
			tw.Logf("sql:    %#v", sql)
			tw.Logf("first:  %#v", first)
			tw.Logf("second: %#v", second)
			tw.Error()
		}
	}
}

func testCompileFuzzNormalForm(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		compiled := Compile(sql)

		if !sqlNormalForm(compiled) {
			tw.Logf("sql:      %#v", sql)
			tw.Logf("compiled: %#v", compiled)
			tw.Error()
		}
	}
}
