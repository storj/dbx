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

func assert(x bool) {
	if !x {
		panic("assertion failed")
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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS a, b, c")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS a, b, c")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	a, err := db.Create_A(ctx)
	erre(err)

	b1, err := db.Create_B(ctx, B_AId(a.Id))
	erre(err)
	c1, err := db.Create_C(ctx, C_Lat(0.0), C_Lon(0.0), C_BId(b1.Id))
	erre(err)

	b2, err := db.Create_B(ctx, B_AId(a.Id))
	erre(err)
	c2, err := db.Create_C(ctx, C_Lat(1.0), C_Lon(1.0), C_BId(b2.Id))
	erre(err)

	rows, err := db.All_A_B_C_By_A_Id_And_C_Lat_Less_And_C_Lat_Greater_And_C_Lon_Less_And_C_Lon_Greater(ctx,
		A_Id(a.Id),
		C_Lat(10.0), C_Lat(-10.0),
		C_Lon(10.0), C_Lon(-10.0))
	erre(err)

	assert(len(rows) == 2)

	assert(rows[0].A.Id == a.Id)
	assert(rows[0].B.Id == b1.Id)
	assert(rows[0].C.Id == c1.Id)
	assert(rows[0].C.Lat == 0)
	assert(rows[0].C.Lon == 0)

	assert(rows[1].A.Id == a.Id)
	assert(rows[1].B.Id == b2.Id)
	assert(rows[1].C.Id == c2.Id)
	assert(rows[1].C.Lat == 1)
	assert(rows[1].C.Lon == 1)
}
