WITH other_user AS (
 SELECT id FROM users WHERE name = $1
)
UPDATE credentials SET password = $2 WHERE user_id = (SELECT id FROM other_user); 
