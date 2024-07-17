// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package testutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	spannerdriver "github.com/googleapis/go-sql-spanner"
	"github.com/jackc/pgx/v5"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func randomText(n int) string {
	letters := "abcdefghijiklmnopqrstuvwxyz"
	data := make([]byte, n)
	for i := range data {
		data[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(data)
}

type TestDB interface {
	io.Closer

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func RunDBTest[T TestDB](t *testing.T, open func(driver, source string) (db T, err error), callback func(t *testing.T, db T)) {
	t.Run("sqlite3", func(t *testing.T) {
		db, err := open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer func() {
			err := db.Close()
			require.NoError(t, err)
		}()
		callback(t, db)
	})

	dsn := os.Getenv("STORJ_TEST_POSTGRES")
	if dsn != "omit" {
		t.Run("pgx", func(t *testing.T) {
			if dsn == "" {
				t.Skip("Skipping pgx tests because environment variable STORJ_TEST_POSTGRES is not set")
			}

			schemaName := pgxSchemaName(t)
			dsn = pgxConnStrWithSchema(dsn, schemaName)

			db, err := open("pgx", dsn)
			require.NoError(t, err)
			defer func() { require.NoError(t, db.Close()) }()

			require.NoError(t, pgxCreateSchema(context.Background(), db, schemaName))
			defer func() { require.NoError(t, pgxDropSchema(context.Background(), db, schemaName)) }()

			callback(t, db)
		})
	}

	dsn = os.Getenv("STORJ_TEST_COCKROACH")
	if dsn != "omit" {
		t.Run("pgxcockroach", func(t *testing.T) {
			if dsn == "" {
				t.Skip("Skipping pgxcockroach tests because environment variable STORJ_TEST_COCKROACH is not set")
			}

			if strings.HasPrefix(dsn, "cockroach://") {
				dsn = "postgres://" + strings.TrimPrefix(dsn, "cockroach://")
			}

			schemaName := pgxSchemaName(t)
			dsn = pgxConnStrWithSchema(dsn, schemaName)

			db, err := open("pgxcockroach", dsn)
			require.NoError(t, err)
			defer func() { require.NoError(t, db.Close()) }()

			require.NoError(t, pgxCreateSchema(context.Background(), db, schemaName))
			defer func() { require.NoError(t, pgxDropSchema(context.Background(), db, schemaName)) }()

			callback(t, db)
		})
	}

	dsn = os.Getenv("STORJ_TEST_SPANNER")
	if dsn != "omit" {
		t.Run("spanner", func(t *testing.T) {
			ctx := context.Background()

			if dsn == "" {
				t.Skip("Skipping spanner tests because environment variable STORJ_TEST_SPANNER is not set")
			}

			spannerurl, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}

			isEmulator := spannerurl.Query().Has("emulator")

			if !isEmulator {
				db, err := open("spanner", dsn)
				require.NoError(t, err)
				defer func() { require.NoError(t, db.Close()) }()
				callback(t, db)
				return
			}

			// add the emulator information to the dsn
			query := ""
			for k, v := range spannerurl.Query() {
				if k == "emulator" {
					continue
				}
				query += ";" + k + "=" + v[0]
			}
			query += ";useplaintext=true"
			query += ";disableroutetoleader=true"

			// parse the parts that we have
			elements := strings.Split(strings.Trim(spannerurl.Path, "/ "), "/")
			if len(elements) == 1 && elements[0] == "" {
				elements = nil
			}

			readElement := func(at int, name string) string {
				if at >= len(elements) {
					return ""
				}
				if elements[at] != name {
					t.Fatalf("expected %q at %v, but found %q in %q", name, at, elements[at], spannerurl.Path)
				}
				if at+1 >= len(elements) {
					t.Fatalf("%v missing in %q", name, spannerurl.Path)
				}
				return elements[at+1]
			}

			projectid := readElement(0, "projects")
			instanceid := readElement(2, "instances")
			databaseid := readElement(4, "databases")

			// create a project id, if necessary
			if projectid == "" {
				projectid = fmt.Sprintf("pid-%v", mathrand.Int63())
			}
			// create a instance id, if necessary
			if instanceid == "" {
				instanceid = fmt.Sprintf("iid-%v", mathrand.Int63())
				spannerEmulatorCreateInstance(ctx, t, spannerurl.Host, projectid, instanceid)
			}
			// create a database id, if necessary
			if databaseid == "" {
				databaseid = fmt.Sprintf("did-%v", mathrand.Int63())
				spannerEmulatoreCreateDatabase(ctx, t, spannerurl.Host, projectid, instanceid, databaseid)
			}

			newdsn := "spanner://" + spannerurl.Host
			newdsn += "/projects/" + projectid + "/instances/" + instanceid + "/databases/" + databaseid
			newdsn += "?" + query

			t.Logf("dsn=%v", newdsn)

			db, err := open("spanner", newdsn)
			require.NoError(t, err)
			defer func() { require.NoError(t, db.Close()) }()

			callback(t, db)
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
	p := time.Millisecond * 250
	for i := 0; i < 10; i++ {
		err = RecreateSchemaOnce(db)
		if err == nil {
			return
		}
		time.Sleep(p)

		p *= 2
		if p > 2*time.Second {
			p = 2 * time.Second
		}
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

// IsSqlite returns true of the db is a sqlite db.
func IsSqlite[DB WithDriver](db *sql.DB) bool {
	_, ok := db.Driver().(*sqlite3.SQLiteDriver)
	return ok
}

func spannerEmulatorCreateInstance(ctx context.Context, t *testing.T, hostport, projectID, instanceID string) {
	admin, err := instance.NewInstanceAdminClient(ctx, spannerEmulatorOptions(hostport)...)
	if err != nil {
		t.Fatalf("failed to create instance admin client: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Close(); err != nil {
			t.Fatal(err)
		}
	})

	op, err := admin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + projectID,
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			Config:      "projects/" + projectID + "/instanceConfigs/emulator-config",
			DisplayName: instanceID,
			NodeCount:   1,
		},
	})
	if err != nil {
		t.Fatalf("could not create instance: %v", err)
	}

	// Wait for the instance creation to finish.
	if _, err := op.Wait(ctx); err != nil {
		t.Fatalf("failed to wait instance creation: %v", err)
	}

	t.Cleanup(func() {
		err := admin.DeleteInstance(ctx, &instancepb.DeleteInstanceRequest{
			Name: "projects/" + projectID + "/instances/" + instanceID,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func spannerEmulatoreCreateDatabase(ctx context.Context, t *testing.T, hostport, projectID, instanceID, databaseID string) {
	admin, err := database.NewDatabaseAdminClient(ctx, spannerEmulatorOptions(hostport)...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Close(); err != nil {
			t.Fatal(err)
		}
	})

	op, err := admin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          "projects/" + projectID + "/instances/" + instanceID,
		CreateStatement: "CREATE DATABASE `" + databaseID + "`",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Wait for the database creation to finish.
	if _, err := op.Wait(ctx); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := admin.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
			Database: "projects/" + projectID + "/instances/" + instanceID + "/databases/" + databaseID,
		}); err != nil {
			t.Fatal(err)
		}
	})
}

func spannerEmulatorOptions(hostport string) []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint(hostport),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
	}
}

func pgxSchemaName(t *testing.T) string {
	const randLen = 8
	const maxLen = 64 - randLen
	testname := strings.TrimPrefix(t.Name(), "Test")
	if len(testname) > maxLen {
		testname = testname[:maxLen]
	}

	nameCleaner := regexp.MustCompile(`[^\w]`)
	testname = nameCleaner.ReplaceAllString(testname, "_")

	return testname + "-" + randomText(randLen)
}

// pgxConnStrWithSchema adds schema to a connection string.
func pgxConnStrWithSchema(connstr, schema string) string {
	if strings.Contains(connstr, "?") {
		connstr += "&options="
	} else {
		connstr += "?options="
	}
	return connstr + url.QueryEscape("--search_path="+pgxQuoteIdentifier(schema))
}

// pgxQuoteIdentifier quotes an identifier for use in an interpolated SQL string.
func pgxQuoteIdentifier(ident string) string {
	return pgx.Identifier{ident}.Sanitize()
}

// pgxCreateSchema creates a schema if it doesn't exist.
func pgxCreateSchema(ctx context.Context, db TestDB, schema string) (err error) {
	for try := 0; try < 5; try++ {
		_, err = db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+pgxQuoteIdentifier(schema)+`;`)

		// Postgres `CREATE SCHEMA IF NOT EXISTS` may return "duplicate key value violates unique constraint".
		// In that case, we will retry rather than doing anything more complicated.
		//
		// See more in: https://stackoverflow.com/a/29908840/192220

		continue
	}

	return err
}

// pgxDropSchema drops the named schema.
func pgxDropSchema(ctx context.Context, db TestDB, schema string) error {
	_, err := db.ExecContext(ctx, `DROP SCHEMA `+pgxQuoteIdentifier(schema)+` CASCADE;`)
	return err
}
