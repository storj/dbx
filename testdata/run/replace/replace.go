package main

import (
	"context"
	"os"
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

var ctx = context.Background()

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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS kvs")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS kvs")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	err = db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val0"))
	erre(err)
	row, err := db.Get_Kv_By_Key(ctx, Kv_Key("key"))
	erre(err)
	eq(row.Val, "val0")

	err = db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val1"))
	erre(err)
	row, err = db.Get_Kv_By_Key(ctx, Kv_Key("key"))
	erre(err)
	eq(row.Val, "val1")

	rows, err := db.All_Kv(ctx)
	erre(err)
	eq(len(rows), 1)
}
