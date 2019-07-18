// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"text/scanner"

	"storj.io/dbx/errutil"
)

func expectedKeyword(pos scanner.Position, actual string, expected ...string) (
	err error) {

	if len(expected) == 1 {
		return errutil.New(pos, "expected %q, got %q", expected[0], actual)
	} else {
		return errutil.New(pos, "expected one of %q, got %q", expected, actual)
	}
}

func expectedToken(pos scanner.Position, actual Token, expected ...Token) (
	err error) {

	if len(expected) == 1 {
		return errutil.New(pos, "expected %q; got %q", expected[0], actual)
	} else {
		return errutil.New(pos, "expected one of %v; got %q", expected, actual)
	}
}

func previouslyDefined(pos scanner.Position, kind, field string,
	where scanner.Position) error {

	return errutil.New(pos,
		"%s already defined on %s. previous definition at %s",
		field, kind, where)
}
