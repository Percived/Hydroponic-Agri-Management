-- v2.4 down: revert batch stage config refs + runtime + active climate profile

DROP TABLE IF EXISTS `batch_stage_runtime`;

ALTER TABLE `crop_batches`
    DROP INDEX `idx_crop_batches_climate_profile`,
    DROP COLUMN `active_climate_profile_id`;

ALTER TABLE `batch_stage_plans`
    DROP COLUMN `climate_profile_id`,
    DROP COLUMN `policy_id`,
    DROP COLUMN `recipe_id`;

