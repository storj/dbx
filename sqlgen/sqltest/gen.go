// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqltest

import (
	"fmt"
	"math/rand"
	"time"

	"storj.io/dbx/sqlgen"
	"storj.io/dbx/testutil"
)

type Generator struct {
	rng   *rand.Rand
	conds []*sqlgen.Condition
	holes []*sqlgen.Hole
}

func NewGenerator(tw *testutil.T) *Generator {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	tw.Logf("seed: %d", seed)
	return &Generator{
		rng: rng,
	}
}

func (g *Generator) Gen() (out sqlgen.SQL) { return g.genRecursive(3) }

func (g *Generator) literal() sqlgen.Literal {
	return sqlgen.Literal(fmt.Sprintf("(literal %d)", g.rng.Intn(1000)))
}

func (g *Generator) condition() *sqlgen.Condition {
	if len(g.conds) == 0 || rand.Intn(2) == 0 {
		num := len(g.conds)
		condition := &sqlgen.Condition{
			Name:  fmt.Sprintf("cond%d", num),
			Left:  fmt.Sprintf("left%d", num),
			Right: fmt.Sprintf("right%d", num),
			Null:  rand.Intn(2) == 0,
			Equal: rand.Intn(2) == 0,
		}
		g.conds = append(g.conds, condition)
		return condition
	}
	return g.conds[rand.Intn(len(g.conds))]
}

func (g *Generator) hole() *sqlgen.Hole {
	if len(g.holes) == 0 || rand.Intn(2) == 0 {
		num := len(g.holes)
		hole := &sqlgen.Hole{
			Name: fmt.Sprintf("hole%d", num),
			SQL:  g.genRecursive(1),
		}
		g.holes = append(g.holes, hole)
		return hole
	}
	return g.holes[rand.Intn(len(g.holes))]
}

func (g *Generator) literals(depth int) sqlgen.Literals {
	amount := rand.Intn(30)

	sqls := make([]sqlgen.SQL, amount)
	for i := range sqls {
		sqls[i] = g.genRecursive(depth - 1)
	}

	join := fmt.Sprintf("|join %d|", g.rng.Intn(1000))
	if rand.Intn(2) == 0 {
		join = ""
	}

	return sqlgen.Literals{
		Join: join,
		SQLs: sqls,
	}
}

func (g *Generator) genRecursive(depth int) (out sqlgen.SQL) {
	if depth == 0 {
		return g.literal()
	}

	switch g.rng.Intn(10) {
	case 0, 1:
		return g.literal()
	case 2, 3:
		return g.condition()
	case 4, 5:
		return g.hole()
	default:
		return g.literals(depth)
	}
}
