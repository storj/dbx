// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

import (
	"strings"
)

type SQL interface {
	Render() string

	private()
}

type Dialect interface {
	// Rebind gives the opportunity to rewrite provided SQL into a SQL dialect.
	Rebind(sql string) string
}

type RenderOp int

const (
	NoFlatten RenderOp = iota
	NoTerminate
)

func Render(dialect Dialect, sql SQL, ops ...RenderOp) string {
	out := sql.Render()

	flatten := true
	terminate := true
	for _, op := range ops {
		switch op {
		case NoFlatten:
			flatten = false
		case NoTerminate:
			terminate = false
		}
	}

	if flatten {
		out = flattenSQL(out)
	}
	if terminate {
		out += ";"
	}

	return dialect.Rebind(out)
}

func flattenSQL(x string) string {
	// trim whitespace from beginning and end
	s, e := 0, len(x)-1
	for s < len(x) && (x[s] == ' ' || x[s] == '\t' || x[s] == '\n') {
		s++
	}
	for s <= e && (x[e] == ' ' || x[e] == '\t' || x[e] == '\n') {
		e--
	}
	if s > e {
		return ""
	}
	x = x[s : e+1]

	// check for whitespace that needs fixing
	wasSpace := false
	for i := 0; i < len(x); i++ {
		r := x[i]
		justSpace := r == ' '
		if (wasSpace && justSpace) || r == '\t' || r == '\n' {
			// whitespace detected, start writing a new string
			var result strings.Builder
			result.Grow(len(x))
			if wasSpace {
				result.WriteString(x[:i-1])
			} else {
				result.WriteString(x[:i])
			}
			for p := i; p < len(x); p++ {
				for p < len(x) && (x[p] == ' ' || x[p] == '\t' || x[p] == '\n') {
					p++
				}
				result.WriteByte(' ')

				start := p
				for p < len(x) && !(x[p] == ' ' || x[p] == '\t' || x[p] == '\n') {
					p++
				}
				result.WriteString(x[start:p])
			}

			return result.String()
		}
		wasSpace = justSpace
	}

	// no problematic whitespace found
	return x
}
