// AUTOGENERATED BY storj.io/dbx
// DO NOT EDIT.

package partial_index

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"cloud.google.com/go/spanner"
	"crypto/rand"
	"encoding/base64"
	_ "github.com/googleapis/go-sql-spanner"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mattn/go-sqlite3"
)

// Prevent conditional imports from causing build failures.
var _ = strconv.Itoa
var _ = strings.LastIndex
var _ = fmt.Sprint
var _ sync.Mutex

var (
	WrapErr = func(err *Error) error { return err }
	Logger  func(format string, args ...any)

	errTooManyRows       = errors.New("too many rows")
	errUnsupportedDriver = errors.New("unsupported driver")
	errEmptyUpdate       = errors.New("empty update")
)

func logError(format string, args ...any) {
	if Logger != nil {
		Logger(format, args...)
	}
}

type ErrorCode int

const (
	ErrorCode_Unknown ErrorCode = iota
	ErrorCode_UnsupportedDriver
	ErrorCode_NoRows
	ErrorCode_TxDone
	ErrorCode_TooManyRows
	ErrorCode_ConstraintViolation
	ErrorCode_EmptyUpdate
)

type Error struct {
	Err         error
	Code        ErrorCode
	Driver      string
	Constraint  string
	QuerySuffix string
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrapErr(e *Error) error {
	if WrapErr == nil {
		return e
	}
	return WrapErr(e)
}

func makeErr(err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return wrapErr(e)
	}
	e = &Error{Err: err}
	switch err {
	case sql.ErrNoRows:
		e.Code = ErrorCode_NoRows
	case sql.ErrTxDone:
		e.Code = ErrorCode_TxDone
	}
	return wrapErr(e)
}

func unsupportedDriver(driver string) error {
	return wrapErr(&Error{
		Err:    errUnsupportedDriver,
		Code:   ErrorCode_UnsupportedDriver,
		Driver: driver,
	})
}

func emptyUpdate() error {
	return wrapErr(&Error{
		Err:  errEmptyUpdate,
		Code: ErrorCode_EmptyUpdate,
	})
}

func tooManyRows(query_suffix string) error {
	return wrapErr(&Error{
		Err:         errTooManyRows,
		Code:        ErrorCode_TooManyRows,
		QuerySuffix: query_suffix,
	})
}

func constraintViolation(err error, constraint string) error {
	return wrapErr(&Error{
		Err:        err,
		Code:       ErrorCode_ConstraintViolation,
		Constraint: constraint,
	})
}

type driver interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DB struct {
	*sql.DB
	dbMethods

	Hooks struct {
		Now func() time.Time
	}
}

func Open(driver, source string) (db *DB, err error) {
	var sql_db *sql.DB
	switch driver {
	case "sqlite3":
		sql_db, err = opensqlite3(source)
	case "pgx":
		sql_db, err = openpgx(source)
	case "pgxcockroach":
		sql_db, err = openpgxcockroach(source)
	case "spanner":
		sql_db, err = openspanner(source)
	default:
		return nil, unsupportedDriver(driver)
	}
	if err != nil {
		return nil, makeErr(err)
	}
	defer func(sql_db *sql.DB) {
		if err != nil {
			_ = sql_db.Close()
		}
	}(sql_db)

	if err := sql_db.Ping(); err != nil {
		return nil, makeErr(err)
	}

	db = &DB{
		DB: sql_db,
	}
	db.Hooks.Now = time.Now

	switch driver {
	case "sqlite3":
		db.dbMethods = newsqlite3(db)
	case "pgx":
		db.dbMethods = newpgx(db)
	case "pgxcockroach":
		db.dbMethods = newpgxcockroach(db)
	case "spanner":
		db.dbMethods = newspanner(db)
	default:
		return nil, unsupportedDriver(driver)
	}

	return db, nil
}

func (obj *DB) Close() (err error) {
	return obj.makeErr(obj.DB.Close())
}

func (obj *DB) Open(ctx context.Context) (*Tx, error) {
	tx, err := obj.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, obj.makeErr(err)
	}

	return &Tx{
		Tx:        tx,
		txMethods: obj.wrapTx(tx),
	}, nil
}

func DeleteAll(ctx context.Context, db *DB) (int64, error) {
	tx, err := db.Open(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err == nil {
			err = db.makeErr(tx.Commit())
			return
		}

		if err_rollback := tx.Rollback(); err_rollback != nil {
			logError("delete-all: rollback failed: %v", db.makeErr(err_rollback))
		}
	}()
	return tx.deleteAll(ctx)
}

type Tx struct {
	Tx *sql.Tx
	txMethods
}

type dialectTx struct {
	tx *sql.Tx
}

func (tx *dialectTx) Commit() (err error) {
	return makeErr(tx.tx.Commit())
}

func (tx *dialectTx) Rollback() (err error) {
	return makeErr(tx.tx.Rollback())
}

type sqlite3Impl struct {
	db      *DB
	dialect __sqlbundle_sqlite3
	driver  driver
	txn     bool
}

func (obj *sqlite3Impl) Rebind(s string) string {
	return obj.dialect.Rebind(s)
}

func (obj *sqlite3Impl) logStmt(stmt string, args ...any) {
	sqlite3LogStmt(stmt, args...)
}

func (obj *sqlite3Impl) makeErr(err error) error {
	constraint, ok := obj.isConstraintError(err)
	if ok {
		return constraintViolation(err, constraint)
	}
	return makeErr(err)
}

type sqlite3DB struct {
	db *DB
	*sqlite3Impl
}

func newsqlite3(db *DB) *sqlite3DB {
	return &sqlite3DB{
		db: db,
		sqlite3Impl: &sqlite3Impl{
			db:     db,
			driver: db.DB,
		},
	}
}

func (obj *sqlite3DB) Schema() []string {
	return []string{

		`CREATE TABLE as (
	pk INTEGER NOT NULL,
	ctime TIMESTAMP NOT NULL,
	mtime TIMESTAMP NOT NULL,
	id TEXT NOT NULL,
	name TEXT NOT NULL,
	PRIMARY KEY ( pk )
)`,

		`CREATE INDEX as_ctime_mtime_index ON as ( ctime, mtime ) WHERE as.id > 'foo' AND as.name > 'bar'`,
	}
}

func (obj *sqlite3DB) DropSchema() []string {
	return []string{

		`DROP TABLE IF EXISTS as`,
	}
}

func (obj *sqlite3DB) wrapTx(tx *sql.Tx) txMethods {
	return &sqlite3Tx{
		dialectTx: dialectTx{tx: tx},
		sqlite3Impl: &sqlite3Impl{
			db:     obj.db,
			driver: tx,
			txn:    true,
		},
	}
}

type sqlite3Tx struct {
	dialectTx
	*sqlite3Impl
}

func sqlite3LogStmt(stmt string, args ...any) {
	// TODO: render placeholders
	if Logger != nil {
		out := fmt.Sprintf("stmt: %s\nargs: %v\n", stmt, pretty(args))
		Logger(out)
	}
}

type pgxImpl struct {
	db      *DB
	dialect __sqlbundle_pgx
	driver  driver
	txn     bool
}

func (obj *pgxImpl) Rebind(s string) string {
	return obj.dialect.Rebind(s)
}

func (obj *pgxImpl) logStmt(stmt string, args ...any) {
	pgxLogStmt(stmt, args...)
}

func (obj *pgxImpl) makeErr(err error) error {
	constraint, ok := obj.isConstraintError(err)
	if ok {
		return constraintViolation(err, constraint)
	}
	return makeErr(err)
}

type pgxDB struct {
	db *DB
	*pgxImpl
}

func newpgx(db *DB) *pgxDB {
	return &pgxDB{
		db: db,
		pgxImpl: &pgxImpl{
			db:     db,
			driver: db.DB,
		},
	}
}

func (obj *pgxDB) Schema() []string {
	return []string{

		`CREATE TABLE as (
	pk bigserial NOT NULL,
	ctime timestamp with time zone NOT NULL,
	mtime timestamp with time zone NOT NULL,
	id text NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( pk )
)`,

		`CREATE INDEX as_ctime_mtime_index ON as ( ctime, mtime ) WHERE as.id > 'foo' AND as.name > 'bar'`,
	}
}

func (obj *pgxDB) DropSchema() []string {
	return []string{

		`DROP TABLE IF EXISTS as`,
	}
}

func (obj *pgxDB) wrapTx(tx *sql.Tx) txMethods {
	return &pgxTx{
		dialectTx: dialectTx{tx: tx},
		pgxImpl: &pgxImpl{
			db:     obj.db,
			driver: tx,
			txn:    true,
		},
	}
}

type pgxTx struct {
	dialectTx
	*pgxImpl
}

func pgxLogStmt(stmt string, args ...any) {
	// TODO: render placeholders
	if Logger != nil {
		out := fmt.Sprintf("stmt: %s\nargs: %v\n", stmt, pretty(args))
		Logger(out)
	}
}

type pgxcockroachImpl struct {
	db      *DB
	dialect __sqlbundle_pgxcockroach
	driver  driver
	txn     bool
}

func (obj *pgxcockroachImpl) Rebind(s string) string {
	return obj.dialect.Rebind(s)
}

func (obj *pgxcockroachImpl) logStmt(stmt string, args ...any) {
	pgxcockroachLogStmt(stmt, args...)
}

func (obj *pgxcockroachImpl) makeErr(err error) error {
	constraint, ok := obj.isConstraintError(err)
	if ok {
		return constraintViolation(err, constraint)
	}
	return makeErr(err)
}

type pgxcockroachDB struct {
	db *DB
	*pgxcockroachImpl
}

func newpgxcockroach(db *DB) *pgxcockroachDB {
	return &pgxcockroachDB{
		db: db,
		pgxcockroachImpl: &pgxcockroachImpl{
			db:     db,
			driver: db.DB,
		},
	}
}

func (obj *pgxcockroachDB) Schema() []string {
	return []string{

		`CREATE TABLE as (
	pk bigserial NOT NULL,
	ctime timestamp with time zone NOT NULL,
	mtime timestamp with time zone NOT NULL,
	id text NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( pk )
)`,

		`CREATE INDEX as_ctime_mtime_index ON as ( ctime, mtime ) STORING ( id, name ) WHERE as.id > 'foo' AND as.name > 'bar'`,
	}
}

func (obj *pgxcockroachDB) DropSchema() []string {
	return []string{

		`DROP TABLE IF EXISTS as`,
	}
}

func (obj *pgxcockroachDB) wrapTx(tx *sql.Tx) txMethods {
	return &pgxcockroachTx{
		dialectTx: dialectTx{tx: tx},
		pgxcockroachImpl: &pgxcockroachImpl{
			db:     obj.db,
			driver: tx,
			txn:    true,
		},
	}
}

type pgxcockroachTx struct {
	dialectTx
	*pgxcockroachImpl
}

func pgxcockroachLogStmt(stmt string, args ...any) {
	// TODO: render placeholders
	if Logger != nil {
		out := fmt.Sprintf("stmt: %s\nargs: %v\n", stmt, pretty(args))
		Logger(out)
	}
}

type spannerImpl struct {
	db      *DB
	dialect __sqlbundle_spanner
	driver  driver
	txn     bool
}

func (obj *spannerImpl) Rebind(s string) string {
	return obj.dialect.Rebind(s)
}

func (obj *spannerImpl) logStmt(stmt string, args ...any) {
	spannerLogStmt(stmt, args...)
}

func (obj *spannerImpl) makeErr(err error) error {
	constraint, ok := obj.isConstraintError(err)
	if ok {
		return constraintViolation(err, constraint)
	}
	return makeErr(err)
}

type spannerDB struct {
	db *DB
	*spannerImpl
}

func newspanner(db *DB) *spannerDB {
	return &spannerDB{
		db: db,
		spannerImpl: &spannerImpl{
			db:     db,
			driver: db.DB,
		},
	}
}

func (obj *spannerDB) Schema() []string {
	return []string{

		`CREATE SEQUENCE as_pk OPTIONS (sequence_kind='bit_reversed_positive')`,

		`CREATE TABLE as (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE as_pk)),
	ctime TIMESTAMP NOT NULL,
	mtime TIMESTAMP NOT NULL,
	id STRING(MAX) NOT NULL,
	name STRING(MAX) NOT NULL
) PRIMARY KEY ( pk )`,

		`CREATE INDEX as_ctime_mtime_index ON as ( ctime, mtime )`,
	}
}

func (obj *spannerDB) DropSchema() []string {
	return []string{

		`DROP INDEX IF EXISTS as_ctime_mtime_index`,

		`ALTER TABLE  as ALTER pk SET DEFAULT (null)`,

		`DROP SEQUENCE IF EXISTS as_pk`,

		`DROP TABLE IF EXISTS as`,
	}
}

func (obj *spannerDB) wrapTx(tx *sql.Tx) txMethods {
	return &spannerTx{
		dialectTx: dialectTx{tx: tx},
		spannerImpl: &spannerImpl{
			db:     obj.db,
			driver: tx,
			txn:    true,
		},
	}
}

type spannerTx struct {
	dialectTx
	*spannerImpl
}

func spannerLogStmt(stmt string, args ...any) {
	// TODO: render placeholders
	if Logger != nil {
		out := fmt.Sprintf("stmt: %s\nargs: %v\n", stmt, pretty(args))
		Logger(out)
	}
}

type pretty []any

func (p pretty) Format(f fmt.State, c rune) {
	fmt.Fprint(f, "[")
nextval:
	for i, val := range p {
		if i > 0 {
			fmt.Fprint(f, ", ")
		}
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				fmt.Fprint(f, "NULL")
				continue
			}
			val = rv.Elem().Interface()
		}
		switch v := val.(type) {
		case string:
			fmt.Fprintf(f, "%q", v)
		case time.Time:
			fmt.Fprintf(f, "%s", v.Format(time.RFC3339Nano))
		case []byte:
			for _, b := range v {
				if !unicode.IsPrint(rune(b)) {
					fmt.Fprintf(f, "%#x", v)
					continue nextval
				}
			}
			fmt.Fprintf(f, "%q", v)
		default:
			fmt.Fprintf(f, "%v", v)
		}
	}
	fmt.Fprint(f, "]")
}

type A struct {
	Pk    int64
	Ctime time.Time
	Mtime time.Time
	Id    string
	Name  string
}

func (A) _Table() string { return "as" }

type A_Update_Fields struct {
}

type A_Pk_Field struct {
	_set   bool
	_null  bool
	_value int64
}

func A_Pk(v int64) A_Pk_Field {
	return A_Pk_Field{_set: true, _value: v}
}

func (f A_Pk_Field) value() any {
	if !f._set || f._null {
		return nil
	}
	return f._value
}

type A_Ctime_Field struct {
	_set   bool
	_null  bool
	_value time.Time
}

func A_Ctime(v time.Time) A_Ctime_Field {
	return A_Ctime_Field{_set: true, _value: v}
}

func (f A_Ctime_Field) value() any {
	if !f._set || f._null {
		return nil
	}
	return f._value
}

type A_Mtime_Field struct {
	_set   bool
	_null  bool
	_value time.Time
}

func A_Mtime(v time.Time) A_Mtime_Field {
	return A_Mtime_Field{_set: true, _value: v}
}

func (f A_Mtime_Field) value() any {
	if !f._set || f._null {
		return nil
	}
	return f._value
}

type A_Id_Field struct {
	_set   bool
	_null  bool
	_value string
}

func A_Id(v string) A_Id_Field {
	return A_Id_Field{_set: true, _value: v}
}

func (f A_Id_Field) value() any {
	if !f._set || f._null {
		return nil
	}
	return f._value
}

type A_Name_Field struct {
	_set   bool
	_null  bool
	_value string
}

func A_Name(v string) A_Name_Field {
	return A_Name_Field{_set: true, _value: v}
}

func (f A_Name_Field) value() any {
	if !f._set || f._null {
		return nil
	}
	return f._value
}

func toUTC(t time.Time) time.Time {
	return t.UTC()
}

func toDate(t time.Time) time.Time {
	// keep up the minute portion so that translations between timezones will
	// continue to reflect properly.
	return t.Truncate(time.Minute)
}

//
// runtime support for building sql statements
//

type __sqlbundle_SQL interface {
	Render() string

	private()
}

type __sqlbundle_Dialect interface {
	// Rebind gives the opportunity to rewrite provided SQL into a SQL dialect.
	Rebind(sql string) string
}

type __sqlbundle_RenderOp int

const (
	__sqlbundle_NoFlatten __sqlbundle_RenderOp = iota
	__sqlbundle_NoTerminate
)

func __sqlbundle_RenderAll(dialect __sqlbundle_Dialect, sqls []__sqlbundle_SQL, ops ...__sqlbundle_RenderOp) []string {
	var rs []string
	for _, sql := range sqls {
		rs = append(rs, __sqlbundle_Render(dialect, sql, ops...))
	}
	return rs
}

func __sqlbundle_Render(dialect __sqlbundle_Dialect, sql __sqlbundle_SQL, ops ...__sqlbundle_RenderOp) string {
	out := sql.Render()

	flatten := true
	terminate := true
	for _, op := range ops {
		switch op {
		case __sqlbundle_NoFlatten:
			flatten = false
		case __sqlbundle_NoTerminate:
			terminate = false
		}
	}

	if flatten {
		out = __sqlbundle_flattenSQL(out)
	}
	if terminate {
		out += ";"
	}

	return dialect.Rebind(out)
}

func __sqlbundle_flattenSQL(x string) string {
	// trim whitespace from beginning and end
	s, e := 0, len(x)-1
	for s < len(x) && (x[s] == ' ' || x[s] == '\t' || x[s] == '\n') {
		s++
	}
	for s <= e && (x[e] == ' ' || x[e] == '\t' || x[e] == '\n') {
		e--
	}
	if s > e {
		return ""
	}
	x = x[s : e+1]

	// check for whitespace that needs fixing
	wasSpace := false
	for i := 0; i < len(x); i++ {
		r := x[i]
		justSpace := r == ' '
		if (wasSpace && justSpace) || r == '\t' || r == '\n' {
			// whitespace detected, start writing a new string
			var result strings.Builder
			result.Grow(len(x))
			if wasSpace {
				result.WriteString(x[:i-1])
			} else {
				result.WriteString(x[:i])
			}
			for p := i; p < len(x); p++ {
				for p < len(x) && (x[p] == ' ' || x[p] == '\t' || x[p] == '\n') {
					p++
				}
				result.WriteByte(' ')

				start := p
				for p < len(x) && !(x[p] == ' ' || x[p] == '\t' || x[p] == '\n') {
					p++
				}
				result.WriteString(x[start:p])
			}

			return result.String()
		}
		wasSpace = justSpace
	}

	// no problematic whitespace found
	return x
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_cockroach struct{}

func (p __sqlbundle_cockroach) Rebind(sql string) string {
	return __sqlbundle_postgres{}.Rebind(sql)
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_pgx struct{}

func (p __sqlbundle_pgx) Rebind(sql string) string {
	return __sqlbundle_postgres{}.Rebind(sql)
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_pgxcockroach struct{}

func (p __sqlbundle_pgxcockroach) Rebind(sql string) string {
	return __sqlbundle_postgres{}.Rebind(sql)
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_postgres struct{}

func (p __sqlbundle_postgres) Rebind(sql string) string {
	type sqlParseState int
	const (
		sqlParseStart sqlParseState = iota
		sqlParseInStringLiteral
		sqlParseInQuotedIdentifier
		sqlParseInComment
	)

	out := make([]byte, 0, len(sql)+10)

	j := 1
	state := sqlParseStart
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch state {
		case sqlParseStart:
			switch ch {
			case '?':
				out = append(out, '$')
				out = append(out, strconv.Itoa(j)...)
				state = sqlParseStart
				j++
				continue
			case '-':
				if i+1 < len(sql) && sql[i+1] == '-' {
					state = sqlParseInComment
				}
			case '"':
				state = sqlParseInQuotedIdentifier
			case '\'':
				state = sqlParseInStringLiteral
			}
		case sqlParseInStringLiteral:
			if ch == '\'' {
				state = sqlParseStart
			}
		case sqlParseInQuotedIdentifier:
			if ch == '"' {
				state = sqlParseStart
			}
		case sqlParseInComment:
			if ch == '\n' {
				state = sqlParseStart
			}
		}
		out = append(out, ch)
	}

	return string(out)
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_spanner struct{}

func (p __sqlbundle_spanner) Rebind(sql string) string {
	return sql
}

// this type is specially named to match up with the name returned by the
// dialect impl in the sql package.
type __sqlbundle_sqlite3 struct{}

func (s __sqlbundle_sqlite3) Rebind(sql string) string {
	return sql
}

type __sqlbundle_Literal string

func (__sqlbundle_Literal) private() {}

func (l __sqlbundle_Literal) Render() string { return string(l) }

type __sqlbundle_Literals struct {
	Join string
	SQLs []__sqlbundle_SQL
}

func (__sqlbundle_Literals) private() {}

func (l __sqlbundle_Literals) Render() string {
	var out bytes.Buffer

	first := true
	for _, sql := range l.SQLs {
		if sql == nil {
			continue
		}
		if !first {
			out.WriteString(l.Join)
		}
		first = false
		out.WriteString(sql.Render())
	}

	return out.String()
}

type __sqlbundle_Condition struct {
	// set at compile/embed time
	Name  string
	Left  string
	Equal bool
	Right string

	// set at runtime
	Null bool
}

func (*__sqlbundle_Condition) private() {}

func (c *__sqlbundle_Condition) Render() string {
	// TODO(jeff): maybe check if we can use placeholders instead of the
	// literal null: this would make the templates easier.

	switch {
	case c.Equal && c.Null:
		return c.Left + " is null"
	case c.Equal && !c.Null:
		return c.Left + " = " + c.Right
	case !c.Equal && c.Null:
		return c.Left + " is not null"
	case !c.Equal && !c.Null:
		return c.Left + " != " + c.Right
	default:
		panic("unhandled case")
	}
}

type __sqlbundle_Hole struct {
	// set at compiile/embed time
	Name string

	// set at runtime or possibly embed time
	SQL __sqlbundle_SQL
}

func (*__sqlbundle_Hole) private() {}

func (h *__sqlbundle_Hole) Render() string {
	if h.SQL == nil {
		return ""
	}
	return h.SQL.Render()
}

//
// end runtime support for building sql statements
//

func (impl sqlite3Impl) isConstraintError(err error) (constraint string, ok bool) {
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrConstraint {
			msg := err.Error()
			colon := strings.LastIndex(msg, ":")
			if colon != -1 {
				return strings.TrimSpace(msg[colon:]), true
			}
			return "", true
		}
	}
	return "", false
}

func (obj *sqlite3Impl) deleteAll(ctx context.Context) (count int64, err error) {
	var __res sql.Result
	var __count int64
	__res, err = obj.driver.ExecContext(ctx, "DELETE FROM as;")
	if err != nil {
		return 0, obj.makeErr(err)
	}

	__count, err = __res.RowsAffected()
	if err != nil {
		return 0, obj.makeErr(err)
	}
	count += __count

	return count, nil

}

func (impl pgxImpl) isConstraintError(err error) (constraint string, ok bool) {
	if e, ok := err.(*pgconn.PgError); ok {
		if e.Code[:2] == "23" {
			return e.ConstraintName, true
		}
	}
	return "", false
}

func (obj *pgxImpl) deleteAll(ctx context.Context) (count int64, err error) {
	var __res sql.Result
	var __count int64
	__res, err = obj.driver.ExecContext(ctx, "DELETE FROM as;")
	if err != nil {
		return 0, obj.makeErr(err)
	}

	__count, err = __res.RowsAffected()
	if err != nil {
		return 0, obj.makeErr(err)
	}
	count += __count

	return count, nil

}

func (impl pgxcockroachImpl) isConstraintError(err error) (constraint string, ok bool) {
	if e, ok := err.(*pgconn.PgError); ok {
		if e.Code[:2] == "23" {
			return e.ConstraintName, true
		}
	}
	return "", false
}

func (obj *pgxcockroachImpl) deleteAll(ctx context.Context) (count int64, err error) {
	var __res sql.Result
	var __count int64
	__res, err = obj.driver.ExecContext(ctx, "DELETE FROM as;")
	if err != nil {
		return 0, obj.makeErr(err)
	}

	__count, err = __res.RowsAffected()
	if err != nil {
		return 0, obj.makeErr(err)
	}
	count += __count

	return count, nil

}

func (impl spannerImpl) isConstraintError(err error) (constraint string, ok bool) {
	return "", false
}

func (obj *spannerImpl) deleteAll(ctx context.Context) (count int64, err error) {
	var __res sql.Result
	var __count int64
	__res, err = obj.driver.ExecContext(ctx, "DELETE FROM as;")
	if err != nil {
		return 0, obj.makeErr(err)
	}

	__count, err = __res.RowsAffected()
	if err != nil {
		return 0, obj.makeErr(err)
	}
	count += __count

	return count, nil

}

type Methods interface {
}

type TxMethods interface {
	Methods

	Rebind(s string) string
	Commit() error
	Rollback() error
}

type txMethods interface {
	TxMethods

	deleteAll(ctx context.Context) (int64, error)
	makeErr(err error) error
}

type DBMethods interface {
	Methods

	Schema() []string
	DropSchema() []string

	Rebind(sql string) string
}

type dbMethods interface {
	DBMethods

	wrapTx(tx *sql.Tx) txMethods
	makeErr(err error) error
}

var sqlite3DriverName = func() string {
	var id [16]byte
	_, _ = rand.Read(id[:])
	return fmt.Sprintf("sqlite3_%x", string(id[:]))
}()

func init() {
	sql.Register(sqlite3DriverName, &sqlite3.SQLiteDriver{
		ConnectHook: sqlite3SetupConn,
	})
}

// SQLite3JournalMode controls the journal_mode pragma for all new connections.
// Since it is read without a mutex, it must be changed to the value you want
// before any Open calls.
var SQLite3JournalMode = "WAL"

func sqlite3SetupConn(conn *sqlite3.SQLiteConn) (err error) {
	_, err = conn.Exec("PRAGMA foreign_keys = ON", nil)
	if err != nil {
		return makeErr(err)
	}
	_, err = conn.Exec("PRAGMA journal_mode = "+SQLite3JournalMode, nil)
	if err != nil {
		return makeErr(err)
	}
	return nil
}

func opensqlite3(source string) (*sql.DB, error) {
	return sql.Open(sqlite3DriverName, source)
}

func openpgx(source string) (*sql.DB, error) {
	return sql.Open("pgx", source)
}

func openpgxcockroach(source string) (*sql.DB, error) {
	// try first with "cockroach" as a driver in case someone has registered
	// some special stuff. if that fails, then try again with "pgx" as
	// the driver.
	db, err := sql.Open("cockroach", source)
	if err != nil {
		db, err = sql.Open("pgx", source)
	}
	return db, err
}

func openspanner(source string) (*sql.DB, error) {
	return sql.Open("spanner", strings.TrimPrefix(source, "spanner://"))
}

func spannerConvertJSON(v any) any {
	if v == nil {
		return spanner.NullJSON{Value: nil, Valid: true}
	}
	if v, ok := v.([]byte); ok {
		return spanner.NullJSON{Value: v, Valid: true}
	}
	if v, ok := v.(*[]byte); ok {
		return &spannerJSON{data: v}
	}
	return v
}

type spannerJSON struct {
	data *[]byte
}

func (s *spannerJSON) Scan(input any) error {
	if input == nil {
		*s.data = nil
		return nil
	}
	if v, ok := input.(spanner.NullJSON); ok {
		if !v.Valid || v.Value == nil {
			*s.data = nil
			return nil
		}

		if str, ok := v.Value.(string); ok {
			bytesVal, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return fmt.Errorf("expected base64 from spanner: %w", err)
			}
			*s.data = bytesVal
			return nil
		}

		return fmt.Errorf("unable to decode spanner.NullJSON with type %T", v.Value)
	}
	return fmt.Errorf("unable to decode %T", input)
}
