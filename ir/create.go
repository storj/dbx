// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package ir

import "fmt"

type Create struct {
	Suffix   []string
	Model    *Model
	Raw      bool
	NoReturn bool
	Replace  bool
}

func (cre *Create) Signature() string {
	prefix := "CREATE"
	if cre.Raw {
		prefix += "_RAW"
	}
	if cre.NoReturn {
		prefix += "_NORETURN"
	}
	if cre.Replace {
		prefix += "_REPLACE"
	}
	return fmt.Sprintf("%s(%q)", prefix, cre.Suffix)
}

func (cre *Create) Fields() (fields []*Field) {
	return cre.Model.Fields
}

func (cre *Create) InsertableStaticFields() (fields []*Field) {
	if cre.Raw {
		return cre.Model.Fields
	}
	return cre.Model.InsertableStaticFields()
}

func (cre *Create) InsertableDynamicFields() (fields []*Field) {
	if cre.Raw {
		return nil
	}
	return cre.Model.InsertableDynamicFields()
}

func (cre *Create) InsertableRequiredFields() (fields []*Field) {
	if cre.Raw {
		return cre.Model.Fields
	}
	return cre.Model.InsertableRequiredFields()
}

func (cre *Create) InsertableOptionalFields() (fields []*Field) {
	if cre.Raw {
		return nil
	}
	return cre.Model.InsertableOptionalFields()
}
