-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE people_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE people (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE people_pk)),
	name STRING(MAX) NOT NULL,
	value INT64 NOT NULL,
	value_up INT64 NOT NULL,
	value_null INT64,
	value_null_up INT64
) PRIMARY KEY ( pk )
