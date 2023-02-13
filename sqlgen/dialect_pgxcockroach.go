// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type pgxcockroach struct{}

func (p pgxcockroach) Rebind(sql string) string {
	return postgres{}.Rebind(sql)
}
