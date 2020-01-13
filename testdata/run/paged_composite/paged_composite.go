package main

import (
	"context"
	"fmt"
	"time"
)

func erre(err error) {
	if err != nil {
		panic(err)
	}
}

func eq(a, b interface{}) {
	if a != b {
		panic("invalid")
	}
}

var ctx = context.Background()

func main() {
	db, err := Open("sqlite3", ":memory:")
	erre(err)
	defer db.Close()

	_, err = db.Exec(db.Schema())
	erre(err)

	for i := 0; i < 1000; i++ {
		err = db.CreateNoReturn_ConsumedSerial(ctx,
			ConsumedSerial_ExpiresAt(time.Now()),
			ConsumedSerial_StorageNodeId([]byte(fmt.Sprintf("node%d", i))),
			ConsumedSerial_ProjectId([]byte(fmt.Sprintf("proj%d", i))),
			ConsumedSerial_BucketName([]byte(fmt.Sprintf("bucket%d", i))),
			ConsumedSerial_Action(1),
			ConsumedSerial_SerialNumber([]byte(fmt.Sprintf("serial%d", i))),
			ConsumedSerial_Settled(100))
		erre(err)
	}

	var total int
	var rows []*ConsumedSerial
	var next *Paged_ConsumedSerial_By_ExpiresAt_Greater_Continuation
again:
	rows, next, err = db.Paged_ConsumedSerial_By_ExpiresAt_Greater(ctx,
		ConsumedSerial_ExpiresAt(time.Now().Add(-time.Minute)),
		10, next)
	erre(err)
	total += len(rows)

	if next != nil {
		goto again
	}
	eq(total, 1000)
}
