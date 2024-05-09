// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package paged_composite

//go:generate dbx golang --package paged_composite -d sqlite3 -d pgx -d postgres -d cockroach -d pgxcockroach -d spanner paged_composite.dbx .
