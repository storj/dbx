// Copyright (C) 2017 Space Monkey, Inc.
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

package sql

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlcompile"
	. "storj.io/dbx/sqlgen/sqlhelpers"
)

func DeleteSQL(ir_del *ir.Delete, dialect Dialect) sqlgen.SQL {
	stmt := Build(Lf("DELETE FROM %s", ir_del.Model.Table))

	var wheres []sqlgen.SQL
	if len(ir_del.Joins) == 0 {
		wheres = WhereSQL(ir_del.Where, dialect)
	} else {
		pk_column := ir_del.Model.PrimaryKey[0].ColumnRef()
		sel := SelectSQL(&ir.Read{
			View:        ir.All,
			From:        ir_del.Model,
			Selectables: []ir.Selectable{ir_del.Model.PrimaryKey[0]},
			Joins:       ir_del.Joins,
			Where:       ir_del.Where,
		}, dialect)
		wheres = append(wheres, J("", L(pk_column), L(" IN ("), sel, L(")")))
	}

	if len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	return sqlcompile.Compile(stmt.SQL())

}
