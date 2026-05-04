DELETE ur
FROM `user_roles` ur
JOIN `users` u ON u.id = ur.user_id
JOIN `roles` r ON r.id = ur.role_id
WHERE u.username = 'admin' AND r.name = 'ADMIN';

DELETE FROM `users`
WHERE `username` = 'admin';

DELETE FROM `roles`
WHERE `name` IN ('ADMIN', 'OPERATOR', 'VIEWER');
