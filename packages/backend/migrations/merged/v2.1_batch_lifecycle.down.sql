-- v2.1 down: revert batch lifecycle changes

ALTER TABLE `control_commands`
    DROP INDEX `idx_commands_batch_id`,
    DROP COLUMN `batch_id`;

ALTER TABLE `alerts`
    DROP INDEX `idx_alerts_batch_id`,
    DROP COLUMN `batch_id`;

DROP TABLE IF EXISTS `planting_records`;
DROP TABLE IF EXISTS `batch_devices`;
