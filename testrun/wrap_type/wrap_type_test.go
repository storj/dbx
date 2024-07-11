// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package wrap_type

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/testrun"
)

func TestWrapType(t *testing.T) {
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		Logger = func(format string, args ...any) { t.Logf(format, args...) }
		defer func() { Logger = nil }()

		ctx := context.Background()

		testrun.RecreateSchema(t, db)

		// Create Person
		person, err := db.Create_Person(ctx,
			Person_Name("P1"),
			Person_U64(100),
			Person_U64Up(101),
			Person_Create_Fields{
				U64Null:   Person_U64Null(102),
				U64NullUp: Person_U64NullUp(103),
			})
		require.NoError(t, err)

		// Read Person
		row, err := db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Read Person
		row, err = db.Get_Person_By_U64_And_U64Up_And_U64Null_And_U64NullUp(ctx,
			Person_U64(100),
			Person_U64Up(101),
			Person_U64Null(102),
			Person_U64NullUp(103),
		)
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		row, err = db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person
		row, err = db.Update_Person_By_Pk_And_U64_And_U64Up_And_U64Null_And_U64NullUp(ctx,
			Person_Pk(person.Pk),
			Person_U64(100),
			Person_U64Up(101),
			Person_U64Null(102),
			Person_U64NullUp(103),
			Person_Update_Fields{
				U64Up:     Person_U64Up(111),
				U64NullUp: Person_U64NullUp(113),
			},
		)
		person.U64Up = 111
		tmp := uint64(113)
		person.U64NullUp = &tmp
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person with nil
		row, err = db.Update_Person_By_Pk_And_U64_And_U64Up_And_U64Null_And_U64NullUp(ctx,
			Person_Pk(person.Pk),
			Person_U64(100),
			Person_U64Up(111),
			Person_U64Null(102),
			Person_U64NullUp(113),
			Person_Update_Fields{
				U64NullUp: Person_U64NullUp_Null(),
			},
		)
		person.U64NullUp = nil
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Delete Person
		deleted, err := db.Delete_Person_By_U64_And_U64Up_And_U64Null_And_U64NullUp(ctx,
			Person_U64(100),
			Person_U64Up(111),
			Person_U64Null(102),
			Person_U64NullUp_Null(),
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), deleted)

		// Bounds checks
		maxint64Bound := uint64(math.MaxInt64) + 1

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_U64(maxint64Bound),
			Person_U64Up(101),
			Person_Create_Fields{
				U64Null:   Person_U64Null(102),
				U64NullUp: Person_U64NullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_U64(100),
			Person_U64Up(maxint64Bound),
			Person_Create_Fields{
				U64Null:   Person_U64Null(102),
				U64NullUp: Person_U64NullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_U64(100),
			Person_U64Up(101),
			Person_Create_Fields{
				U64Null:   Person_U64Null(maxint64Bound),
				U64NullUp: Person_U64NullUp(103),
			})
		require.Error(t, err)

		_, err = db.Create_Person(ctx,
			Person_Name("P2"),
			Person_U64(100),
			Person_U64Up(101),
			Person_Create_Fields{
				U64Null:   Person_U64Null(102),
				U64NullUp: Person_U64NullUp(maxint64Bound),
			})
		require.Error(t, err)
	})
}
