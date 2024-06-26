// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package unique_checking

//go:generate dbx golang --package unique_checking -d sqlite3 -d pgx -d pgxcockroach -d spanner unique_checking.dbx .
//go:generate dbx schema -d sqlite3 -d pgx -d pgxcockroach -d spanner unique_checking.dbx .
