// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import "fmt"

type Delete struct {
	Suffix []string
	Model  *Model
	Joins  []*Join
	Where  []*Where
}

func (r *Delete) Signature() string {
	return fmt.Sprintf("DELETE(%q)", r.Suffix)
}

func (d *Delete) Distinct() bool {
	return queryUnique([]*Model{d.Model}, d.Joins, d.Where)
}
