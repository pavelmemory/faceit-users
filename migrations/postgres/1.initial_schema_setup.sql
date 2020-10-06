-- TODO: this should be part of the database automatic migration flow

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id         UUID DEFAULT UUID_GENERATE_V1() PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name  VARCHAR(50) NOT NULL,
    nickname   VARCHAR(30) NOT NULL UNIQUE,
    email      VARCHAR(30) NOT NULL UNIQUE,
    password   VARCHAR(34) NOT NULL,
    country    CHAR(2),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
