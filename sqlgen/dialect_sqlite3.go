// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package sqlgen

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type sqlite3 struct{}

func newsqlite3() sqlite3 {
	return sqlite3{}
}

func (p sqlite3) Scanner(dest interface{}) interface{} {
	return dest
}

func (s sqlite3) Rebind(sql string) string {
	return sql
}
