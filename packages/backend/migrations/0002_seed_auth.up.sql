INSERT INTO `roles` (`name`, `description`)
VALUES
  ('ADMIN', 'System administrator'),
  ('OPERATOR', 'Greenhouse operator'),
  ('VIEWER', 'Read-only user')
ON DUPLICATE KEY UPDATE
  `description` = VALUES(`description`);

INSERT INTO `users` (`username`, `password_hash`, `nickname`, `status`)
VALUES
  ('admin', '$2a$10$2TMlR.5ZCg08VGvAu2uDCO4W4EHwjvtoGUO.XZc..DQQHI/.8R3HW', 'Administrator', 'ENABLED')
ON DUPLICATE KEY UPDATE
  `password_hash` = VALUES(`password_hash`),
  `nickname` = VALUES(`nickname`),
  `status` = VALUES(`status`);

INSERT INTO `user_roles` (`user_id`, `role_id`)
SELECT u.id, r.id
FROM `users` u
JOIN `roles` r ON r.name = 'ADMIN'
WHERE u.username = 'admin'
ON DUPLICATE KEY UPDATE
  `user_id` = VALUES(`user_id`),
  `role_id` = VALUES(`role_id`);
