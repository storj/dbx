package main

import "context"

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

	err = db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val0"))
	erre(err)
	row, err := db.Get_Kv_By_Key(ctx, Kv_Key("key"))
	erre(err)
	eq(row.Val, "val0")

	err = db.ReplaceNoReturn_Kv(ctx, Kv_Key("key"), Kv_Val("val1"))
	erre(err)
	row, err = db.Get_Kv_By_Key(ctx, Kv_Key("key"))
	erre(err)
	eq(row.Val, "val1")

	rows, err := db.All_Kv(ctx)
	erre(err)
	eq(len(rows), 1)
}
