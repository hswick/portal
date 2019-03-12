SELECT password FROM credentials WHERE user_id = $1 LIMIT 1;
