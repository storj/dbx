// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package where_or

//go:generate dbx golang --package where_or -d sqlite3 -d pgx -d pgxcockroach -d spanner where_or.dbx .
//go:generate dbx schema -d sqlite3 -d pgx -d pgxcockroach -d spanner where_or.dbx .
