// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

//go:build dialects
// +build dialects

package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)
