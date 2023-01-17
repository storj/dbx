package main

import (
	"context"
	"fmt"
	"os"
	"time"
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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS consumed_serials")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS consumed_serials")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	for i := 0; i < 1000; i++ {
		err = db.CreateNoReturn_ConsumedSerial(ctx,
			ConsumedSerial_ExpiresAt(time.Now()),
			ConsumedSerial_StorageNodeId([]byte(fmt.Sprintf("node%d", i))),
			ConsumedSerial_ProjectId([]byte(fmt.Sprintf("proj%d", i))),
			ConsumedSerial_BucketName([]byte(fmt.Sprintf("bucket%d", i))),
			ConsumedSerial_Action(1),
			ConsumedSerial_SerialNumber([]byte(fmt.Sprintf("serial%d", i))),
			ConsumedSerial_Settled(100))
		erre(err)
	}

	var total int
	var rows []*ConsumedSerial
	var next *Paged_ConsumedSerial_By_ExpiresAt_Greater_Continuation
again:
	rows, next, err = db.Paged_ConsumedSerial_By_ExpiresAt_Greater(ctx,
		ConsumedSerial_ExpiresAt(time.Now().Add(-time.Minute)),
		10, next)
	erre(err)
	total += len(rows)

	if next != nil {
		goto again
	}
	eq(total, 1000)
}
