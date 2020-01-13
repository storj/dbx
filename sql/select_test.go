package sql

import (
	"fmt"
	"testing"

	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen/sqlcompile"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

func TestPagedWhereFromPK(t *testing.T) {
	pk := []*ir.Field{
		{Model: &ir.Model{Table: "t"}, Column: "a1", Nullable: true},
		{Model: &ir.Model{Table: "t"}, Column: "a2"},
		{Model: &ir.Model{Table: "t"}, Column: "a3", Nullable: true},
		{Model: &ir.Model{Table: "t"}, Column: "a4"},
	}

	where := pagedWhereFromPK(pk)
	sqls := WhereSQL([]*ir.Where{where}, SQLite3())
	for _, sql := range sqls {
		fmt.Printf("%+v\n", sqlembedgo.Embed("prefix", sqlcompile.Compile(sql)))
	}
}
