// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

import (
	"testing"

	"storj.io/dbx/testutil"
)

func TestTypes(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("render", testTypesRender)
}

func testTypesRender(tw *testutil.T) {
	type renderTestCase struct {
		in  SQL
		out string
	}

	tests := []renderTestCase{
		{in: Literal(""), out: ""},
		{in: Literal("foo bar sql"), out: "foo bar sql"},
		{in: Literal("`"), out: "`"},
		{in: Literal(`"`), out: "\""},

		{in: Literals{}, out: ""},
		{in: Literals{Join: "foo"}, out: ""},
		{in: Literals{Join: "`"}, out: ""},
		{in: Literals{Join: `"`}, out: ""},
		{
			in: Literals{Join: " bar ", SQLs: []SQL{
				Literal("foo baz"),
				Literal("another"),
			}},
			out: "foo baz bar another",
		},
		{
			in: Literals{Join: " bar ", SQLs: []SQL{
				Literal("inside first"),
				Literals{},
				Literal("inside second"),
			}},
			out: "inside first bar  bar inside second",
		},
		{
			in: Literals{Join: " recursive ", SQLs: []SQL{
				Literals{Join: " bif ", SQLs: []SQL{
					Literals{},
					Literal("inside"),
				}},
				Literal("outside"),
			}},
			out: " bif inside recursive outside",
		},

		{in: &Condition{Equal: false, Null: false}, out: " != "},
		{in: &Condition{Equal: false, Null: true}, out: " is not null"},
		{in: &Condition{Equal: true, Null: false}, out: " = "},
		{in: &Condition{Equal: true, Null: true}, out: " is null"},
		{
			in:  &Condition{Left: "f", Right: "?", Equal: false, Null: false},
			out: "f != ?",
		},
		{
			in:  &Condition{Left: "f", Right: "?", Equal: false, Null: true},
			out: "f is not null",
		},
		{
			in:  &Condition{Left: "f", Right: "?", Equal: true, Null: false},
			out: "f = ?",
		},
		{
			in:  &Condition{Left: "f", Right: "?", Equal: true, Null: true},
			out: "f is null",
		},

		{in: &Hole{SQL: Literal("hello")}, out: "hello"},
	}
	for i, test := range tests {
		if got := test.in.Render(); got != test.out {
			tw.Errorf("%d: %q != %q", i, got, test.out)
		}
	}
}
