-- v2.4: Batch stage config refs + runtime + active climate profile

ALTER TABLE `batch_stage_plans`
    ADD COLUMN `recipe_id` BIGINT UNSIGNED DEFAULT NULL AFTER `growth_stage_id`,
    ADD COLUMN `policy_id` BIGINT UNSIGNED DEFAULT NULL AFTER `recipe_id`,
    ADD COLUMN `climate_profile_id` BIGINT UNSIGNED DEFAULT NULL AFTER `policy_id`;

ALTER TABLE `crop_batches`
    ADD COLUMN `active_climate_profile_id` BIGINT UNSIGNED DEFAULT NULL AFTER `active_policy_id`,
    ADD INDEX `idx_crop_batches_climate_profile` (`active_climate_profile_id`);

CREATE TABLE IF NOT EXISTS `batch_stage_runtime` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `current_stage_plan_id` BIGINT UNSIGNED DEFAULT NULL,
  `current_growth_stage_id` BIGINT UNSIGNED DEFAULT NULL,
  `last_switched_at` DATETIME(3) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_stage_runtime_batch_id` (`batch_id`),
  KEY `idx_batch_stage_runtime_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

