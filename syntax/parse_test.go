// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import (
	"testing"

	"storj.io/dbx/testutil"
)

func TestEmptyParse(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	_, err := Parse("", nil)
	tw.AssertNoError(err)
}
