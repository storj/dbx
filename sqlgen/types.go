// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

import (
	"bytes"
)

type Literal string

func (Literal) private() {}

func (l Literal) Render() string { return string(l) }

type Literals struct {
	Join string
	SQLs []SQL
}

func (Literals) private() {}

func (l Literals) Render() string {
	var out bytes.Buffer

	first := true
	for _, sql := range l.SQLs {
		if sql == nil {
			continue
		}
		if !first {
			out.WriteString(l.Join)
		}
		first = false
		out.WriteString(sql.Render())
	}

	return out.String()
}

type Condition struct {
	// set at compile/embed time
	Name  string
	Left  string
	Equal bool
	Right string

	// set at runtime
	Null bool
}

func (*Condition) private() {}

func (c *Condition) Render() string {
	// TODO(jeff): maybe check if we can use placeholders instead of the
	// literal null: this would make the templates easier.

	switch {
	case c.Equal && c.Null:
		return c.Left + " is null"
	case c.Equal && !c.Null:
		return c.Left + " = " + c.Right
	case !c.Equal && c.Null:
		return c.Left + " is not null"
	case !c.Equal && !c.Null:
		return c.Left + " != " + c.Right
	default:
		panic("unhandled case")
	}
}

type Hole struct {
	// set at compiile/embed time
	Name string

	// set at runtime or possibly embed time
	SQL SQL
}

func (*Hole) private() {}

func (h *Hole) Render() string {
	if h.SQL == nil {
		return ""
	}
	return h.SQL.Render()
}
