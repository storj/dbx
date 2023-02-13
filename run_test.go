// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"testing"

	"storj.io/dbx/testutil"
)

func TestRun(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	testdir := filepath.Join("testdata", "run")
	entries, err := os.ReadDir(testdir)
	tw.AssertNoError(err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(testdir, entry.Name())
		tw.Runp(entry.Name(), func(tw *testutil.T) {
			testRunFile(tw, path)
		})
	}
}

func testRunFile(t *testutil.T, dbxdir string) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("%s\n%s", val, string(debug.Stack()))
		}
	}()

	dbxfiles, err := filepath.Glob(filepath.Join(dbxdir, "*.dbx"))
	t.AssertNoError(err)

	var d directives
	for _, dbxfile := range dbxfiles {
		source, err := os.ReadFile(dbxfile)
		t.AssertNoError(err)
		t.Context("dbx:"+filepath.Base(dbxfile), linedSource(source))

		sd := loadDirectives(t, source)
		d.join(sd)
	}

	tempDir := t.TempDir()

	t.Logf("[%s] generating... {rx:%t, userdata:%t}", dbxfiles,
		d.has("rx"), d.has("userdata"))
	err = newGlobal().golangCmd("main", []string{"sqlite3", "postgres", "pgx"}, "",
		d.has("rx"), d.has("userdata"), dbxfiles, tempDir)
	if d.has("fail_gen") {
		t.AssertError(err, d.get("fail_gen"))
		return
	} else {
		t.AssertNoError(err)
	}

	gofiles, err := filepath.Glob(filepath.Join(dbxdir, "*.go"))
	t.AssertNoError(err)

	for _, gofile := range gofiles {
		go_source, err := os.ReadFile(gofile)
		t.AssertNoError(err)
		t.Context("go", linedSource(go_source))

		t.Logf("[%s] copying go source...", dbxdir)
		t.AssertNoError(os.WriteFile(
			filepath.Join(tempDir, filepath.Base(gofile)), go_source, 0644))
	}

	t.Logf("[%s] running output...", dbxdir)
	files, err := filepath.Glob(filepath.Join(tempDir, "*.go"))
	t.AssertNoError(err)

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("go", append([]string{"run"}, files...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	t.Context("stdout", stdout.String())
	t.Context("stderr", stderr.String())

	if d.has("fail") {
		t.AssertError(err, "")
		t.AssertContains(stderr.String(), d.get("fail"))
	} else {
		t.AssertNoError(err)
	}
}
