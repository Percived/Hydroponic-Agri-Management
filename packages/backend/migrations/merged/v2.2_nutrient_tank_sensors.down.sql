-- v2.2 down: Remove sensor channel binding columns from nutrient_tanks
ALTER TABLE nutrient_tanks
  DROP FOREIGN KEY fk_nutrient_tanks_temp,
  DROP FOREIGN KEY fk_nutrient_tanks_level,
  DROP FOREIGN KEY fk_nutrient_tanks_ph,
  DROP FOREIGN KEY fk_nutrient_tanks_ec,
  DROP COLUMN temp_sensor_channel_id,
  DROP COLUMN level_sensor_channel_id,
  DROP COLUMN ph_sensor_channel_id,
  DROP COLUMN ec_sensor_channel_id;
