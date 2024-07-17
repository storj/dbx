-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE bars (
	id bigserial NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( id ),
	UNIQUE ( name )
) ;
CREATE TABLE bazs (
	id bigserial NOT NULL,
	name text NOT NULL,
	PRIMARY KEY ( id ),
	UNIQUE ( name )
) ;
CREATE TABLE foos (
	id bigserial NOT NULL,
	bar_id bigint NOT NULL REFERENCES bars( id ),
	baz_id bigint NOT NULL REFERENCES bazs( id ),
	name text NOT NULL,
	PRIMARY KEY ( id ),
	UNIQUE ( name ),
	UNIQUE ( bar_id ),
	UNIQUE ( baz_id )
)
