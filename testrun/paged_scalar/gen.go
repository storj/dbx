// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package paged_scalar

//go:generate dbx golang --package paged_scalar -d sqlite3 -d pgx -d postgres -d cockroach -d pgxcockroach paged_scalar.dbx .
