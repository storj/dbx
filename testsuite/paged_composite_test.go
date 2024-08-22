// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "storj.io/dbx/testsuite/generated/paged_composite"
	"storj.io/dbx/testutil"
)

func TestPagedComposite(t *testing.T) {
	ctx := context.Background()
	testutil.RunDBTest[*DB](t, Open, func(t *testing.T, db *DB) {

		testutil.RecreateSchema(t, db)

		for i := 0; i < 1000; i++ {
			err := db.CreateNoReturn_ConsumedSerial(ctx,
				ConsumedSerial_ExpiresAt(time.Now()),
				ConsumedSerial_StorageNodeId([]byte(fmt.Sprintf("node%d", i))),
				ConsumedSerial_ProjectId([]byte(fmt.Sprintf("proj%d", i))),
				ConsumedSerial_BucketName([]byte(fmt.Sprintf("bucket%d", i))),
				ConsumedSerial_Action(1),
				ConsumedSerial_SerialNumber([]byte(fmt.Sprintf("serial%d", i))),
				ConsumedSerial_Settled(100))
			require.NoError(t, err)
		}

		var total int
		var rows []*ConsumedSerial
		var next *Paged_ConsumedSerial_By_ExpiresAt_Greater_Continuation
	again:
		rows, next, err := db.Paged_ConsumedSerial_By_ExpiresAt_Greater(ctx,
			ConsumedSerial_ExpiresAt(time.Now().Add(-time.Minute)),
			10, next)
		require.NoError(t, err)
		total += len(rows)

		if next != nil {
			goto again
		}
		require.Equal(t, 1000, total)
	})
}
