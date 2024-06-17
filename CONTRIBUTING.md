# Contributing to dbx

## Running tests via docker

To run all the linters and tests quickly locally you can use:

```
docker buildx bake lint

docker buildx bake integration-test
```

## Running tests without docker

### Pointing tests at databases

Using docker might be cumbersome when trying to run specific tests. If you
are doing more finegrained development then usual `go test` will work as
well, as long as you setup your `STORJ_TEST_*` environment variables.

You can take a look at [test-environment.sh](./scripts/test-environment.sh)
for the full setup and starting the necessary endpoints.

The short version is, point these variables to working databases:

```
# for Postgres
export STORJ_TEST_POSTGRES="postgres://postgres@localhost/testdb?sslmode=disable"

# for Spanner Production
export STORJ_TEST_SPANNER=spanner://projects/PROJECTID/instances/INSTANCEID/databases/metainfo

# for Spanner Emulator
export STORJ_TEST_SPANNER=spanner://127.0.0.1:9010?emulator
```

If you wish to completely ignore tests for a specific database, then you can
use:

```
export STORJ_TEST_POSTGRES=omit

export STORJ_TEST_SPANNER=omit
```

### Running tests

Once you have setup your database configuration. Reinstall the `dbx` command with:

```
go install .
```

Regenerate any necessary parts:

```
go generate ./...
```

Finally run tests:

```
go test ./...
```