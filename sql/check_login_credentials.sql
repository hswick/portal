SELECT id, name FROM users INNER JOIN credentials ON users.id = credentials.user_id WHERE users.name = $1 AND credentials.password = $2 LIMIT 1;
