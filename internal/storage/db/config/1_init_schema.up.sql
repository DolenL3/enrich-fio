CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS person (
    id uuid DEFAULT (uuid_generate_v4()),
    name varchar(50) NOT NULL,
    surname varchar(50) NOT NULL,
    pantronymic varchar(50),
    gender varchar(10),
    nationality varchar(100),
    age smallint
);

