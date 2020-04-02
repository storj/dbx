package main

import (
	"context"
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

	err = db.CreateNoReturn_Foo(ctx, Foo_A("a"), Foo_B("b"), Foo_C("c"))
	erre(err)

	rows, err := db.All_Foo_By__A_Or__B_Or_C(ctx, Foo_A("x"), Foo_B("x"), Foo_C("c"))
	erre(err)
	eq(len(rows), 1)
	eq(*rows[0], Foo{Pk: 1, A: "a", B: "b", C: "c"})
}
