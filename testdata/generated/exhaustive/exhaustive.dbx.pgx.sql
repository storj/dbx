-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE as (
	pk bigserial NOT NULL,
	ctime timestamp with time zone NOT NULL,
	mtime timestamp with time zone NOT NULL,
	id text NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id ),
	UNIQUE ( name )
) ;
CREATE TABLE bs (
	pk bigserial NOT NULL,
	id text NOT NULL,
	data text NOT NULL,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
) ;
CREATE TABLE foos (
	id bigserial NOT NULL,
	int integer NOT NULL,
	int64 bigint NOT NULL,
	uint integer NOT NULL,
	uint64 bigint NOT NULL,
	float real NOT NULL,
	float64 double precision NOT NULL,
	string text NOT NULL,
	blob bytea NOT NULL,
	timestamp timestamp with time zone NOT NULL,
	utimestamp timestamp NOT NULL,
	bool boolean NOT NULL,
	date date NOT NULL,
	json jsonb NOT NULL,
	null_int integer,
	null_int64 bigint,
	null_uint integer,
	null_uint64 bigint,
	null_float real,
	null_float64 double precision,
	null_string text,
	null_blob bytea,
	null_timestamp timestamp with time zone,
	null_utimestamp timestamp,
	null_bool boolean,
	null_date date,
	null_json jsonb,
	PRIMARY KEY ( id )
) ;
CREATE TABLE a_bs (
	b_pk bigint NOT NULL REFERENCES bs( pk ) ON DELETE CASCADE,
	a_pk bigint NOT NULL REFERENCES as( pk ) ON DELETE CASCADE,
	PRIMARY KEY ( b_pk, a_pk )
) ;
CREATE TABLE cs (
	pk bigserial NOT NULL,
	id text NOT NULL,
	b_pk bigint NOT NULL REFERENCES bs( pk ) ON DELETE CASCADE,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
) ;
CREATE TABLE es (
	pk bigserial NOT NULL,
	id text NOT NULL,
	a_id text NOT NULL REFERENCES as( id ) ON DELETE CASCADE,
	PRIMARY KEY ( pk ),
	UNIQUE ( id )
) ;
CREATE TABLE ds (
	pk bigserial NOT NULL,
	id text NOT NULL,
	alias text,
	date timestamp with time zone NOT NULL,
	e_id text NOT NULL REFERENCES es( id ),
	a_id text NOT NULL REFERENCES as( id ) ON DELETE CASCADE,
	PRIMARY KEY ( pk ),
	UNIQUE ( id ),
	UNIQUE ( a_id, alias )
) ;
CREATE UNIQUE INDEX as_ctime_mtime_unique_index ON as ( ctime, mtime )