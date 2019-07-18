// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package code

import (
	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
)

type Renderer interface {
	RenderCode(root *ir.Root, dialects []sql.Dialect) ([]byte, error)
}
