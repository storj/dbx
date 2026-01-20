// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:generate go run generate.go

func main() {
	dbxs, err := filepath.Glob("good/*.dbx")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	exitcode := 0
	for _, dbx := range dbxs {
		fullname := dbx[:len(dbx)-len(filepath.Ext(dbx))]
		pkgname := filepath.Base(fullname)
		outdir := filepath.Join("generated", pkgname)

		_ = os.MkdirAll(outdir, 0755)

		ctx := context.Background()
		out, err := exec.CommandContext(ctx, "dbx", "golang",
			"--package", pkgname,
			"-d", "sqlite3",
			"-d", "pgx",
			"-d", "pgxcockroach",
			"-d", "spanner",
			dbx,
			outdir,
		).CombinedOutput()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, string(out))
			_, _ = fmt.Fprintln(os.Stderr, err)
			exitcode = 1
		}

		out, err = exec.CommandContext(ctx, "dbx", "schema",
			"-d", "sqlite3",
			"-d", "pgx",
			"-d", "pgxcockroach",
			"-d", "spanner",
			dbx,
			outdir,
		).CombinedOutput()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, string(out))
			_, _ = fmt.Fprintln(os.Stderr, err)
			exitcode = 1
		}
	}

	os.Exit(exitcode)
}
