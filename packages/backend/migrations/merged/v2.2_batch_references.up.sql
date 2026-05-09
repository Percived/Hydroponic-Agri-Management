-- v2.2: Batch references — add FK columns for recipe and policy to crop_batches
-- Part of Phase 2: Planting flow + stage tracking

ALTER TABLE `crop_batches`
    ADD COLUMN `active_recipe_id` BIGINT UNSIGNED DEFAULT NULL AFTER `policy_version`,
    ADD COLUMN `active_policy_id` BIGINT UNSIGNED DEFAULT NULL AFTER `active_recipe_id`,
    ADD INDEX `idx_crop_batches_recipe` (`active_recipe_id`),
    ADD INDEX `idx_crop_batches_policy` (`active_policy_id`);
