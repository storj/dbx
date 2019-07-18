// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import "fmt"

type Update struct {
	Suffix   []string
	Model    *Model
	Joins    []*Join
	Where    []*Where
	NoReturn bool
}

func (r *Update) Signature() string {
	prefix := "UPDATE"
	if r.NoReturn {
		prefix += "_NORETURN"
	}
	return fmt.Sprintf("%s(%q)", prefix, r.Suffix)
}

func (upd *Update) AutoUpdatableFields() (fields []*Field) {
	return upd.Model.AutoUpdatableFields()
}

func (upd *Update) One() bool {
	return queryUnique([]*Model{upd.Model}, upd.Joins, upd.Where)
}
