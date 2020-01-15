// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlcompile

import "storj.io/dbx/sqlgen"

func sqlEqual(a, b sqlgen.SQL) bool {
	switch a := a.(type) {
	case sqlgen.Literal:
		if b, ok := b.(sqlgen.Literal); ok {
			return a == b
		}
		return false

	case sqlgen.Literals:
		if b, ok := b.(sqlgen.Literals); ok {
			return a.Join == b.Join && sqlsEqual(a.SQLs, b.SQLs)
		}
		return false

	case *sqlgen.Condition:
		if b, ok := b.(*sqlgen.Condition); ok {
			return a == b // pointer equality is correct
		}
		return false

	case *sqlgen.Hole:
		if b, ok := b.(*sqlgen.Hole); ok {
			return a == b // pointer equality is correct
		}
		return false

	default:
		panic("unhandled sql type")
	}
}

func sqlsEqual(as, bs []sqlgen.SQL) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if !sqlEqual(as[i], bs[i]) {
			return false
		}
	}
	return true
}

func sqlNormalForm(sql sqlgen.SQL) bool {
	switch sql := sql.(type) {
	case sqlgen.Literal, *sqlgen.Condition:
		return true

	case *sqlgen.Hole:
		return sql.SQL == nil || sqlNormalForm(sql.SQL)

	case sqlgen.Literals:
		if sql.Join != "" {
			return false
		}

		// only allow Hole, Condition and Literal but disallow two Literal in
		// a row.

		last := ""

		for _, sql := range sql.SQLs {
			switch sql.(type) {
			case *sqlgen.Condition:
				last = "condition"

			case *sqlgen.Hole:
				last = "hole"

			case sqlgen.Literal:
				if last == "literal" {
					return false
				}
				last = "literal"

			case sqlgen.Literals:
				return false

			default:
				panic("unhandled sql type")
			}
		}

		return true

	default:
		panic("unhandled sql type")
	}
}
