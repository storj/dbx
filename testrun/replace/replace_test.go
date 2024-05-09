// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package replace

import (
	"context"
	"storj.io/dbx/testrun"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplace(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {

		if testrun.IsSpanner[*DB](db.DB) {
			t.Skip("TODO: ON CONFLICT has different syntax with Spanner Google SQL")
		}

		for _, stmt := range db.DropSchema() {
			_, _ = db.Exec(stmt)
		}

		for _, stmt := range db.Schema() {
			_, err := db.Exec(stmt)
			require.NoError(t, err)
		}

		err := db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val0"))
		require.NoError(t, err)
		row, err := db.Get_Kv_By_Key(ctx, Kv_Key("key"))
		require.NoError(t, err)
		require.Equal(t, "val0", row.Val)

		err = db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val1"))
		require.NoError(t, err)
		row, err = db.Get_Kv_By_Key(ctx, Kv_Key("key"))
		require.NoError(t, err)
		require.Equal(t, "val1", row.Val)

		rows, err := db.All_Kv(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 1)
	})
}
