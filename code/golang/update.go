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

package golang

import (
	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

type Update struct {
	Suffix            string
	Struct            *ModelStruct
	Return            *Var
	Args              []*Var
	AutoFields        []*Var
	SQLPrefix         string
	SQLSuffix         string
	SupportsReturning bool
	NeedsNow          bool
}

func UpdateFromIR(ir_upd *ir.Update, dialect sql.Dialect) *Update {
	sql_prefix, sql_suffix := sql.RenderUpdate(dialect, ir_upd)
	upd := &Update{
		Suffix:            inflect.Camelize(ir_upd.FuncSuffix()),
		Struct:            ModelStructFromIR(ir_upd.Model),
		SQLPrefix:         sql_prefix,
		SQLSuffix:         sql_suffix,
		Return:            VarFromModel(ir_upd.Model),
		SupportsReturning: dialect.Features().Returning,
	}

	for _, where := range ir_upd.Where {
		if where.Right == nil {
			upd.Args = append(upd.Args, VarFromField(where.Left))
		}
	}

	for _, field := range ir_upd.AutoFields() {
		upd.NeedsNow = upd.NeedsNow || field.IsTime()
		upd.AutoFields = append(upd.AutoFields, VarFromField(field))
	}

	return upd
}
