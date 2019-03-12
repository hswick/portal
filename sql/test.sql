\i sql/create_users.sql
\i sql/create_credentials.sql

WITH new_user AS (
 INSERT INTO users (name, admin, created_at) VALUES ('shiba', TRUE, NOW()) RETURNING id
)
INSERT INTO credentials (user_id, password, created_at) VALUES (
 (SELECT id FROM new_user),
 'foobar',
 NOW()
);
 

