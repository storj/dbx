// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

type Root struct {
	Models  []*Model
	Creates []*Create
	Reads   []*Read
	Updates []*Update
	Deletes []*Delete
}
