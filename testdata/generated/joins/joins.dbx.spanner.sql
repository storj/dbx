-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE SEQUENCE users_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE users (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE users_pk))
) PRIMARY KEY ( pk ) ;
CREATE SEQUENCE associated_accounts_pk OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE associated_accounts (
	pk INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE associated_accounts_pk)),
	user_pk INT64 NOT NULL,
	CONSTRAINT associated_accounts_user_pk_fkey FOREIGN KEY (user_pk) REFERENCES users (pk)
) PRIMARY KEY ( pk ) ;
CREATE SEQUENCE sessions_id OPTIONS (sequence_kind='bit_reversed_positive') ;
CREATE TABLE sessions (
	id INT64 NOT NULL DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE sessions_id)),
	user_pk INT64 NOT NULL,
	CONSTRAINT sessions_user_pk_fkey FOREIGN KEY (user_pk) REFERENCES users (pk)
) PRIMARY KEY ( id )
