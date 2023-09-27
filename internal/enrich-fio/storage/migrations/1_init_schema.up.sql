CREATE TABLE IF NOT EXISTS person (
    id uuid NOT NULL,
    name varchar(50) NOT NULL,
    surname varchar(50) NOT NULL,
    patronymic varchar(50),
    gender varchar(10),
    nationality varchar(10),
    age smallint
);

