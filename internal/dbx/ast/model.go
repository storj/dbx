// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"fmt"

	"bitbucket.org/pkg/inflect"
)

type Model struct {
	Name       string
	Table      string
	Fields     []*Field
	PrimaryKey []*Field
	Unique     [][]*Field
	Indexes    []*Index
}

func (m *Model) TableName() string {
	if m.Table != "" {
		return m.Table
	}
	return inflect.Pluralize(m.Name)
}

func (m *Model) BasicPrimaryKey() *Field {
	if len(m.PrimaryKey) == 1 && m.PrimaryKey[0].IsInt() {
		return m.PrimaryKey[0]
	}
	return nil
}

func (m *Model) InsertableFields() (fields []*Field) {
	for _, field := range m.Fields {
		if field.Insertable() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Model) SelectRefs() (refs []string) {
	for _, field := range m.Fields {
		refs = append(refs,
			fmt.Sprintf("%s.%s", m.TableName(), field.ColumnName()))
	}
	return refs
}

func (m *Model) selectable() {}