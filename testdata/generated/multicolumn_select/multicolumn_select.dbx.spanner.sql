-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE foos_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE foos (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE foos_pk)),
	one INT64 NOT NULL,
	two INT64 NOT NULL
) PRIMARY KEY ( pk )
