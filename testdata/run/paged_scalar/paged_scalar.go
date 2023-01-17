package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"
)

type Desc struct {
	Name  string
	Model reflect.Type
	Field reflect.Value
	Cont  reflect.Type
	Auto  bool
}

var descs = []Desc{
	{"Blob", typ(Blob{}), val(Blob_Id), typ(Paged_Blob_Continuation{}), false},
	{"Date", typ(Date{}), val(Date_Id), typ(Paged_Date_Continuation{}), false},
	{"Float", typ(Float{}), val(Float_Id), typ(Paged_Float_Continuation{}), false},
	{"Float64", typ(Float64{}), val(Float64_Id), typ(Paged_Float64_Continuation{}), false},
	{"Int", typ(Int{}), val(Int_Id), typ(Paged_Int_Continuation{}), false},
	{"Int64", typ(Int64{}), val(Int64_Id), typ(Paged_Int64_Continuation{}), false},
	{"Serial", typ(Serial{}), val(Serial_Id), typ(Paged_Serial_Continuation{}), true},
	{"Serial64", typ(Serial64{}), val(Serial64_Id), typ(Paged_Serial64_Continuation{}), true},
	{"Text", typ(Text{}), val(Text_Id), typ(Paged_Text_Continuation{}), false},
	{"Timestamp", typ(Timestamp{}), val(Timestamp_Id), typ(Paged_Timestamp_Continuation{}), false},
	{"Uint", typ(Uint{}), val(Uint_Id), typ(Paged_Uint_Continuation{}), false},
	{"Uint64", typ(Uint64{}), val(Uint64_Id), typ(Paged_Uint64_Continuation{}), false},
	{"Utimestamp", typ(Utimestamp{}), val(Utimestamp_Id), typ(Paged_Utimestamp_Continuation{}), false},
}

func main() {
	sqliteDb, err := Open("sqlite3", ":memory:")
	erre(err)
	runDb(sqliteDb)
	err = sqliteDb.Close()
	erre(err)

	dsn := os.Getenv("STORJ_TEST_POSTGRES")
	if dsn == "" {
		println("Skipping pq and pgx tests because environment variable STORJ_TEST_POSTGRES is not set")
		return
	}

	pqDb, err := Open("postgres", dsn)
	erre(err)
	_, err = pqDb.Exec("DROP TABLE IF EXISTS _blobs, _dates, _floats, _float64s, _ints, _int64s, _jsons, _serials, _serial64s, _texts, _timestamps, _uints, _uint64s, _utimestamps")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS _blobs, _dates, _floats, _float64s, _ints, _int64s, _jsons, _serials, _serial64s, _texts, _timestamps, _uints, _uint64s, _utimestamps")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	for _, desc := range descs {
		runDesc(db, desc)
	}
}

func runDesc(db *DB, desc Desc) {
	create := val(db).MethodByName(fmt.Sprintf("Create_%s", desc.Name))
	paged := val(db).MethodByName(fmt.Sprintf("Paged_%s", desc.Name))
	id := reflect.Zero(desc.Field.Type().In(0)).Interface()
	field := func(in interface{}) reflect.Value { return desc.Field.Call(vs{val(in)})[0] }

	// create 10 models
	for i := 0; i < 10; i++ {
		id = next(id)
		args := vs{val(ctx)}
		if !desc.Auto {
			args = append(args, field(id))
		}
		err_r(create.Call(args)[1])
	}

	// paged iterate over it 2 at a time
	count := 0
	cont := reflect.Zero(reflect.PtrTo(desc.Cont))
	for j := 0; j < 6; j++ {
		out := paged.Call(vs{val(ctx), val(2), cont})
		err_r(out[2])
		count += out[0].Len()
		cont = out[1]

		if !out[1].IsNil() {
			continue
		}
		if count != 10 {
			panic("didn't iterate all of them")
		}
		return
	}
	panic("too many iterations")
}

func next(in interface{}) interface{} {
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

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

func err_r(err reflect.Value) {
	if !err.IsNil() {
		panic(err.Interface().(error))
	}
}

var ctx = context.Background()
var typ = reflect.TypeOf
var val = reflect.ValueOf

type vs = []reflect.Value
