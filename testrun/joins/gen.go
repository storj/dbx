// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package joins

//go:generate dbx golang --package joins -d sqlite3 -d pgx -d postgres -d cockroach -d pgxcockroach -i session.dbx user.dbx .
