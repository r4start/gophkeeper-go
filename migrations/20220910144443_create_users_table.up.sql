create table users (
    id uuid primary key,
    login varchar(2048) not null unique,
    key_salt bytea not null,
    salt bytea not null,
    secret bytea not null,
    created timestamptz not null default now(),
    last_update timestamptz not null default now(),
    is_deleted boolean default false
);
