-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE people_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE people (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE people_pk)),
	a STRING(MAX) NOT NULL,
	b INT64 NOT NULL,
	c INT64,
	d INT64 NOT NULL,
	e INT64
) PRIMARY KEY ( pk )
