-- v2.2 down: remove reference columns from crop_batches

ALTER TABLE `crop_batches`
    DROP INDEX `idx_crop_batches_policy`,
    DROP INDEX `idx_crop_batches_recipe`,
    DROP COLUMN `active_policy_id`,
    DROP COLUMN `active_recipe_id`;
