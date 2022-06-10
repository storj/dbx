// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package consts

import "fmt"

type JoinType int

const (
	InnerJoin JoinType = iota
)

type Operator string

const (
	LT   Operator = "<"
	LE   Operator = "<="
	GT   Operator = ">"
	GE   Operator = ">="
	EQ   Operator = "="
	NE   Operator = "!="
	Like Operator = "like"
)

func (o Operator) Suffix() string {
	switch o {
	case LT:
		return "less"
	case LE:
		return "less_or_equal"
	case GT:
		return "greater"
	case GE:
		return "greater_or_equal"
	case EQ:
		return "equal"
	case NE:
		return "not"
	case Like:
		return "like"
	default:
		panic(fmt.Sprintf("unhandled operation %q", o))
	}
}

type FieldType int

const (
	SerialField FieldType = iota
	Serial64Field
	IntField
	Int64Field
	UintField
	Uint64Field
	FloatField
	Float64Field
	TextField
	BoolField
	TimestampField
	TimestampUTCField
	BlobField
	DateField
	JsonField
)

func (f FieldType) String() string {
	switch f {
	case SerialField:
		return "serial"
	case Serial64Field:
		return "serial64"
	case IntField:
		return "int"
	case Int64Field:
		return "int64"
	case UintField:
		return "uint"
	case Uint64Field:
		return "uint64"
	case FloatField:
		return "float"
	case Float64Field:
		return "float64"
	case TextField:
		return "text"
	case BoolField:
		return "bool"
	case TimestampField:
		return "timestamp"
	case TimestampUTCField:
		return "utimestamp"
	case BlobField:
		return "blob"
	case DateField:
		return "date"
	case JsonField:
		return "json"
	default:
		return "<UNKNOWN-FIELD>"
	}
}

func (f FieldType) AsLink() FieldType {
	switch f {
	case SerialField:
		return IntField
	case Serial64Field:
		return Int64Field
	default:
		return f
	}
}

type RelationKind int

const (
	SetNull RelationKind = iota
	Cascade
	Restrict
)
