CREATE TABLE IF NOT EXISTS users
(
    id serial primary key,
    name text not null,
    email text not null unique,
    pass_hash text not null
);

CREATE INDEX IF NOT EXISTS idx_email on users (email);

CREATE TABLE IF NOT EXISTS apps
(
    id integer primary key,
    name text not null unique,
    secret text not null unique
);