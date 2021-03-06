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
