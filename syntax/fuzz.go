// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

//go:build gofuzz
// +build gofuzz

package syntax

func Fuzz(data []byte) int {
	if _, err := Parse("", data); err != nil {
		return 0
	}
	return 1
}
