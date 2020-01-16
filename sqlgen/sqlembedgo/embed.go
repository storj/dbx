// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlembedgo

import (
	"bytes"
	"fmt"

	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlbundle"
)

type Condition struct {
	Name       string
	Expression string
}

type Hole struct {
	Name       string
	Expression string
}

type Info struct {
	Expression string
	Conditions []Condition
	Holes      []Hole
}

func Embed(prefix string, sql sqlgen.SQL) Info {
	return Info{
		Expression: quote(prefix, sql, true),
		Conditions: gatherConditions(prefix, sql),
		Holes:      gatherHoles(prefix, sql),
	}
}

//
// quoting
//

func quote(prefix string, sql sqlgen.SQL, literal bool) string {
	switch sql := sql.(type) {
	case sqlgen.Literal:
		return quoteLiteral(sql)

	case *sqlgen.Condition:
		return quoteCondition(prefix, sql, literal)

	case *sqlgen.Hole:
		return quoteHole(prefix, sql, literal)

	case sqlgen.Literals:
		return quoteLiterals(prefix, sql)

	default:
		panic("unhandled sql type")
	}
}

func quoteLiteral(sql sqlgen.Literal) string {
	const format = "%[1]sLiteral(%[2]q)"

	return fmt.Sprintf(format, sqlbundle.Prefix, string(sql))
}

func quoteCondition(prefix string, sql *sqlgen.Condition, literal bool) string {
	if !literal {
		return prefix + sql.Name
	}

	const format = "&%[1]sCondition{Left:%q, Equal:%t, Right: %q, Null:true}"

	return fmt.Sprintf(format, sqlbundle.Prefix, sql.Left, sql.Equal, sql.Right)
}

func quoteHole(prefix string, sql *sqlgen.Hole, literal bool) string {
	if !literal {
		return prefix + sql.Name
	}

	if sql.SQL == nil {
		const format = "&%[1]sHole{}"

		return fmt.Sprintf(format, sqlbundle.Prefix)
	}

	const format = "&%[1]sHole{SQL: %[2]s}"

	return fmt.Sprintf(format, sqlbundle.Prefix, quote(prefix, sql.SQL, false))
}

func quoteLiterals(prefix string, sql sqlgen.Literals) string {
	const format = "%[1]sLiterals{Join:%[2]q,SQLs:[]%[1]sSQL{"

	var expr bytes.Buffer
	fmt.Fprintf(&expr, format, sqlbundle.Prefix, sql.Join)

	first := true
	for _, sql := range sql.SQLs {
		if !first {
			expr.WriteString(",")
		}
		first = false
		fmt.Fprint(&expr, quote(prefix, sql, false))
	}
	expr.WriteString("}}")

	return expr.String()
}

//
// gathering conditions
//

func gatherConditions(prefix string, sql sqlgen.SQL) []Condition {
	switch sql := sql.(type) {
	case sqlgen.Literal:
		return nil

	case *sqlgen.Hole:
		return gatherConditions(prefix, sql.SQL)

	case *sqlgen.Condition:
		return []Condition{{
			Name:       prefix + sql.Name,
			Expression: quoteCondition(prefix, sql, true),
		}}

	case sqlgen.Literals:
		var conds []Condition
		for _, sql := range sql.SQLs {
			conds = append(conds, gatherConditions(prefix, sql)...)
		}
		return conds

	case nil:
		return nil

	default:
		panic("unhandled sql type")
	}
}

//
// gathering holes
//

func gatherHoles(prefix string, sql sqlgen.SQL) []Hole {
	switch sql := sql.(type) {
	case sqlgen.Literal, *sqlgen.Condition:
		return nil

	case *sqlgen.Hole:
		return append(gatherHoles(prefix, sql.SQL), Hole{
			Name:       prefix + sql.Name,
			Expression: quoteHole(prefix, sql, true),
		})

	case sqlgen.Literals:
		var holes []Hole
		for _, sql := range sql.SQLs {
			holes = append(holes, gatherHoles(prefix, sql)...)
		}
		return holes

	case nil:
		return nil

	default:
		panic("unhandled sql type")
	}
}
