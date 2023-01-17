package main

import (
	"context"
	"fmt"
	"os"
)

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

func eq(a, b interface{}) {
	if a != b {
		panic(fmt.Sprintf("invalid: %#v != %#v", a, b))
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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS foos, bars, bazs")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS foos, bars, bazs")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	{
		err := db.CreateNoReturn_Foo(ctx, Foo_Create_Fields{})
		erre(err)
		row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(1))
		erre(err)
		eq(int(row.A), 10)
		eq(int(row.B), 0)
		eq(int(row.C), 20)
	}

	{
		err := db.CreateNoReturn_Foo(ctx, Foo_Create_Fields{
			C: Foo_C(25),
		})
		erre(err)
		row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(2))
		erre(err)
		eq(int(row.A), 10)
		eq(int(row.B), 0)
		eq(int(row.C), 25)
	}

	{
		err := db.CreateNoReturn_Bar(ctx, Bar_A(200), Bar_B(100), Bar_Create_Fields{})
		erre(err)
		row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(1))
		erre(err)
		eq(int(row.A), 200)
		eq(int(row.B), 100)
		eq(int(row.C), 40)
	}

	{
		err := db.CreateNoReturn_Bar(ctx, Bar_A(250), Bar_B(150), Bar_Create_Fields{
			C: Bar_C(45),
		})
		erre(err)
		row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(2))
		erre(err)
		eq(int(row.A), 250)
		eq(int(row.B), 150)
		eq(int(row.C), 45)
	}

	for i := 0; i < 8; i++ {
		expA := 50
		expB := 60
		expC := 70

		var optional Baz_Create_Fields
		if i%2 == 0 {
			optional.A = Baz_A(i)
			expA = i
		}
		if (i/2)%2 == 0 {
			optional.B = Baz_B(i)
			expB = i
		}
		if (i/4)%2 == 0 {
			optional.C = Baz_C(i)
			expC = i
		}

		err := db.CreateNoReturn_Baz(ctx, optional)
		erre(err)

		row, err := db.Get_Baz_By_Pk(ctx, Baz_Pk(int64(i+1)))
		erre(err)

		eq(int(row.A), expA)
		eq(int(row.B), expB)
		eq(int(row.C), expC)
	}
}
