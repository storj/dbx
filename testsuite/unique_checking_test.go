// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/unique_checking"
	"storj.io/dbx/testutil"
)

func TestUniqueChecking(t *testing.T) {
	ctx := context.Background()
	testutil.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		testutil.RecreateSchema(t, db)

		a, err := db.Create_A(ctx)
		require.NoError(t, err)

		b1, err := db.Create_B(ctx, B_AId(a.Id))
		require.NoError(t, err)
		c1, err := db.Create_C(ctx, C_Lat(0.0), C_Lon(0.0), C_BId(b1.Id))
		require.NoError(t, err)

		b2, err := db.Create_B(ctx, B_AId(a.Id))
		require.NoError(t, err)
		c2, err := db.Create_C(ctx, C_Lat(1.0), C_Lon(1.0), C_BId(b2.Id))
		require.NoError(t, err)

		rows, err := db.All_A_B_C_By_A_Id_And_C_Lat_Less_And_C_Lat_Greater_And_C_Lon_Less_And_C_Lon_Greater(ctx,
			A_Id(a.Id),
			C_Lat(10.0), C_Lat(-10.0),
			C_Lon(10.0), C_Lon(-10.0))
		require.NoError(t, err)

		require.Len(t, rows, 2)

		require.Equal(t, rows[0].A.Id, a.Id)
		require.Equal(t, rows[0].B.Id, b1.Id)
		require.Equal(t, rows[0].C.Id, c1.Id)
		require.Equal(t, float32(0), rows[0].C.Lat)
		require.Equal(t, float32(0), rows[0].C.Lon)

		require.Equal(t, rows[1].A.Id, a.Id)
		require.Equal(t, rows[1].B.Id, b2.Id)
		require.Equal(t, rows[1].C.Id, c2.Id)
		require.Equal(t, float32(1), rows[1].C.Lat)
		require.Equal(t, float32(1), rows[1].C.Lon)

		_, err = db.Create_D(ctx, D_A(0), D_B(0), D_C(0))
		require.NoError(t, err)
		// don't allow the same A
		_, err = db.Create_D(ctx, D_A(0), D_B(1), D_C(1))
		requireConstraintError(t, err)

		// don't allow same B, C
		_, err = db.Create_D(ctx, D_A(1), D_B(0), D_C(0))
		requireConstraintError(t, err)
	})
}

func requireConstraintError(t testing.TB, err error) {
	var internal *Error
	require.True(t, errors.As(err, &internal))
	require.Equal(t, ErrorCode_ConstraintViolation, internal.Code)
}
