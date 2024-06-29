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
		ctx := context.Background()

		testrun.RecreateSchema(t, db)

		//Create Person
		person, err := db.Create_Person(ctx, Person_A("P1"), Person_B(12345), Person_D(5678), Person_Create_Fields{
			C: Person_C(4567),
			E: Person_E_Null(),
		})
		require.NoError(t, err)

		//Read Person
		row, err := db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.Equal(t, person.A, row.A)
		require.Equal(t, person.B, row.B)
		require.Equal(t, person.C, row.C)
		require.Equal(t, person.D, row.D)
		require.Equal(t, person.E, row.E)

		//Read Person
		row, err = db.Get_Person_By_D_And_E(ctx, Person_D(person.D), Person_E_Raw(person.E))
		require.NoError(t, err)
		require.Equal(t, person.Pk, row.Pk)
		require.Equal(t, person.A, row.A)
		require.Equal(t, person.B, row.B)
		require.Equal(t, person.C, row.C)

		//Update Person
		person.B = 54321
		person.C = nil
		row, err = db.Update_Person_By_Pk_And_D_And_E(ctx, Person_Pk(person.Pk), Person_D(person.D), Person_E_Raw(person.E), Person_Update_Fields{
			B: Person_B(54321),
			C: Person_C_Null(),
		})
		require.NoError(t, err)
		require.Equal(t, person.B, row.B)
		require.Equal(t, person.C, row.C)

		//Delete Person
		deleted, err := db.Delete_Person_By_Pk_And_D_And_E(ctx, Person_Pk(person.Pk), Person_D(person.D), Person_E_Raw(person.E))
		require.NoError(t, err)
		require.Equal(t, true, deleted)

		//Bounds checks
		maxint64Bound := uint64(math.MaxInt64 + 1)
		_, err = db.Create_Person(ctx, Person_A("P1"), Person_B(12345), Person_D(maxint64Bound), Person_Create_Fields{
			C: Person_C(4567),
			E: Person_E_Null(),
		})
		require.Error(t, err)

		//Bounds checks
		_, err = db.Create_Person(ctx, Person_A("P1"), Person_B(12345), Person_D(5678), Person_Create_Fields{
			C: Person_C(4567),
			E: Person_E(maxint64Bound),
		})
		require.Error(t, err)
	})
}
