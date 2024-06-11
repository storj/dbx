-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE a (
	id bigserial NOT NULL,
	PRIMARY KEY ( id )
) ;
CREATE TABLE b (
	id bigserial NOT NULL,
	a_id bigint NOT NULL REFERENCES a( id ),
	PRIMARY KEY ( id )
) ;
CREATE TABLE c (
	id bigserial NOT NULL,
	lat real NOT NULL,
	lon real NOT NULL,
	b_id bigint NOT NULL REFERENCES b( id ),
	PRIMARY KEY ( id ),
	UNIQUE ( b_id )
)
