// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/in_filter"
	"storj.io/dbx/testutil"
)

func TestInFilter(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(ctx context.Context, t *testing.T, db *DB) {
		testutil.RecreateSchema(ctx, t, db)

		create := func(pk string, a int64, b string, tag *string) {
			require.NoError(t, db.CreateNoReturn_Foo(ctx,
				Foo_Pk([]byte(pk)), Foo_A(a), Foo_B(b), Foo_Val(0),
				Foo_Create_Fields{Tag: Foo_Tag_Raw(tag)}))
		}
		tag := "tag"
		create("p1", 1, "x", nil)
		create("p2", 2, "y", &tag)
		create("p3", 3, "z", &tag)

		as := func(rows []*Foo) (out []int64) {
			for _, row := range rows {
				out = append(out, row.A)
			}
			return out
		}

		rows, err := db.All_Foo_By_Pk_In_OrderBy_Asc_A(ctx, [][]byte{[]byte("p3"), []byte("p1"), []byte("missing")})
		require.NoError(t, err)
		require.Equal(t, []int64{1, 3}, as(rows))

		// empty list matches nothing
		rows, err = db.All_Foo_By_Pk_In_OrderBy_Asc_A(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, rows)

		// arguments before and after the list keep their positions, also
		// when a null condition drops its argument
		rows, err = db.All_Foo_By_Tag_And_B_In_And_A_Greater_OrderBy_Asc_A(ctx,
			Foo_Tag(tag), []string{"x", "y", "z"}, Foo_A(2))
		require.NoError(t, err)
		require.Equal(t, []int64{3}, as(rows))

		rows, err = db.All_Foo_By_Tag_And_B_In_And_A_Greater_OrderBy_Asc_A(ctx,
			Foo_Tag_Null(), []string{"x", "y"}, Foo_A(0))
		require.NoError(t, err)
		require.Equal(t, []int64{1}, as(rows))

		// "in" inside an or group
		rows, err = db.All_Foo_By__A_In_Or_B_OrderBy_Asc_A(ctx, []int64{1}, Foo_B("z"))
		require.NoError(t, err)
		require.Equal(t, []int64{1, 3}, as(rows))

		rows, err = db.All_Foo_By__A_In_Or_B_OrderBy_Asc_A(ctx, nil, Foo_B("y"))
		require.NoError(t, err)
		require.Equal(t, []int64{2}, as(rows))

		count, err := db.Count_Foo_By_A_In(ctx, []int64{1, 2, 5})
		require.NoError(t, err)
		require.EqualValues(t, 2, count)

		pks, err := db.Limited_Foo_Pk_By_A_In_OrderBy_Asc_A(ctx, []int64{1, 2, 3}, 2, 1)
		require.NoError(t, err)
		require.Len(t, pks, 2)
		require.Equal(t, []byte("p2"), pks[0].Pk)
		require.Equal(t, []byte("p3"), pks[1].Pk)

		// update arguments come before the where arguments
		err = db.UpdateNoReturn_Foo_By_Pk_And_A_In(ctx, Foo_Pk([]byte("p2")), []int64{2, 3},
			Foo_Update_Fields{Val: Foo_Val(7)})
		require.NoError(t, err)
		err = db.UpdateNoReturn_Foo_By_Pk_And_A_In(ctx, Foo_Pk([]byte("p3")), []int64{1},
			Foo_Update_Fields{Val: Foo_Val(9)})
		require.NoError(t, err)
		rows, err = db.All_Foo_By_Pk_In_OrderBy_Asc_A(ctx, [][]byte{[]byte("p1"), []byte("p2"), []byte("p3")})
		require.NoError(t, err)
		require.Equal(t, 0, rows[0].Val)
		require.Equal(t, 7, rows[1].Val)
		require.Equal(t, 0, rows[2].Val)

		deleted, err := db.Delete_Foo_By_A_In(ctx, []int64{1, 3, 9})
		require.NoError(t, err)
		require.EqualValues(t, 2, deleted)

		deleted, err = db.Delete_Foo_By_A_In(ctx, nil)
		require.NoError(t, err)
		require.EqualValues(t, 0, deleted)

		count, err = db.Count_Foo_By_A_In(ctx, []int64{1, 2, 3})
		require.NoError(t, err)
		require.EqualValues(t, 1, count)
	})
}

func TestInFilterTimes(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(ctx context.Context, t *testing.T, db *DB) {
		testutil.RecreateSchema(ctx, t, db)

		ts1 := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
		ts2 := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
		require.NoError(t, db.CreateNoReturn_Bar(ctx, Bar_Id(1), Bar_Ts(ts1)))
		require.NoError(t, db.CreateNoReturn_Bar(ctx, Bar_Id(2), Bar_Ts(ts2)))

		// the same instant in a non-UTC zone must match, like "=" does
		zone := time.FixedZone("test", 5*60*60)
		bars, err := db.All_Bar_By_Ts_In_OrderBy_Asc_Id(ctx, []time.Time{ts2.In(zone)})
		require.NoError(t, err)
		require.Len(t, bars, 1)
		require.Equal(t, 2, bars[0].Id)

		if testutil.IsSpanner[*DB](db.DB) {
			t.Skip("TODO(spanner): time.Time can't be inserted into a DATE column")
		}

		day1 := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
		day2 := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
		require.NoError(t, db.CreateNoReturn_Baz(ctx, Baz_Id(1), Baz_Day(day1)))
		require.NoError(t, db.CreateNoReturn_Baz(ctx, Baz_Id(2), Baz_Day(day2)))

		bazs, err := db.All_Baz_By_Day_In_OrderBy_Asc_Id(ctx, []time.Time{day1, day2})
		require.NoError(t, err)
		require.Len(t, bazs, 2)
		require.Equal(t, 1, bazs[0].Id)
		require.Equal(t, 2, bazs[1].Id)
	})
}
