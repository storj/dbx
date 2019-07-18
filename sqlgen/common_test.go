// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

import (
	"testing"
)

func TestFlattenSQL(t *testing.T) {
	for _, test := range []struct {
		in  string
		exp string
	}{
		{"", ""},
		{" ", ""},
		{" x ", "x"},
		{" x\t\t", "x"},
		{"  x ", "x"},
		{"\t\tx\t\t", "x"},
		{" \tx \t", "x"},
		{"\t x\t ", "x"},
		{"x\tx", "x x"},
		{"  x  ", "x"},
		{" \tx x\t ", "x x"},
		{"  x  x  x  ", "x x x"},
		{"\t\tx\t\tx\t\tx\t\t", "x x x"},
		{"  x  \n\t    x  \n\t   x  ", "x x x"},
	} {
		got := flattenSQL(test.in)
		if got != test.exp {
			t.Logf(" in: %q", test.in)
			t.Logf("got: %q", got)
			t.Logf("exp: %q", test.exp)
			t.Fail()
		}
	}
}

var benchStrings = []string{
	`INSERT INTO example ( alpha, beta, gamma, delta, iota, kappa, lambda ) VALUES ( $1, $2, $3, $4, $5, $6, $7 ) RETURNING example.alpha, example.beta, example.gamma, example.delta, example.iota, example.kappa, example.lambda;`,
	`INSERT INTO example
	 ( alpha, beta,
	 	gamma, delta, iota,

	 	 kappa, lambda ) VALUES ( $1, $2, $3, $4,
	 	 	$5, $6, $7 ) RETURNING example.alpha,
	 	 	example.beta, example.gamma,
	 	 	example.delta, example.iota, example.kappa,

	 	 	example.lambda;`,
}

func BenchmarkFlattenSQL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range benchStrings {
			_ = flattenSQL(s)
		}
	}
}
