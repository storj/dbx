package main

import (
	"context"
	"os"
	"reflect"
)

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

func eq(a, b interface{}) {
	if a != b {
		panic("invalid")
	}
}

func eqarray(a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		panic("invalid")
	}
}

var ctx = context.Background()

// Test is only run on PostgreSQL and Cockroach because arrays aren't supported in SQLite.
func main() {
	dsnPostgres := os.Getenv("STORJ_TEST_POSTGRES")
	if dsnPostgres == "" {
		println("Skipping array test because environment variable STORJ_TEST_POSTGRES is not set")
		return
	}

	dbPgx, err := Open("pgx", dsnPostgres)
	erre(err)
	_, err = dbPgx.DB.Exec("DROP TABLE IF EXISTS frummels")
	erre(err)
	runDb(dbPgx)
	err = dbPgx.Close()
	erre(err)

	dsnCockroach := os.Getenv("STORJ_TEST_COCKROACH")
	if dsnCockroach == "" {
		println("Skipping array test because environment variable STORJ_TEST_COCKROACH is not set")
		return
	}

	db, err := Open("pgxcockroach", dsnCockroach)
	erre(err)
	_, err = db.DB.Exec("DROP TABLE IF EXISTS frummels")
	erre(err)
	runDb(db)
	err = db.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	frummel, err := db.Create_Frummel(ctx, Frummel_Id(42), Frummel_Items([]int{6, 7, 8}))
	erre(err)
	eq(frummel.Id, 42)
	eqarray(frummel.Items, []int{6, 7, 8})

	frummel, err = db.Update_Frummel_By_Id(ctx,
		Frummel_Id(42),
		Frummel_Update_Fields{
			Items: Frummel_Items([]int{9, 10, 11}),
		},
	)
	erre(err)
	eq(frummel.Id, 42)
	eqarray(frummel.Items, []int{9, 10, 11})

	frummels, err := db.All_Frummel(ctx)
	erre(err)
	eq(frummels[0].Id, 42)
	eqarray(frummels[0].Items, []int{9, 10, 11})

	_, err = db.DB.Exec("DROP TABLE IF EXISTS frummels")
	erre(err)
}
