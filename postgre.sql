create schema message
    create table message (
        id         bigserial primary key,
        uuid_v4    character varying,
        app_id     character varying,
        topic      character varying,
        data       jsonb,
        created_at timestamp without time zone default current_timestamp,
        auto_emit  boolean default true
        )
    create table publish (
        id         bigserial primary key,
        msg_id     bigint,
        publish_at timestamp without time zone default current_timestamp
        );