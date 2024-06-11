// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package replace

//go:generate dbx golang --package replace -d sqlite3 -d pgx -d pgxcockroach -d spanner replace.dbx .
