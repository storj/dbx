// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package testrun

import (
	"io"
	"os"
	"testing"

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
		t.Log("Skipping pq and pgx tests because environment variable STORJ_TEST_POSTGRES is not set")
	} else {

		t.Run("postgres", func(t *testing.T) {
			pqDb, err := open("postgres", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})

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
		t.Log("Skipping cockroach and pgxcockroach tests because environment variable STORJ_TEST_COCKROACH is not set")
	} else {

		t.Run("cockroach", func(t *testing.T) {
			pqDb, err := open("cockroach", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})

		t.Run("pgxcockroach", func(t *testing.T) {
			pqDb, err := open("pgxcockroach", dsn)
			require.NoError(t, err)
			defer func() {
				err := pqDb.Close()
				require.NoError(t, err)
			}()

			callback(t, pqDb)
		})
	}
}
