// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package replace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/testrun"
)

func TestReplace(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {
		testrun.RecreateSchema(t, db)

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
