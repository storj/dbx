-- AUTOGENERATED BY storj.io/dbx
-- DO NOT EDIT
CREATE TABLE people (
	pk bigserial NOT NULL,
	name text NOT NULL,
	value jsonb NOT NULL,
	value_up jsonb NOT NULL,
	value_null jsonb,
	value_null_up jsonb,
	value_default jsonb NOT NULL DEFAULT '{}',
	PRIMARY KEY ( pk )
)