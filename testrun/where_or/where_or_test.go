// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package where_or

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/testrun"
)

func TestWhereOr(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		_, err := db.Exec(strings.Join(db.DropSchema(), "\n"))
		require.NoError(t, err)

		_, err = db.Exec(strings.Join(db.Schema(), "\n"))
		require.NoError(t, err)

		foo, err := db.Create_Foo(ctx, Foo_A("a"), Foo_B("b"), Foo_C("c"))
		require.NoError(t, err)

		rows, err := db.All_Foo_By__A_Or__B_Or_C(ctx, Foo_A("x"), Foo_B("x"), Foo_C("c"))
		require.NoError(t, err)
		require.Equal(t, 1, len(rows))
		require.Equal(t, Foo{Pk: foo.Pk, A: "a", B: "b", C: "c"}, *rows[0])
	})
}
