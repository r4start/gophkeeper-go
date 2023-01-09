create table user_data (
    id bigserial primary key,
    user_id uuid not null,
    resource_id uuid not null,
    data_id oid not null default 0,
    salt bytea not null,
    created  timestamptz not null default now(),
    last_update timestamptz not null default now(),
    is_deleted boolean default false,

    foreign key (user_id)
      references users(id),

    unique (user_id, data_id),
    unique (resource_id)
);
