// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package joins

import (
	"context"
	"storj.io/dbx/testrun"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoin(t *testing.T) {
	ctx := context.Background()
	testrun.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {

		if testrun.IsSpanner[*DB](db.DB) {
			t.Skip("TODO: REFERENCES syntax is not valid for Spanner Google SQL")
		}

		testrun.RecreateSchema(t, db)

		user, err := db.Create_User(ctx)
		require.NoError(t, err)

		aa, err := db.Create_AssociatedAccount(ctx,
			AssociatedAccount_UserPk(user.Pk))
		require.NoError(t, err)

		sess, err := db.Create_Session(ctx,
			Session_UserPk(user.Pk))
		require.NoError(t, err)

		rows, err := db.All_Session_Id_By_AssociatedAccount_Pk(ctx,
			AssociatedAccount_Pk(aa.Pk))
		require.NoError(t, err)

		if len(rows) != 1 || rows[0].Id != sess.Id {
			panic("invalid")
		}
	})

}
