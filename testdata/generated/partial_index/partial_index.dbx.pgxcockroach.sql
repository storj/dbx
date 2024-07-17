-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE as (
	pk bigserial NOT NULL,
	ctime timestamp with time zone NOT NULL,
	mtime timestamp with time zone NOT NULL,
	id text NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( pk )
) ;
CREATE INDEX as_ctime_mtime_index ON as ( ctime, mtime ) STORING ( id, name ) WHERE as.id > 'foo' AND as.name > 'bar'
