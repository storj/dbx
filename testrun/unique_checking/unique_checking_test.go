// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package unique_checking

import (
	"context"
	"storj.io/dbx/testrun"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniqueChecking(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {

		if testrun.IsSpanner[*DB](db.DB) {
			t.Skip("TODO: REFERENCES syntax is not valid for Spanner Google SQL")
		}

		for _, stmt := range db.DropSchema() {
			_, _ = db.Exec(stmt)
		}

		for _, stmt := range db.Schema() {
			_, err := db.Exec(stmt)
			require.NoError(t, err)
		}

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
	})
}
