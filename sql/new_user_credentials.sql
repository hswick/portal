WITH new_user AS(
 INSERT INTO users (name, created_at) VALUES ($1, NOW()) RETURNING id
)
INSERT INTO credentials (user_id, password, created_at) VALUES (
 (SELECT id FROM new_user),
 $2,
 NOW()
) RETURNING user_id;
