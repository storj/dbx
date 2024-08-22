// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/default_fields"
	"storj.io/dbx/testutil"
)

func TestDefaultFields(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		ctx := context.Background()

		testutil.RecreateSchema(t, db)

		{
			foo, err := db.Create_Foo(ctx, Foo_Create_Fields{})

			require.NoError(t, err)
			row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(foo.Pk))
			require.NoError(t, err)
			require.Equal(t, 10, row.A)
			require.Equal(t, 0, row.B)
			require.Equal(t, 20, row.C)
		}

		{
			foo, err := db.Create_Foo(ctx, Foo_Create_Fields{
				C: Foo_C(25),
			})
			require.NoError(t, err)
			row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(foo.Pk))
			require.NoError(t, err)
			require.Equal(t, 10, row.A)
			require.Equal(t, 0, row.B)
			require.Equal(t, 25, row.C)
		}

		{
			bar, err := db.Create_Bar(ctx, Bar_A(200), Bar_B(100), Bar_Create_Fields{})
			require.NoError(t, err)
			row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(bar.Pk))
			require.NoError(t, err)
			require.Equal(t, 200, row.A)
			require.Equal(t, 100, row.B)
			require.Equal(t, 40, row.C)
		}

		{
			bar, err := db.Create_Bar(ctx, Bar_A(250), Bar_B(150), Bar_Create_Fields{
				C: Bar_C(45),
			})
			require.NoError(t, err)
			row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(bar.Pk))
			require.NoError(t, err)
			require.Equal(t, 250, row.A)
			require.Equal(t, 150, row.B)
			require.Equal(t, 45, row.C)
		}

		{
			_, err := db.Create_Minimal(ctx)
			require.NoError(t, err)
		}

		for i := 0; i < 8; i++ {
			expA := 50
			expB := 60
			expC := 70

			var optional Baz_Create_Fields
			if i%2 == 0 {
				optional.A = Baz_A(i)
				expA = i
			}
			if (i/2)%2 == 0 {
				optional.B = Baz_B(i)
				expB = i
			}
			if (i/4)%2 == 0 {
				optional.C = Baz_C(i)
				expC = i
			}

			baz, err := db.Create_Baz(ctx, optional)
			require.NoError(t, err)

			row, err := db.Get_Baz_By_Pk(ctx, Baz_Pk(baz.Pk))
			require.NoError(t, err)

			require.Equal(t, expA, row.A)
			require.Equal(t, expB, row.B)
			require.Equal(t, expC, row.C)
		}
	})
}
