CREATE USER admin
WITH PASSWORD 'admin123'
    CREATEDB;

CREATE DATABASE db
    WITH OWNER
admin;

\connect db;

CREATE EXTENSION pgcrypto;

CREATE SCHEMA board AUTHORIZATION admin;

GRANT ALL PRIVILEGES
    ON ALL TABLES
    IN SCHEMA board
    TO admin;


-- \connect db;

-- CREATE EXTENSION pgcrypto;

-- CREATE TABLE board.user
-- (
--     Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     Username varchar(250) UNIQUE,
--     Emailaddress varchar(250) UNIQUE,
--     UserPassword bytea,
--     UserRole int
-- );

-- CREATE TABLE board.thread
-- (
--     Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     UserId UUID,
--     Title varchar(250),
--     PostedAt TIMESTAMP DEFAULT now(),
--     Deleted boolean DEFAULT false
-- );

-- CREATE INDEX thread_deleted_idx ON board.thread (Deleted);

-- CREATE TABLE board.thread_post
-- (
--     Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     ThreadId UUID REFERENCES board.thread (Id),
--     UserId UUID,
--     Body text,
--     PostedAt TIMESTAMP DEFAULT now(),
--     Deleted boolean DEFAULT false
-- );

-- CREATE INDEX thread_post_deleted_idx ON board.thread_post (Deleted);

