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

func main() {
	dsn := os.Getenv("STORJ_TEST_POSTGRES")
	if dsn == "" {
		println("Skipping array test because environment variable STORJ_TEST_POSTGRES is not set")
		return
	}

	return

	// Need to use PostgreSQL because arrays aren't supported in SQLite.
	db, err := Open("pgx", dsn)
	erre(err)
	defer db.Close()

	_, err = db.DB.Exec("DROP TABLE IF EXISTS frummels")
	erre(err)

	_, err = db.Exec(db.Schema())
	erre(err)

	frummel, err := db.Create_Frummel(ctx, Frummel_Id(42), Frummel_Items([]int64{6, 7, 8}))
	erre(err)
	eq(frummel.Id, 42)
	eqarray(frummel.Items, []int64{6, 7, 8})

	frummel, err = db.Update_Frummel_By_Id(ctx,
		Frummel_Id(42),
		Frummel_Update_Fields{
			Items: Frummel_Items([]int64{9, 10, 11}),
		},
	)
	erre(err)
	eq(frummel.Id, 42)
	eqarray(frummel.Items, []int64{9, 10, 11})

	frummels, err := db.All_Frummel(ctx)
	erre(err)
	eq(frummels[0].Id, 42)
	eqarray(frummels[0].Items, []int64{9, 10, 11})

	_, err = db.DB.Exec("DROP TABLE IF EXISTS frummels")
	erre(err)
}
