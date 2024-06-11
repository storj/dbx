// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package joins

//go:generate dbx golang --package joins -d sqlite3 -d pgx -d pgxcockroach -d spanner -i session.dbx user.dbx .
//go:generate dbx schema -d sqlite3 -d pgx -d pgxcockroach -d spanner -i session.dbx user.dbx .
