// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package wrap_type

//go:generate dbx golang --package wrap_type -d sqlite3 -d pgx -d pgxcockroach -d spanner wrap_type.dbx .
//go:generate dbx schema -d sqlite3 -d pgx -d pgxcockroach -d spanner wrap_type.dbx .
