// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

type Index struct {
	Name    string
	Model   *Model
	Fields  []*Field
	Unique  bool
	Where   []*Where
	Storing []*Field
}
