-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE bars_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE bars (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE bars_pk)),
	data STRING(MAX) NOT NULL
) PRIMARY KEY ( pk ) ;
CREATE TABLE foos (
	pk BYTES(MAX) NOT NULL,
	data STRING(MAX) NOT NULL
) PRIMARY KEY ( pk )