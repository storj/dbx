// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"golang.org/x/tools/go/packages"

	"storj.io/dbx/testutil"
)

var buildConfig = &packages.Config{
	Mode: 0 |
		packages.NeedTypes |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedCompiledGoFiles,
}

func TestBuild(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	data_dir := filepath.Join("testdata", "build")

	names, err := filepath.Glob(filepath.Join(data_dir, "*.dbx"))
	tw.AssertNoError(err)

	for _, name := range names {
		name := name
		tw.Runp(filepath.Base(name), func(tw *testutil.T) {
			testBuildFile(tw, name)
		})
	}
}

func testBuildFile(t *testutil.T, file string) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("%s\n%s", val, string(debug.Stack()))
		}
	}()

	dir, err := os.MkdirTemp("", "dbx")
	t.AssertNoError(err)
	defer os.RemoveAll(dir)

	dbx_source, err := os.ReadFile(file)
	t.AssertNoError(err)
	t.Context("dbx", linedSource(dbx_source))
	d := loadDirectives(t, dbx_source)

	dialects := []string{"postgres", "sqlite3", "cockroach", "pgx", "pgxcockroach"}
	if other := d.lookup("dialects"); other != nil {
		dialects = other
		t.Logf("using dialects: %q", dialects)
	}

	type options struct {
		rx       bool
		userdata bool
	}

	runBuild := func(opts options) {
		t.Logf("[%s] generating... %+v", file, opts)
		err = newGlobal().golangCmd("", dialects, "", opts.rx, opts.userdata, []string{file}, dir)
		if d.has("fail_gen") {
			t.AssertError(err, d.get("fail_gen"))
			return
		} else {
			t.AssertNoError(err)
		}

		t.Logf("[%s] loading...", file)
		go_file := filepath.Join(dir, filepath.Base(file)+".go")
		go_source, err := os.ReadFile(go_file)
		t.AssertNoError(err)
		t.Context("go", linedSource(go_source))

		t.Logf("[%s] parsing...", file)
		fset := token.NewFileSet()
		_, err = parser.ParseFile(fset, go_file, go_source, parser.AllErrors)
		t.AssertNoError(err)

		t.Logf("[%s] compiling...", file)
		pkg, err := packages.Load(buildConfig, go_file)

		if d.has("fail") {
			t.AssertError(err, d.get("fail"))
		} else {
			t.AssertNoError(err)
			if len(pkg[0].Errors) > 0 {
				errMsg := ""
				for _, err := range pkg[0].Errors {
					errMsg += "    " + err.Error()
				}
				t.Logf("[%s] errors:\n%s", file, errMsg)
			}
			t.Assert(!pkg[0].IllTyped)
		}
	}

	runBuild(options{rx: false, userdata: false})
	runBuild(options{rx: false, userdata: true})
	runBuild(options{rx: true, userdata: false})
	runBuild(options{rx: true, userdata: true})
}
