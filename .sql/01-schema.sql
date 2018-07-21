CREATE USER admin
WITH PASSWORD 'admin123'
    CREATEDB;

CREATE DATABASE db
    WITH OWNER
admin;

\connect db;

CREATE EXTENSION pgcrypto;

CREATE SCHEMA board AUTHORIZATION admin;

CREATE TABLE board.user
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Username varchar(250) UNIQUE,
    Emailaddress varchar(250) UNIQUE,
    UserPassword bytea,
    IsAdmin boolean DEFAULT FALSE
);

CREATE TABLE board.thread
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    UserId UUID,
    Title varchar(250),
    PostedAt TIMESTAMP DEFAULT now(),
    LastPost TIMESTAMP DEFAULT now()
);

CREATE INDEX thread_id_idx ON board.thread (Id);
CREATE INDEX thread_posted_at_idx ON board.thread (PostedAt);

CREATE TABLE board.thread_post
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ThreadId UUID REFERENCES board.thread (Id),
    UserId UUID,
    Body text,
    PostedAt TIMESTAMP DEFAULT now()
);

GRANT ALL PRIVILEGES
    ON ALL TABLES
    IN SCHEMA board
    TO admin;
