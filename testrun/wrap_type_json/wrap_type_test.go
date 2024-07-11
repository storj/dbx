// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package wrap_type

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/testrun"
)

func Test(t *testing.T) {
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		if testrun.IsSpanner[*DB](db.DB) {
			t.Skip("spanner dialect does not handle json, yet")
		}

		ctx := context.Background()

		testrun.RecreateSchema(t, db)

		// Create Person
		person, err := db.Create_Person(ctx,
			Person_Name("P1"),
			Person_Value([]byte("100")),
			Person_ValueUp([]byte("101")),
			Person_Create_Fields{
				ValueNull:   Person_ValueNull([]byte("102")),
				ValueNullUp: Person_ValueNullUp([]byte("103")),
			})
		require.NoError(t, err)

		// Read Person
		row, err := db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Read Person
		row, err = db.Get_Person_By_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Value([]byte("100")),
			Person_ValueUp([]byte("101")),
			Person_ValueNull([]byte("102")),
			Person_ValueNullUp([]byte("103")),
		)
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		row, err = db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person
		row, err = db.Update_Person_By_Pk_And_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Pk(person.Pk),
			Person_Value([]byte("100")),
			Person_ValueUp([]byte("101")),
			Person_ValueNull([]byte("102")),
			Person_ValueNullUp([]byte("103")),
			Person_Update_Fields{
				ValueUp:     Person_ValueUp([]byte("111")),
				ValueNullUp: Person_ValueNullUp([]byte("113")),
			},
		)
		person.ValueUp = []byte("111")
		person.ValueNullUp = []byte("113")
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person with nil
		row, err = db.Update_Person_By_Pk_And_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Pk(person.Pk),
			Person_Value([]byte("100")),
			Person_ValueUp([]byte("111")),
			Person_ValueNull([]byte("102")),
			Person_ValueNullUp([]byte("113")),
			Person_Update_Fields{
				ValueNullUp: Person_ValueNullUp_Null(),
			},
		)
		person.ValueNullUp = nil
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Delete Person
		deleted, err := db.Delete_Person_By_Value_And_ValueUp_And_ValueNull_And_ValueNullUp(ctx,
			Person_Value([]byte("100")),
			Person_ValueUp([]byte("111")),
			Person_ValueNull([]byte("102")),
			Person_ValueNullUp_Null(),
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), deleted)
	})
}
