// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package default_fields

import (
	"context"
	"storj.io/dbx/testrun"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultFields(t *testing.T) {
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		ctx := context.Background()

		_, err := db.Exec(strings.Join(db.DropSchema(), "\n"))
		require.NoError(t, err)

		_, err = db.Exec(strings.Join(db.Schema(), "\n"))
		require.NoError(t, err)

		{
			err := db.CreateNoReturn_Foo(ctx, Foo_Create_Fields{})
			require.NoError(t, err)
			row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(1))
			require.NoError(t, err)
			require.Equal(t, 10, int(row.A))
			require.Equal(t, 0, int(row.B))
			require.Equal(t, 20, int(row.C))
		}

		{
			err := db.CreateNoReturn_Foo(ctx, Foo_Create_Fields{
				C: Foo_C(25),
			})
			require.NoError(t, err)
			row, err := db.Get_Foo_By_Pk(ctx, Foo_Pk(2))
			require.NoError(t, err)
			require.Equal(t, 10, int(row.A))
			require.Equal(t, 0, int(row.B))
			require.Equal(t, 25, int(row.C))
		}

		{
			err := db.CreateNoReturn_Bar(ctx, Bar_A(200), Bar_B(100), Bar_Create_Fields{})
			require.NoError(t, err)
			row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(1))
			require.NoError(t, err)
			require.Equal(t, 200, int(row.A))
			require.Equal(t, 100, int(row.B))
			require.Equal(t, 40, int(row.C))
		}

		{
			err := db.CreateNoReturn_Bar(ctx, Bar_A(250), Bar_B(150), Bar_Create_Fields{
				C: Bar_C(45),
			})
			require.NoError(t, err)
			row, err := db.Get_Bar_By_Pk(ctx, Bar_Pk(2))
			require.NoError(t, err)
			require.Equal(t, 250, int(row.A))
			require.Equal(t, 150, int(row.B))
			require.Equal(t, 45, int(row.C))
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

			err := db.CreateNoReturn_Baz(ctx, optional)
			require.NoError(t, err)

			row, err := db.Get_Baz_By_Pk(ctx, Baz_Pk(int64(i+1)))
			require.NoError(t, err)

			require.Equal(t, expA, int(row.A))
			require.Equal(t, expB, int(row.B))
			require.Equal(t, expC, int(row.C))
		}
	})
}
