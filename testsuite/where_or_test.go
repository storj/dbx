// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/where_or"
	"storj.io/dbx/testutil"
)

func TestWhereOr(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(ctx context.Context, t *testing.T, db *DB) {
		testutil.RecreateSchema(ctx, t, db)

		foo, err := db.Create_Foo(ctx, Foo_A(1), Foo_B("b"), Foo_C("c"))
		require.NoError(t, err)

		rows, err := db.All_Foo_By__A_Or__B_Or_C(ctx, Foo_A(1), Foo_B("x"), Foo_C("c"))
		require.NoError(t, err)
		require.Equal(t, 1, len(rows))
		require.Equal(t, Foo{Pk: foo.Pk, A: 1, B: "b", C: "c"}, *rows[0])
	})
}
