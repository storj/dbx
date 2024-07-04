# Dialects

## Spanner

Spanner dialect uses the Spanner driver at https://github.com/googleapis/go-sql-spanner.
There are a few considerations when developing features for Spanner.

Spanner does not support nested transactions. Existing transaction should be used instead
of creating a new one.

go-sql-spanner driver has an issue where it starts any `QueryContext` call with a
read-only transaction. This should not be the case in situations like "UPDATE ... THEN
RETURN ...". See https://github.com/googleapis/go-sql-spanner/issues/235 for more
details.

go-sql-spanner does not automatically convert uint64 to int64. Spanner in general pretty
much only supports int64 from the integer types, so any other datatype needs to be
converted. See https://github.com/googleapis/go-sql-spanner/pull/272 for the fix. Once it
is merged the workarounds for uint64 can be removed.

go-sql-spanner only allows using spanner.NullJSON for json arguments. To make it
compatible with our `[]byte` queries, we need to wrap `[]byte` with `spanner.NullJSON`.

Spanner does not support row and tuple comparisons, which means it needs to unfold the
comparison.