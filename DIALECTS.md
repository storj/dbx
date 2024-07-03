# Dialects

## Spanner

Spanner dialect uses the Spanner driver at https://github.com/googleapis/go-sql-spanner. There are a few considerations when developing features for Spanner.
- Spanner does not support nested transactions; an existing transaction should be used or an entirely separate transaction
- `go-sql-spanner` has an issue currently where, in the default AutoCommitDMLMode == Transactional mode, queries with a returning clause (using `THEN RETURN`) use a read-only transaction. All Spanner queries with a returning clause should therefore execute inside a transaction. See https://github.com/googleapis/go-sql-spanner/issues/235 for more info and to track the issue.