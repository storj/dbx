// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package where_or

import (
	"context"
	"storj.io/dbx/testrun"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWhereOr(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		_, err := db.Exec(strings.Join(db.DropSchema(), "\n"))
		require.NoError(t, err)

		_, err = db.Exec(strings.Join(db.Schema(), "\n"))
		require.NoError(t, err)

		err = db.CreateNoReturn_Foo(ctx, Foo_A("a"), Foo_B("b"), Foo_C("c"))
		require.NoError(t, err)

		rows, err := db.All_Foo_By__A_Or__B_Or_C(ctx, Foo_A("x"), Foo_B("x"), Foo_C("c"))
		require.NoError(t, err)
		require.Equal(t, 1, len(rows))
		require.Equal(t, Foo{Pk: 1, A: "a", B: "b", C: "c"}, *rows[0])
	})
}
