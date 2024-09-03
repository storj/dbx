-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE a (
	id INTEGER NOT NULL,
	PRIMARY KEY ( id )
) ;
CREATE TABLE d (
	id INTEGER NOT NULL,
	a INTEGER NOT NULL,
	b INTEGER NOT NULL,
	c INTEGER NOT NULL,
	PRIMARY KEY ( id ),
	UNIQUE ( a ),
	UNIQUE ( b, c )
) ;
CREATE TABLE b (
	id INTEGER NOT NULL,
	a_id INTEGER NOT NULL REFERENCES a( id ),
	PRIMARY KEY ( id )
) ;
CREATE TABLE c (
	id INTEGER NOT NULL,
	lat REAL NOT NULL,
	lon REAL NOT NULL,
	b_id INTEGER NOT NULL REFERENCES b( id ),
	PRIMARY KEY ( id ),
	UNIQUE ( b_id )
)
