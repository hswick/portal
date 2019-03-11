CREATE TABLE users(
 id serial PRIMARY KEY,
 name text,
 admin BOOLEAN NOT NULL,
 created_at TIMESTAMP NOT NULL,
 updated_at TIMESTAMP
);
