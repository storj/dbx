// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package default_fields

//go:generate dbx golang --package default_fields -d sqlite3 -d pgx -d pgxcockroach -d spanner default_fields.dbx .
