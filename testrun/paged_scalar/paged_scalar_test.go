// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package paged_scalar

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"storj.io/dbx/testrun"

	"github.com/stretchr/testify/require"
)

type Desc struct {
	Name  string
	Model reflect.Type
	Field reflect.Value
	Cont  reflect.Type
	Auto  bool
}

var descs = []Desc{
	{"Blob", typ(DataBlob{}), val(DataBlob_Id), typ(Paged_DataBlob_Continuation{}), false},
	{"Date", typ(DataDate{}), val(DataDate_Id), typ(Paged_DataDate_Continuation{}), false},
	{"Float", typ(DataFloat{}), val(DataFloat_Id), typ(Paged_DataFloat_Continuation{}), false},
	{"Float64", typ(DataFloat64{}), val(DataFloat64_Id), typ(Paged_DataFloat64_Continuation{}), false},
	{"Int", typ(DataInt{}), val(DataInt_Id), typ(Paged_DataInt_Continuation{}), false},
	{"Int64", typ(DataInt64{}), val(DataInt64_Id), typ(Paged_DataInt64_Continuation{}), false},
	{"Serial", typ(DataSerial{}), val(DataSerial_Id), typ(Paged_DataSerial_Continuation{}), true},
	{"Serial64", typ(DataSerial64{}), val(DataSerial64_Id), typ(Paged_DataSerial64_Continuation{}), true},
	{"Text", typ(DataText{}), val(DataText_Id), typ(Paged_DataText_Continuation{}), false},
	{"Timestamp", typ(DataTimestamp{}), val(DataTimestamp_Id), typ(Paged_DataTimestamp_Continuation{}), false},
	{"Uint", typ(DataUint{}), val(DataUint_Id), typ(Paged_DataUint_Continuation{}), false},
	{"Uint64", typ(DataUint64{}), val(DataUint64_Id), typ(Paged_DataUint64_Continuation{}), false},
	{"Utimestamp", typ(DataUtimestamp{}), val(DataUtimestamp_Id), typ(Paged_DataUtimestamp_Continuation{}), false},
}

func TestPagedScalar(t *testing.T) {
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		if testrun.IsSpanner[*DB](db.DB) {
			t.Skip("TODO(spanner): column data_jsons.id has type JSON, but is part of the primary key")
		}

		testrun.RecreateSchema(t, db)

		for _, desc := range descs {
			t.Run(strings.ToLower(desc.Name), func(t *testing.T) {
				runDesc(t, db, desc)
			})
		}
	})
}

func runDesc(t *testing.T, db *DB, desc Desc) {
	ctx := context.Background()
	create := val(db).MethodByName(fmt.Sprintf("Create_Data%s", desc.Name))
	paged := val(db).MethodByName(fmt.Sprintf("Paged_Data%s", desc.Name))
	id := reflect.Zero(desc.Field.Type().In(0)).Interface()
	field := func(in any) reflect.Value { return desc.Field.Call(vs{val(in)})[0] }

	// create 10 models
	for i := 0; i < 10; i++ {
		id = next(id)
		args := vs{val(ctx)}
		if !desc.Auto {
			args = append(args, field(id))
		}

		require.True(t, create.Call(args)[1].IsNil())
	}

	// paged iterate over it 2 at a time
	count := 0
	cont := reflect.Zero(reflect.PtrTo(desc.Cont))
	for j := 0; j < 6; j++ {
		out := paged.Call(vs{val(ctx), val(2), cont})
		require.True(t, out[2].IsNil())
		count += out[0].Len()
		cont = out[1]

		if !out[1].IsNil() {
			continue
		}
		if count != 10 {
			require.Fail(t, "didn't iterate all of them")
		}
		return
	}
	require.Fail(t, "too many iterations")
}

func next(in any) any {
	switch in := in.(type) {
	case []byte:
		return append(in, 'b')
	case time.Time:
		if in.IsZero() {
			return time.Now()
		}
		// add one day so there can be a primary key of type date
		return in.Add(time.Hour * 24)
	case float32:
		return in + 1
	case float64:
		return in + 1
	case int:
		return in + 1
	case int64:
		return in + 1
	case string:
		return string(append([]byte(in), 's'))
	case uint:
		return in + 1
	case uint64:
		return in + 1
	}
	panic(in)
}

var (
	typ = reflect.TypeOf
	val = reflect.ValueOf
)

type vs = []reflect.Value
