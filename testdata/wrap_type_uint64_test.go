// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testdata/generated/wrap_type_uint64"
	"storj.io/dbx/testutil"
)

func TestWrapTypeUint64(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		ctx := context.Background()

		testutil.RecreateSchema(t, db)

		// Create Person
		person, err := db.Create_Person(ctx,
			Person_Name("P1"),
			Person_Value(100),
			Person_ValueUp(101),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull(102),
				ValueNullUp: Person_ValueNullUp(103),
			})
		require.NoError(t, err)

		// Read Person
		row, err := db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Read Person
		row, err = db.Get_Person_By_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Value(100),
			Person_ValueUp(101),
			Person_ValueNull(102),
			Person_ValueNullUp(103),
		)
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		row, err = db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person
		row, err = db.Update_Person_By_Pk_And_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Pk(person.Pk),
			Person_Value(100),
			Person_ValueUp(101),
			Person_ValueNull(102),
			Person_ValueNullUp(103),
			Person_Update_Fields{
				ValueUp:     Person_ValueUp(111),
				ValueNullUp: Person_ValueNullUp(113),
			},
		)
		person.ValueUp = 111
		tmp := uint64(113)
		person.ValueNullUp = &tmp
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person with nil
		row, err = db.Update_Person_By_Pk_And_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Pk(person.Pk),
			Person_Value(100),
			Person_ValueUp(111),
			Person_ValueNull(102),
			Person_ValueNullUp(113),
			Person_Update_Fields{
				ValueNullUp: Person_ValueNullUp_Null(),
			},
		)
		person.ValueNullUp = nil
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Delete Person
		deleted, err := db.Delete_Person_By_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Value(100),
			Person_ValueUp(111),
			Person_ValueNull(102),
			Person_ValueNullUp_Null(),
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), deleted)

		// Bounds checks
		maxint64Bound := uint64(math.MaxInt64) + 1

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_Value(maxint64Bound),
			Person_ValueUp(101),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull(102),
				ValueNullUp: Person_ValueNullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_Value(100),
			Person_ValueUp(maxint64Bound),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull(102),
				ValueNullUp: Person_ValueNullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_Value(100),
			Person_ValueUp(101),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull(maxint64Bound),
				ValueNullUp: Person_ValueNullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_Value(100),
			Person_ValueUp(101),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull(102),
				ValueNullUp: Person_ValueNullUp(maxint64Bound),
			})
		require.Error(t, err)
	})
}
