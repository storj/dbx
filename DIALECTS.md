# Dialects

## Spanner

Spanner dialect uses the Spanner driver at https://github.com/googleapis/go-sql-spanner.
There are a few considerations when developing features for Spanner.

Spanner does not support nested transactions. Existing transaction should be used instead
of creating a new one.

go-sql-spanner only allows using spanner.NullJSON for json arguments. To make it
compatible with our `[]byte` queries, we need to wrap `[]byte` with `spanner.NullJSON`.

Spanner does not support row and tuple comparisons, which means it needs to unfold the
comparison.

For constraint errors Spanner returns the following codes in these scenarios:

* The unique key already exists: `codes.AlreadyExists`
* The change fails a check constraint: `codes.OutOfRange`
* A foreign key does not exist: `codes.FailedPrecondition`