// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const prefix = "__sqlbundle_"

const template = `%s
//
// DO NOT EDIT: automatically generated code.
//

//go:generate go run gen_bundle.go

package sqlbundle

const (
	Source = %q
	Prefix = %q
)
`

var afterImportRe = regexp.MustCompile(`(?m)^\)$`)

func main() {
	copyright, bundle := loadCopyright(), loadBundle()
	output := []byte(fmt.Sprintf(template, copyright, bundle, prefix))

	err := os.WriteFile("bundle.go", output, 0644)
	if err != nil {
		panic(err)
	}
}

func loadCopyright() string {
	fh, err := os.Open("gen_bundle.go")
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	var buf bytes.Buffer
	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "//") {
			return buf.String()
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	panic("unreachable")
}

func loadBundle() string {
	source, err := exec.Command("bundle",
		"-dst", "storj.io/dbx/sqlgen/sqlbundle",
		"-prefix", prefix,
		"storj.io/dbx/sqlgen").Output()
	if err != nil {
		fmt.Fprintln(os.Stdout, `ensure "golang.org/x/tools/cmd/bundle" is installed`)
		panic(err)
	}

	index := afterImportRe.FindIndex(source)
	if index == nil {
		panic("unable to find package clause")
	}

	return string(bytes.TrimSpace(source[index[1]:]))
}
