CREATE TABLE credentials(
 user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
 password text,
 created_at TIMESTAMP NOT NULL,
 updated_at TIMESTAMP
);
