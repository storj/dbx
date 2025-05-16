// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/replace"
	"storj.io/dbx/testutil"
)

func TestReplace(t *testing.T) {
	testutil.RunDBTest[*DB](t, Open, func(ctx context.Context, t *testing.T, db *DB) {
		testutil.RecreateSchema(ctx, t, db)

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
