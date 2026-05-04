INSERT INTO `metrics` (`code`, `name`, `unit`, `min_value`, `max_value`)
VALUES
  ('TEMP', '温度', 'C', -20, 60),
  ('HUMIDITY', '湿度', '%', 0, 100),
  ('PH', '酸碱度', 'pH', 0, 14),
  ('EC', '电导率', 'mS/cm', 0, 10),
  ('CO2', '二氧化碳', 'ppm', 0, 2000),
  ('LIGHT', '光照', 'lx', 0, 200000)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `unit` = VALUES(`unit`),
  `min_value` = VALUES(`min_value`),
  `max_value` = VALUES(`max_value`);
