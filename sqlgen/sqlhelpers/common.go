// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlhelpers

import (
	"fmt"

	"storj.io/dbx/sqlgen"
)

// ls is the basic primitive for constructing larger SQLs. The first argument
// may be nil, and the result is a Literals.
func ls(sql sqlgen.SQL, join string, sqls ...sqlgen.SQL) sqlgen.SQL {
	var joined []sqlgen.SQL
	if sql != nil {
		joined = append(joined, sql)
	}
	joined = append(joined, sqls...)

	return sqlgen.Literals{Join: join, SQLs: joined}
}

// Placeholder is a placeholder literal
const Placeholder = sqlgen.Literal("?")

// L constructs a Literal
func L(sql string) sqlgen.SQL {
	return sqlgen.Literal(sql)
}

// Lf constructs a literal from a format string
func Lf(sqlf string, args ...interface{}) sqlgen.SQL {
	return sqlgen.Literal(fmt.Sprintf(sqlf, args...))
}

// J constructs a SQL by joining the given sqls with the string.
func J(join string, sqls ...sqlgen.SQL) sqlgen.SQL {
	return ls(nil, join, sqls...)
}

// Strings turns a slice of strings into a slice of Literal.
func Strings(parts []string) (out []sqlgen.SQL) {
	for _, part := range parts {
		out = append(out, sqlgen.Literal(part))
	}
	return out
}

// Placeholders returns a slice of placeholder literals of the right size.
func Placeholders(n int) (out []sqlgen.SQL) {
	for i := 0; i < n; i++ {
		out = append(out, Placeholder)
	}
	return out
}

func Hole(name string) *sqlgen.Hole {
	return &sqlgen.Hole{Name: name}
}

//
// Builder constructs larger SQL statements by joining in pieces with spaces.
//

type Builder struct {
	sql sqlgen.SQL
}

func Build(sql sqlgen.SQL) *Builder {
	return &Builder{
		sql: sql,
	}
}

func (b *Builder) Add(sqls ...sqlgen.SQL) {
	b.sql = ls(b.sql, " ", sqls...)
}

func (b *Builder) SQL() sqlgen.SQL {
	return b.sql
}
