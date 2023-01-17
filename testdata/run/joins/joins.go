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
	_, err = pqDb.Exec("DROP TABLE IF EXISTS sessions, associated_accounts, users")
	erre(err)
	runDb(pqDb)
	err = pqDb.Close()
	erre(err)

	pgxDb, err := Open("pgx", dsn)
	erre(err)
	_, err = pgxDb.Exec("DROP TABLE IF EXISTS sessions, associated_accounts, users")
	erre(err)
	runDb(pgxDb)
	err = pgxDb.Close()
	erre(err)
}

func runDb(db *DB) {
	_, err := db.Exec(db.Schema())
	erre(err)

	user, err := db.Create_User(ctx)
	erre(err)

	aa, err := db.Create_AssociatedAccount(ctx,
		AssociatedAccount_UserPk(user.Pk))
	erre(err)

	sess, err := db.Create_Session(ctx,
		Session_UserPk(user.Pk))
	erre(err)

	rows, err := db.All_Session_Id_By_AssociatedAccount_Pk(ctx,
		AssociatedAccount_Pk(aa.Pk))
	erre(err)

	if len(rows) != 1 || rows[0].Id != sess.Id {
		panic("invalid")
	}
}
