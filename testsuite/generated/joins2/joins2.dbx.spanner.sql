-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE bars_id OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE bars (
	id INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE bars_id)),
	name STRING(MAX) NOT NULL
) PRIMARY KEY ( id ) ;
CREATE UNIQUE INDEX index_bars_name ON bars ( name ) ;
CREATE SEQUENCE bazs_id OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE bazs (
	id INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE bazs_id)),
	name STRING(MAX) NOT NULL
) PRIMARY KEY ( id ) ;
CREATE UNIQUE INDEX index_bazs_name ON bazs ( name ) ;
CREATE SEQUENCE foos_id OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE foos (
	id INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE foos_id)),
	bar_id INT64 NOT NULL,
	baz_id INT64 NOT NULL,
	name STRING(MAX) NOT NULL,
	CONSTRAINT foos_bar_id_fkey FOREIGN KEY (bar_id) REFERENCES bars (id),
	CONSTRAINT foos_baz_id_fkey FOREIGN KEY (baz_id) REFERENCES bazs (id)
) PRIMARY KEY ( id ) ;
CREATE UNIQUE INDEX index_foos_name ON foos ( name ) ;
CREATE UNIQUE INDEX index_foos_bar_id ON foos ( bar_id ) ;
CREATE UNIQUE INDEX index_foos_baz_id ON foos ( baz_id )