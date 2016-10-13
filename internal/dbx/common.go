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

package dbx

import (
	"io"

	"github.com/spacemonkeygo/errors"
)

var Error = errors.NewClass("dbx")

type Language interface {
	Name() string
	Render(w io.Writer, schema *Schema) error
	AddSelect(w io.Writer, sql string, params *SelectParams)
	AddCount(w io.Writer, sql string, params *SelectParams)
	AddDelete(w io.Writer, sql string, params *DeleteParams)
	AddInsert(w io.Writer, sql string, params *InsertParams)
	AddUpdate(w io.Writer, sql string, params *UpdateParams)
	Format([]byte) ([]byte, error)
}

type Dialect interface {
	Name() string
	ColumnName(column *Column) string
	ListTablesSQL() string
	RenderSchema(schema *Schema) (string, error)
	RenderSelect(params *SelectParams) (string, error)
	RenderCount(params *SelectParams) (string, error)
	RenderDelete(params *DeleteParams) (string, error)
	RenderInsert(params *InsertParams) (string, error)
	RenderUpdate(params *UpdateParams) (string, error)
	SupportsReturning() bool
}
