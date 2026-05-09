-- v2.2: Add sensor channel binding columns to nutrient_tanks
ALTER TABLE nutrient_tanks
  ADD COLUMN ec_sensor_channel_id BIGINT UNSIGNED NULL AFTER status,
  ADD COLUMN ph_sensor_channel_id BIGINT UNSIGNED NULL AFTER ec_sensor_channel_id,
  ADD COLUMN level_sensor_channel_id BIGINT UNSIGNED NULL AFTER ph_sensor_channel_id,
  ADD COLUMN temp_sensor_channel_id BIGINT UNSIGNED NULL AFTER level_sensor_channel_id,
  ADD CONSTRAINT fk_nutrient_tanks_ec FOREIGN KEY (ec_sensor_channel_id) REFERENCES sensor_channels(id) ON DELETE SET NULL,
  ADD CONSTRAINT fk_nutrient_tanks_ph FOREIGN KEY (ph_sensor_channel_id) REFERENCES sensor_channels(id) ON DELETE SET NULL,
  ADD CONSTRAINT fk_nutrient_tanks_level FOREIGN KEY (level_sensor_channel_id) REFERENCES sensor_channels(id) ON DELETE SET NULL,
  ADD CONSTRAINT fk_nutrient_tanks_temp FOREIGN KEY (temp_sensor_channel_id) REFERENCES sensor_channels(id) ON DELETE SET NULL;
