// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package testrun

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"testing"
	"time"

	spannerdriver "github.com/googleapis/go-sql-spanner"

	"github.com/stretchr/testify/require"
)

func RunDBTest[T io.Closer](t *testing.T, open func(driver, source string) (db T, err error), callback func(t *testing.T, db T)) {
	t.Run("sqlite3", func(t *testing.T) {
		sqliteDb, err := open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer func() {
			err := sqliteDb.Close()
			require.NoError(t, err)
		}()
		callback(t, sqliteDb)
	})

	dsn := os.Getenv("STORJ_TEST_POSTGRES")
	if dsn == "" {
		t.Log("Skipping pgx tests because environment variable STORJ_TEST_POSTGRES is not set")
	} else {
		t.Run("pgx", func(t *testing.T) {
			pqDb, err := open("pgx", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})
	}

	dsn = os.Getenv("STORJ_TEST_COCKROACH")
	if dsn == "" {
		t.Log("Skipping pgxcockroach tests because environment variable STORJ_TEST_COCKROACH is not set")
	} else {
		t.Run("pgxcockroach", func(t *testing.T) {
			pqDb, err := open("pgx", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})
	}

	dsn = os.Getenv("STORJ_TEST_SPANNER")
	if dsn == "" {
		t.Log("Skipping spanner tests because environment variable STORJ_TEST_SPANNER is not set")
	} else {

		t.Run("spanner", func(t *testing.T) {
			pqDb, err := open("spanner", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})
	}
}

// SchemaHandler contains methods required for a schema recreation.
type SchemaHandler interface {
	Schema() []string
	DropSchema() []string
	Exec(query string, args ...any) (sql.Result, error)
}

// RecreateSchema will drop and recreate schema. Will try it with multiple times with increased sleep time.
// To void errors like "Schema change operation rejected because a concurrent schema change operation or read-write transaction is already in progress.".
func RecreateSchema(t *testing.T, db SchemaHandler) {
	var err error
	p := time.Millisecond * 500
	for i := 0; i < 10; i++ {
		err = RecreateSchemaOnce(db)
		if err == nil {
			return
		}
		time.Sleep(p)
		p *= 2
	}
	require.NoError(t, err)
}

// RecreateSchemaOnce will drop and recreate schema.
func RecreateSchemaOnce(db SchemaHandler) (err error) {
	// TODO(spanner): should use START BATCH DDL here, however there's no
	// easy way to get the conn via methods at the moment.

	for _, stmt := range db.DropSchema() {
		_, _ = db.Exec(stmt)
	}

	for _, stmt := range db.Schema() {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

// WithDriver represents a DB which has a driver (usually *sql.DB).
type WithDriver interface {
	Driver() driver.Driver
}

// IsSpanner returns true of the db is a spanner db.
func IsSpanner[DB WithDriver](db *sql.DB) bool {
	_, ok := db.Driver().(*spannerdriver.Driver)
	return ok
}
