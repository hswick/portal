WITH other_user AS (
 SELECT id FROM users WHERE name = $1
)
UPDATE users SET admin = $2 WHERE id = (SELECT id FROM other_user);
