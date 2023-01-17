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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS foos")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS foos")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	err = db.CreateNoReturn_Foo(ctx, Foo_A("a"), Foo_B("b"), Foo_C("c"))
	erre(err)

	rows, err := db.All_Foo_By__A_Or__B_Or_C(ctx, Foo_A("x"), Foo_B("x"), Foo_C("c"))
	erre(err)
	eq(len(rows), 1)
	eq(*rows[0], Foo{Pk: 1, A: "a", B: "b", C: "c"})
}
