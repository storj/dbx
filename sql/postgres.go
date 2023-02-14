// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sql

import (
	"fmt"
	"strconv"
	"strings"

	"storj.io/dbx/consts"
	"storj.io/dbx/ir"
)

type postgres struct {
}

func Postgres() Dialect {
	return &postgres{}
}

func (p *postgres) Name() string {
	return "postgres"
}

func (p *postgres) Features() Features {
	return Features{
		Returning:           true,
		PositionalArguments: true,
		NoLimitToken:        "ALL",
		ReplaceStyle:        ReplaceStyle_OnConflictUpdate,
	}
}

func (p *postgres) RowId() string {
	return ""
}

func (p *postgres) ColumnType(field *ir.Field) string {
	switch field.Type {
	case consts.SerialField:
		return "serial"
	case consts.Serial64Field:
		return "bigserial"
	case consts.IntField:
		return "integer"
	case consts.Int64Field:
		return "bigint"
	case consts.UintField:
		return "integer"
	case consts.Uint64Field:
		return "bigint"
	case consts.FloatField:
		return "real"
	case consts.Float64Field:
		return "double precision"
	case consts.TextField:
		if field.Length > 0 {
			return fmt.Sprintf("varchar(%d)", field.Length)
		}
		return "text"
	case consts.BoolField:
		return "boolean"
	case consts.TimestampField:
		return "timestamp with time zone"
	case consts.TimestampUTCField:
		return "timestamp"
	case consts.BlobField:
		return "bytea"
	case consts.DateField:
		return "date"
	case consts.JsonField:
		return "jsonb"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (s *postgres) Scanner(dest interface{}) interface{} {
	panic("unused function to maintain interface compatibility with sqlgen.Dialect")
}

func (p *postgres) Rebind(sql string) string {
	type sqlParseState int
	const (
		sqlParseStart sqlParseState = iota
		sqlParseInStringLiteral
		sqlParseInQuotedIdentifier
		sqlParseInComment
	)

	out := make([]byte, 0, len(sql)+10)

	j := 1
	state := sqlParseStart
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch state {
		case sqlParseStart:
			switch ch {
			case '?':
				out = append(out, '$')
				out = append(out, strconv.Itoa(j)...)
				state = sqlParseStart
				j++
				continue
			case '-':
				if i+1 < len(sql) && sql[i+1] == '-' {
					state = sqlParseInComment
				}
			case '"':
				state = sqlParseInQuotedIdentifier
			case '\'':
				state = sqlParseInStringLiteral
			}
		case sqlParseInStringLiteral:
			if ch == '\'' {
				state = sqlParseStart
			}
		case sqlParseInQuotedIdentifier:
			if ch == '"' {
				state = sqlParseStart
			}
		case sqlParseInComment:
			if ch == '\n' {
				state = sqlParseStart
			}
		}
		out = append(out, ch)
	}

	return string(out)
}

func (p *postgres) ArgumentPrefix() string { return "$" }

var postgresEscaper = strings.NewReplacer(
	`'`, `''`,
)

func (p *postgres) EscapeString(s string) string {
	return postgresEscaper.Replace(s)
}

func (p *postgres) BoolLit(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func (p *postgres) DefaultLit(v string) string {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return `'` + p.EscapeString(v[1:len(v)-1]) + `'`
	}
	return v
}
