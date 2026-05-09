-- v2.1: Batch lifecycle management — device binding, planting records, batch_id references
-- Part of Phase 1: Schema fix + device binding + state machine

-- 1. batch_devices: device-to-batch binding table
CREATE TABLE IF NOT EXISTS `batch_devices` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `batch_id` BIGINT UNSIGNED NOT NULL,
    `device_type` ENUM('sensor','actuator') NOT NULL,
    `device_id` BIGINT UNSIGNED NOT NULL COMMENT 'sensor_device_id or actuator_device_id',
    `bound_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `unbound_at` DATETIME DEFAULT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_batch_device` (`batch_id`, `device_type`, `device_id`),
    INDEX `idx_batch_id` (`batch_id`),
    INDEX `idx_device` (`device_type`, `device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. planting_records: planting/transplanting details per batch (1:1 with crop_batches)
CREATE TABLE IF NOT EXISTS `planting_records` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `batch_id` BIGINT UNSIGNED NOT NULL,
    `seed_source` VARCHAR(128) DEFAULT NULL COMMENT '种子来源',
    `seed_batch_no` VARCHAR(64) DEFAULT NULL COMMENT '种子批号',
    `seedling_age_days` INT UNSIGNED DEFAULT NULL COMMENT '苗龄（天）',
    `seeded_at` DATETIME DEFAULT NULL COMMENT '播种时间',
    `planted_at` DATETIME DEFAULT NULL COMMENT '定植时间',
    `actual_plant_count` INT UNSIGNED DEFAULT NULL COMMENT '实际定植株数',
    `initial_ec` DECIMAL(12,4) DEFAULT NULL COMMENT '定植时EC值',
    `initial_ph` DECIMAL(12,4) DEFAULT NULL COMMENT '定植时pH值',
    `initial_water_temp` DECIMAL(12,4) DEFAULT NULL COMMENT '定植时水温',
    `initial_nutrient_recipe_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '定植时使用的营养液配方ID',
    `planted_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '定植操作人',
    `note` VARCHAR(255) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_batch_id` (`batch_id`),
    INDEX `idx_batch_id` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. alerts: add batch_id column for batch-scoped alert filtering
ALTER TABLE `alerts`
    ADD COLUMN `batch_id` BIGINT UNSIGNED DEFAULT NULL AFTER `actuator_channel_id`,
    ADD INDEX `idx_alerts_batch_id` (`batch_id`);

-- 4. control_commands: add batch_id column for batch-scoped command tracing
ALTER TABLE `control_commands`
    ADD COLUMN `batch_id` BIGINT UNSIGNED DEFAULT NULL AFTER `actuator_channel_id`,
    ADD INDEX `idx_commands_batch_id` (`batch_id`);
