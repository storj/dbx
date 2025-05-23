// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/wrap_type_json"
	"storj.io/dbx/testutil"
)

func TestWrapTypeJson(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(ctx context.Context, t *testing.T, db *DB) {
		testutil.RecreateSchema(ctx, t, db)

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
		require.Equal(t, []byte("100"), person.Value)
		require.Equal(t, []byte("101"), person.ValueUp)
		require.Equal(t, []byte("102"), person.ValueNull)
		require.Equal(t, []byte("103"), person.ValueNullUp)
		// check the default value specified for a field is set and scanned properly
		require.Equal(t, []byte("{}"), person.ValueDefault)

		// Read Person
		row, err := db.Get_Person_By_Pk(ctx, Person_Pk(person.Pk))
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Update Person
		row, err = db.Update_Person_By_Pk(ctx,
			Person_Pk(person.Pk),
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
		row, err = db.Update_Person_By_Pk(ctx,
			Person_Pk(person.Pk),
			Person_Update_Fields{
				ValueNullUp: Person_ValueNullUp_Null(),
			},
		)
		person.ValueNullUp = nil
		require.NoError(t, err)
		require.EqualValues(t, person, row)

		// Delete Person
		deleted, err := db.Delete_Person_By_Pk(ctx,
			Person_Pk(row.Pk),
		)
		require.NoError(t, err)
		require.Equal(t, true, deleted)
	})
}
