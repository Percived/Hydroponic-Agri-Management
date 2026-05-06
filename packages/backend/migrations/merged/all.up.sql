-- ============================================================
-- Hydroponic Agriculture Management System
-- Complete Domain Model Refactor (DWC-focused)
-- Version: v2.0.0
-- ============================================================

-- ============================================================
-- 域 1：组织与设施
-- ============================================================

CREATE TABLE `greenhouses` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `location` VARCHAR(128) DEFAULT NULL,
  `area_sqm` DECIMAL(10,2) DEFAULT NULL COMMENT '面积（平方米）',
  `description` VARCHAR(255) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_greenhouses_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `growing_zones` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `system_type` VARCHAR(16) NOT NULL DEFAULT 'DWC' COMMENT 'DWC/NFT/EBB_FLOW/DRIP',
  `tank_volume_liter` DECIMAL(10,2) DEFAULT NULL COMMENT '营养液槽容积（升）',
  `planting_density_per_sqm` DECIMAL(8,2) DEFAULT NULL COMMENT '定植密度（株/m²）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_growing_zones_code` (`greenhouse_id`, `code`),
  CONSTRAINT `fk_zones_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 2：设备（传感器 + 执行器分表）
-- ============================================================

CREATE TABLE `sensor_devices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '所属种植区，环境传感器可为NULL',
  `device_code` VARCHAR(64) NOT NULL COMMENT 'MQTT client_id，如 GH1-ENV-01',
  `name` VARCHAR(64) NOT NULL COMMENT '如 "1号温室环境节点"',
  `model` VARCHAR(64) DEFAULT NULL COMMENT '硬件型号，如 ESP32-DEVKIT',
  `firmware_version` VARCHAR(64) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ONLINE' COMMENT 'ONLINE/OFFLINE/FAULT',
  `last_seen_at` DATETIME(3) DEFAULT NULL,
  `protocol` VARCHAR(16) NOT NULL DEFAULT 'MQTT',
  `metadata` JSON DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sensor_devices_code` (`device_code`),
  KEY `idx_sensors_greenhouse_status` (`greenhouse_id`, `status`),
  KEY `idx_sensors_zone` (`growing_zone_id`),
  CONSTRAINT `fk_sensors_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_sensors_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sensor_channels` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sensor_device_id` BIGINT UNSIGNED NOT NULL,
  `channel_code` VARCHAR(64) NOT NULL COMMENT '该设备内唯一，如 TEMP/HUMIDITY/CO2/LIGHT',
  `metric_code` VARCHAR(32) NOT NULL COMMENT '指标编码，关联 metric_definitions.code',
  `unit` VARCHAR(16) NOT NULL,
  `precision_digits` TINYINT UNSIGNED NOT NULL DEFAULT 2,
  `range_min` DECIMAL(12,4) DEFAULT NULL COMMENT '探头量程下限',
  `range_max` DECIMAL(12,4) DEFAULT NULL COMMENT '探头量程上限',
  `sampling_interval_sec` INT UNSIGNED NOT NULL DEFAULT 60 COMMENT '该通道的采样间隔',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `last_reported_at` DATETIME(3) DEFAULT NULL,
  `metadata` JSON DEFAULT NULL COMMENT '探头校准参数等',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sensor_channels_device_code` (`sensor_device_id`, `channel_code`),
  KEY `idx_sc_metric_enabled` (`metric_code`, `enabled`),
  CONSTRAINT `fk_sc_device` FOREIGN KEY (`sensor_device_id`) REFERENCES `sensor_devices` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `actuator_devices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL,
  `device_code` VARCHAR(64) NOT NULL COMMENT 'MQTT client_id，如 GH1-RELAY-01',
  `name` VARCHAR(64) NOT NULL COMMENT '如 "1号温室继电器模块"',
  `model` VARCHAR(64) DEFAULT NULL COMMENT '硬件型号',
  `firmware_version` VARCHAR(64) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ONLINE',
  `last_seen_at` DATETIME(3) DEFAULT NULL,
  `protocol` VARCHAR(16) NOT NULL DEFAULT 'MQTT',
  `metadata` JSON DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_actuator_devices_code` (`device_code`),
  KEY `idx_actuators_greenhouse_status` (`greenhouse_id`, `status`),
  KEY `idx_actuators_zone` (`growing_zone_id`),
  CONSTRAINT `fk_actuators_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_actuators_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `actuator_channels` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `actuator_device_id` BIGINT UNSIGNED NOT NULL,
  `channel_code` VARCHAR(64) NOT NULL COMMENT '该设备内唯一，如 CH1/CH2 或 FAN/PUMP',
  `actuator_type` VARCHAR(16) NOT NULL COMMENT 'PUMP/AERATOR/FAN/VALVE/SHADE/LED/HEATER/CO2_GEN/FOGGER',
  `current_state` VARCHAR(16) NOT NULL DEFAULT 'OFF' COMMENT 'ON/OFF/PERCENTAGE',
  `rated_power_watt` DECIMAL(10,2) DEFAULT NULL COMMENT '额定功率（瓦）',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `metadata` JSON DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_actuator_channels_device_code` (`actuator_device_id`, `channel_code`),
  KEY `idx_ac_type_enabled` (`actuator_type`, `enabled`),
  CONSTRAINT `fk_ac_device` FOREIGN KEY (`actuator_device_id`) REFERENCES `actuator_devices` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 3：统一指标定义
-- ============================================================

CREATE TABLE `metric_definitions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `unit` VARCHAR(16) NOT NULL,
  `precision_digits` TINYINT UNSIGNED NOT NULL DEFAULT 2,
  `normal_range_min` DECIMAL(12,4) DEFAULT NULL COMMENT '植物正常范围下限',
  `normal_range_max` DECIMAL(12,4) DEFAULT NULL COMMENT '植物正常范围上限',
  `is_core` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '核心指标（温度/pH/EC等）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_metric_definitions_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 4：遥测数据
-- ============================================================

CREATE TABLE `telemetry_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sensor_channel_id` BIGINT UNSIGNED NOT NULL,
  `metric_code` VARCHAR(32) NOT NULL,
  `value` DECIMAL(12,4) NOT NULL,
  `raw_value` DECIMAL(12,4) DEFAULT NULL,
  `quality_flag` VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT 'normal/outlier/missing/interpolated',
  `collected_at` DATETIME(3) NOT NULL,
  `ingested_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `batch_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联种植批次',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_telemetry_ch_metric_time` (`sensor_channel_id`, `metric_code`, `collected_at`),
  KEY `idx_telemetry_metric_time` (`metric_code`, `collected_at`),
  KEY `idx_telemetry_batch_time` (`batch_id`, `collected_at`),
  CONSTRAINT `fk_telemetry_channel` FOREIGN KEY (`sensor_channel_id`) REFERENCES `sensor_channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 5：营养液管理
-- ============================================================

CREATE TABLE `nutrient_tanks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `growing_zone_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) NOT NULL,
  `total_volume_liter` DECIMAL(10,2) NOT NULL COMMENT '总容积（升）',
  `current_volume_liter` DECIMAL(10,2) DEFAULT NULL COMMENT '当前液量（估算）',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_nutrient_tanks_code` (`growing_zone_id`, `code`),
  CONSTRAINT `fk_tanks_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `solution_change_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tank_id` BIGINT UNSIGNED NOT NULL,
  `change_type` VARCHAR(16) NOT NULL COMMENT 'FULL_REPLACE/PARTIAL_REFRESH/TOP_UP',
  `volume_replaced_liter` DECIMAL(10,2) NOT NULL COMMENT '更换/补水量（升）',
  `source_water_ec` DECIMAL(12,4) DEFAULT NULL COMMENT '源水 EC',
  `source_water_ph` DECIMAL(12,4) DEFAULT NULL COMMENT '源水 pH',
  `before_ec` DECIMAL(12,4) DEFAULT NULL,
  `before_ph` DECIMAL(12,4) DEFAULT NULL,
  `after_ec` DECIMAL(12,4) DEFAULT NULL,
  `after_ph` DECIMAL(12,4) DEFAULT NULL,
  `nutrient_a_added_ml` DECIMAL(10,2) DEFAULT NULL COMMENT 'A 浓缩液添加量（ml）',
  `nutrient_b_added_ml` DECIMAL(10,2) DEFAULT NULL COMMENT 'B 浓缩液添加量（ml）',
  `acid_added_ml` DECIMAL(10,2) DEFAULT NULL COMMENT 'pH 调节酸添加量',
  `alkali_added_ml` DECIMAL(10,2) DEFAULT NULL COMMENT 'pH 调节碱添加量',
  `note` VARCHAR(255) DEFAULT NULL,
  `operated_by` BIGINT UNSIGNED DEFAULT NULL,
  `operated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_solution_change_tank_time` (`tank_id`, `operated_at`),
  CONSTRAINT `fk_solution_change_tank` FOREIGN KEY (`tank_id`) REFERENCES `nutrient_tanks` (`id`),
  CONSTRAINT `fk_solution_change_user` FOREIGN KEY (`operated_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `ion_test_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tank_id` BIGINT UNSIGNED NOT NULL,
  `batch_id` BIGINT UNSIGNED DEFAULT NULL,
  `sample_code` VARCHAR(64) NOT NULL,
  `sampled_at` DATETIME(3) NOT NULL COMMENT '取样时间',
  `tested_at` DATETIME(3) DEFAULT NULL COMMENT '检测完成时间',
  `test_method` VARCHAR(16) NOT NULL DEFAULT 'LAB' COMMENT 'LAB/STRIP/METER',
  `no3_n` DECIMAL(10,2) DEFAULT NULL COMMENT '硝态氮 NO3-N (mg/L)',
  `nh4_n` DECIMAL(10,2) DEFAULT NULL COMMENT '铵态氮 NH4-N (mg/L)',
  `p` DECIMAL(10,2) DEFAULT NULL COMMENT '磷 P (mg/L)',
  `k` DECIMAL(10,2) DEFAULT NULL COMMENT '钾 K (mg/L)',
  `ca` DECIMAL(10,2) DEFAULT NULL COMMENT '钙 Ca (mg/L)',
  `mg` DECIMAL(10,2) DEFAULT NULL COMMENT '镁 Mg (mg/L)',
  `s` DECIMAL(10,2) DEFAULT NULL COMMENT '硫 S (mg/L)',
  `fe` DECIMAL(10,4) DEFAULT NULL COMMENT '铁 Fe (mg/L)',
  `mn` DECIMAL(10,4) DEFAULT NULL COMMENT '锰 Mn (mg/L)',
  `zn` DECIMAL(10,4) DEFAULT NULL COMMENT '锌 Zn (mg/L)',
  `b` DECIMAL(10,4) DEFAULT NULL COMMENT '硼 B (mg/L)',
  `cu` DECIMAL(10,4) DEFAULT NULL COMMENT '铜 Cu (mg/L)',
  `mo` DECIMAL(10,4) DEFAULT NULL COMMENT '钼 Mo (mg/L)',
  `ec_at_sample` DECIMAL(12,4) DEFAULT NULL COMMENT '取样时实测 EC',
  `ph_at_sample` DECIMAL(12,4) DEFAULT NULL COMMENT '取样时实测 pH',
  `lab_name` VARCHAR(64) DEFAULT NULL COMMENT '检测机构',
  `report_url` VARCHAR(255) DEFAULT NULL COMMENT '检测报告链接',
  `note` VARCHAR(255) DEFAULT NULL,
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ion_test_sample_code` (`sample_code`),
  KEY `idx_ion_test_tank_time` (`tank_id`, `sampled_at`),
  KEY `idx_ion_test_batch_time` (`batch_id`, `sampled_at`),
  CONSTRAINT `fk_ion_test_tank` FOREIGN KEY (`tank_id`) REFERENCES `nutrient_tanks` (`id`),
  CONSTRAINT `fk_ion_test_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_ion_test_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `nutrient_concentrate_inventory` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `concentrate_type` VARCHAR(8) NOT NULL COMMENT 'A/B/ACID/ALKALI',
  `brand` VARCHAR(64) DEFAULT NULL,
  `product_name` VARCHAR(128) DEFAULT NULL,
  `total_volume_ml` DECIMAL(12,2) NOT NULL COMMENT '购入总量（ml）',
  `remaining_volume_ml` DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '剩余量（ml）',
  `unit_price` DECIMAL(10,2) DEFAULT NULL,
  `batch_no` VARCHAR(64) DEFAULT NULL COMMENT '厂家批次号',
  `expired_at` DATE DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'IN_USE' COMMENT 'IN_USE/EMPTY/EXPIRED',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_concentrate_greenhouse` (`greenhouse_id`, `status`),
  CONSTRAINT `fk_concentrate_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 6：作物管理
-- ============================================================

CREATE TABLE `crop_varieties` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `default_cycle_days` INT UNSIGNED DEFAULT NULL COMMENT '默认全周期天数',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_crop_varieties_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `growth_stages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `sort_order` INT UNSIGNED NOT NULL DEFAULT 0,
  `default_duration_days` INT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_growth_stages_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `crop_batches` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_no` VARCHAR(64) NOT NULL,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL,
  `crop_variety_id` BIGINT UNSIGNED NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'PLANNED' COMMENT 'PLANNED/RUNNING/HARVESTING/COMPLETED/ABORTED',
  `planting_density` DECIMAL(8,2) DEFAULT NULL COMMENT '实际定植密度（株/m²）',
  `total_plants` INT UNSIGNED DEFAULT NULL COMMENT '总株数',
  `started_at` DATETIME(3) DEFAULT NULL,
  `ended_at` DATETIME(3) DEFAULT NULL,
  `expected_harvest_at` DATETIME(3) DEFAULT NULL,
  `recipe_version` VARCHAR(32) DEFAULT NULL,
  `policy_version` VARCHAR(32) DEFAULT NULL,
  `note` VARCHAR(255) DEFAULT NULL,
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_crop_batches_no` (`batch_no`),
  KEY `idx_crop_batches_greenhouse_status` (`greenhouse_id`, `status`, `started_at`),
  KEY `idx_crop_batches_zone` (`growing_zone_id`),
  CONSTRAINT `fk_batches_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_batches_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`),
  CONSTRAINT `fk_batches_variety` FOREIGN KEY (`crop_variety_id`) REFERENCES `crop_varieties` (`id`),
  CONSTRAINT `fk_batches_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `batch_stage_plans` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `growth_stage_id` BIGINT UNSIGNED NOT NULL,
  `stage_start_at` DATETIME(3) NOT NULL,
  `stage_end_at` DATETIME(3) NOT NULL,
  `target_ec_min` DECIMAL(12,4) DEFAULT NULL,
  `target_ec_max` DECIMAL(12,4) DEFAULT NULL,
  `target_ph_min` DECIMAL(12,4) DEFAULT NULL,
  `target_ph_max` DECIMAL(12,4) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_stage_plans` (`batch_id`, `growth_stage_id`, `stage_start_at`),
  CONSTRAINT `fk_stage_plan_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_stage_plan_stage` FOREIGN KEY (`growth_stage_id`) REFERENCES `growth_stages` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `harvest_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `harvested_at` DATETIME(3) NOT NULL,
  `harvest_weight_kg` DECIMAL(10,3) NOT NULL COMMENT '采收重量（kg）',
  `grade` VARCHAR(8) NOT NULL DEFAULT 'A' COMMENT '品质等级 A/B/C/Waste',
  `grade_weight_kg` DECIMAL(10,3) NOT NULL COMMENT '该等级重量',
  `note` VARCHAR(255) DEFAULT NULL,
  `harvested_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_harvest_batch_time` (`batch_id`, `harvested_at`),
  CONSTRAINT `fk_harvest_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_harvest_user` FOREIGN KEY (`harvested_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 7：营养液配方
-- ============================================================

CREATE TABLE `nutrient_recipes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `recipe_code` VARCHAR(64) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `crop_variety_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '适用的作物品种',
  `description` VARCHAR(255) DEFAULT NULL,
  `version` VARCHAR(32) NOT NULL DEFAULT 'v1',
  `status` VARCHAR(16) NOT NULL DEFAULT 'DRAFT' COMMENT 'DRAFT/ACTIVE/ARCHIVED',
  `effective_from` DATETIME(3) DEFAULT NULL,
  `effective_to` DATETIME(3) DEFAULT NULL,
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `published_by` BIGINT UNSIGNED DEFAULT NULL,
  `published_at` DATETIME(3) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_nutrient_recipes_code` (`recipe_code`),
  CONSTRAINT `fk_recipe_variety` FOREIGN KEY (`crop_variety_id`) REFERENCES `crop_varieties` (`id`),
  CONSTRAINT `fk_recipe_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_recipe_publisher` FOREIGN KEY (`published_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `recipe_stage_targets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `recipe_id` BIGINT UNSIGNED NOT NULL,
  `growth_stage_id` BIGINT UNSIGNED DEFAULT NULL,
  `metric_code` VARCHAR(32) NOT NULL,
  `target_min` DECIMAL(12,4) DEFAULT NULL,
  `target_max` DECIMAL(12,4) DEFAULT NULL,
  `tolerance` DECIMAL(12,4) DEFAULT NULL COMMENT '容差',
  `unit` VARCHAR(16) DEFAULT NULL,
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_recipe_targets_recipe_stage` (`recipe_id`, `growth_stage_id`, `enabled`),
  CONSTRAINT `fk_recipe_target_recipe` FOREIGN KEY (`recipe_id`) REFERENCES `nutrient_recipes` (`id`),
  CONSTRAINT `fk_recipe_target_stage` FOREIGN KEY (`growth_stage_id`) REFERENCES `growth_stages` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `recipe_ion_targets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `recipe_id` BIGINT UNSIGNED NOT NULL,
  `growth_stage_id` BIGINT UNSIGNED DEFAULT NULL,
  `ion_code` VARCHAR(8) NOT NULL COMMENT 'NO3_N/NH4_N/P/K/Ca/Mg/S/Fe/Mn/Zn/B/Cu/Mo',
  `target_min_mg_l` DECIMAL(10,4) DEFAULT NULL COMMENT '目标下限 (mg/L)',
  `target_max_mg_l` DECIMAL(10,4) DEFAULT NULL COMMENT '目标上限 (mg/L)',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recipe_ion_targets` (`recipe_id`, `growth_stage_id`, `ion_code`),
  CONSTRAINT `fk_ion_target_recipe` FOREIGN KEY (`recipe_id`) REFERENCES `nutrient_recipes` (`id`),
  CONSTRAINT `fk_ion_target_stage` FOREIGN KEY (`growth_stage_id`) REFERENCES `growth_stages` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `batch_recipe_bindings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `recipe_id` BIGINT UNSIGNED NOT NULL,
  `binding_type` VARCHAR(16) NOT NULL DEFAULT 'PRIMARY',
  `version` VARCHAR(32) NOT NULL DEFAULT 'v1',
  `effective_from` DATETIME(3) NOT NULL,
  `effective_to` DATETIME(3) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_batch_recipe_bindings_batch` (`batch_id`, `status`, `effective_from`),
  KEY `idx_batch_recipe_bindings_recipe` (`recipe_id`, `status`, `effective_from`),
  CONSTRAINT `fk_brb_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_brb_recipe` FOREIGN KEY (`recipe_id`) REFERENCES `nutrient_recipes` (`id`),
  CONSTRAINT `fk_brb_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 8：气候控制
-- ============================================================

CREATE TABLE `climate_profiles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(64) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `trigger_metric_code` VARCHAR(32) NOT NULL COMMENT '触发指标（TEMP/HUMIDITY/CO2）',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_climate_profiles_code` (`greenhouse_id`, `code`),
  CONSTRAINT `fk_climate_profile_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `climate_stages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `profile_id` BIGINT UNSIGNED NOT NULL,
  `stage_level` TINYINT UNSIGNED NOT NULL COMMENT '级别序号 1=最低, 2, 3...',
  `name` VARCHAR(64) NOT NULL COMMENT '如 "通风降温阶段1"',
  `trigger_operator` VARCHAR(4) NOT NULL COMMENT '>/>=/</<=' ,
  `trigger_threshold` DECIMAL(12,4) NOT NULL COMMENT '触发阈值',
  `hysteresis` DECIMAL(12,4) NOT NULL DEFAULT 1.0 COMMENT '回差（防抖）',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_climate_stages_level` (`profile_id`, `stage_level`),
  CONSTRAINT `fk_climate_stage_profile` FOREIGN KEY (`profile_id`) REFERENCES `climate_profiles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `climate_stage_actions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `stage_id` BIGINT UNSIGNED NOT NULL,
  `actuator_channel_id` BIGINT UNSIGNED NOT NULL,
  `command_type` VARCHAR(32) NOT NULL COMMENT 'SWITCH/SET_SPEED/SET_ANGLE',
  `command_payload` JSON NOT NULL COMMENT '目标状态',
  `execution_order` SMALLINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '动作执行顺序',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_climate_stage_action` (`stage_id`, `actuator_channel_id`, `execution_order`),
  CONSTRAINT `fk_csa_stage` FOREIGN KEY (`stage_id`) REFERENCES `climate_stages` (`id`),
  CONSTRAINT `fk_csa_actuator_channel` FOREIGN KEY (`actuator_channel_id`) REFERENCES `actuator_channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `climate_execution_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `profile_id` BIGINT UNSIGNED NOT NULL,
  `from_stage_level` TINYINT UNSIGNED DEFAULT NULL COMMENT '从哪个级别切换',
  `to_stage_level` TINYINT UNSIGNED NOT NULL COMMENT '切换到哪个级别',
  `trigger_value` DECIMAL(12,4) NOT NULL COMMENT '触发时的实际值',
  `executed_actions_count` INT UNSIGNED NOT NULL DEFAULT 0,
  `executed_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_climate_log_profile_time` (`profile_id`, `executed_at`),
  CONSTRAINT `fk_climate_log_profile` FOREIGN KEY (`profile_id`) REFERENCES `climate_profiles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 9：控制指令
-- ============================================================

CREATE TABLE `control_commands` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `actuator_channel_id` BIGINT UNSIGNED NOT NULL,
  `command_type` VARCHAR(32) NOT NULL,
  `payload` JSON NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/QUEUED/SENT/ACKED/TIMEOUT/FAILED',
  `sent_at` DATETIME(3) DEFAULT NULL,
  `acked_at` DATETIME(3) DEFAULT NULL,
  `request_id` VARCHAR(64) DEFAULT NULL,
  `created_by` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_control_commands_channel_time` (`actuator_channel_id`, `created_at`),
  KEY `idx_control_commands_status` (`status`),
  CONSTRAINT `fk_control_commands_channel` FOREIGN KEY (`actuator_channel_id`) REFERENCES `actuator_channels` (`id`),
  CONSTRAINT `fk_control_commands_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `control_command_receipts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `command_id` BIGINT UNSIGNED NOT NULL,
  `receipt_seq` INT UNSIGNED NOT NULL DEFAULT 1,
  `receipt_status` VARCHAR(16) NOT NULL,
  `ack_code` VARCHAR(32) DEFAULT NULL,
  `ack_message` VARCHAR(255) DEFAULT NULL,
  `ack_payload` JSON DEFAULT NULL,
  `ack_at` DATETIME(3) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_command_receipts_seq` (`command_id`, `receipt_seq`),
  CONSTRAINT `fk_receipt_command` FOREIGN KEY (`command_id`) REFERENCES `control_commands` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 10：策略引擎
-- ============================================================

CREATE TABLE `control_policies` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `policy_code` VARCHAR(64) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `policy_type` VARCHAR(16) NOT NULL COMMENT 'THRESHOLD/SCHEDULE/DURATION',
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL,
  `priority` INT NOT NULL DEFAULT 100,
  `retry_limit` TINYINT UNSIGNED NOT NULL DEFAULT 3,
  `timeout_sec` INT UNSIGNED NOT NULL DEFAULT 30,
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `version` VARCHAR(32) NOT NULL DEFAULT 'v1',
  `effective_from` DATETIME(3) DEFAULT NULL,
  `effective_to` DATETIME(3) DEFAULT NULL,
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `published_by` BIGINT UNSIGNED DEFAULT NULL,
  `published_at` DATETIME(3) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_control_policies_code` (`policy_code`),
  KEY `idx_policies_greenhouse_enabled` (`greenhouse_id`, `enabled`, `priority`),
  CONSTRAINT `fk_policy_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_policy_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`),
  CONSTRAINT `fk_policy_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_policy_publisher` FOREIGN KEY (`published_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `policy_conditions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `policy_id` BIGINT UNSIGNED NOT NULL,
  `metric_code` VARCHAR(32) NOT NULL,
  `operator` VARCHAR(8) NOT NULL,
  `threshold_value` DECIMAL(12,4) NOT NULL,
  `hysteresis` DECIMAL(12,4) DEFAULT NULL,
  `window_sec` INT UNSIGNED DEFAULT NULL COMMENT '聚合窗口',
  `required_duration_sec` INT UNSIGNED DEFAULT NULL COMMENT '持续满足时长',
  `aggregation` VARCHAR(16) DEFAULT NULL COMMENT 'avg/max/min/last',
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_policy_conditions_policy` (`policy_id`, `enabled`),
  CONSTRAINT `fk_pc_policy` FOREIGN KEY (`policy_id`) REFERENCES `control_policies` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `policy_targets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `policy_id` BIGINT UNSIGNED NOT NULL,
  `actuator_channel_id` BIGINT UNSIGNED NOT NULL,
  `command_type` VARCHAR(32) NOT NULL,
  `command_payload` JSON NOT NULL,
  `execution_order` SMALLINT UNSIGNED NOT NULL DEFAULT 1,
  `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_policy_targets_policy` (`policy_id`, `enabled`, `execution_order`),
  CONSTRAINT `fk_pt_policy` FOREIGN KEY (`policy_id`) REFERENCES `control_policies` (`id`),
  CONSTRAINT `fk_pt_actuator_channel` FOREIGN KEY (`actuator_channel_id`) REFERENCES `actuator_channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `policy_executions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `policy_id` BIGINT UNSIGNED NOT NULL,
  `trigger_source` VARCHAR(16) NOT NULL COMMENT 'TELEMETRY/SCHEDULE/MANUAL',
  `trigger_metric_code` VARCHAR(32) DEFAULT NULL,
  `trigger_value` DECIMAL(12,4) DEFAULT NULL,
  `decision` VARCHAR(16) NOT NULL COMMENT 'EXECUTED/SKIPPED/FAILED/CONFLICT',
  `decision_reason` VARCHAR(255) DEFAULT NULL,
  `command_id` BIGINT UNSIGNED DEFAULT NULL,
  `batch_id` BIGINT UNSIGNED DEFAULT NULL,
  `executed_at` DATETIME(3) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_policy_exec_policy_time` (`policy_id`, `created_at`),
  KEY `idx_policy_exec_decision_time` (`decision`, `created_at`),
  CONSTRAINT `fk_pe_policy` FOREIGN KEY (`policy_id`) REFERENCES `control_policies` (`id`),
  CONSTRAINT `fk_pe_command` FOREIGN KEY (`command_id`) REFERENCES `control_commands` (`id`),
  CONSTRAINT `fk_pe_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 11：告警系统
-- ============================================================

CREATE TABLE `alerts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` VARCHAR(32) NOT NULL COMMENT 'THRESHOLD/DEVICE_OFFLINE/SYSTEM',
  `level` VARCHAR(16) NOT NULL COMMENT 'INFO/WARN/CRITICAL',
  `metric_code` VARCHAR(32) DEFAULT NULL,
  `sensor_channel_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '传感器通道告警',
  `actuator_channel_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '执行器通道告警',
  `trigger_value` DECIMAL(12,4) DEFAULT NULL,
  `message` VARCHAR(255) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'OPEN' COMMENT 'OPEN/ACKNOWLEDGED/RESOLVED/IGNORED',
  `triggered_at` DATETIME(3) NOT NULL,
  `resolved_at` DATETIME(3) DEFAULT NULL,
  `resolved_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_alerts_status_triggered` (`status`, `triggered_at`),
  KEY `idx_alerts_sensor_ch` (`sensor_channel_id`, `triggered_at`),
  CONSTRAINT `fk_alerts_sensor_ch` FOREIGN KEY (`sensor_channel_id`) REFERENCES `sensor_channels` (`id`),
  CONSTRAINT `fk_alerts_actuator_ch` FOREIGN KEY (`actuator_channel_id`) REFERENCES `actuator_channels` (`id`),
  CONSTRAINT `fk_alerts_resolver` FOREIGN KEY (`resolved_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `alert_timeline_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `alert_id` BIGINT UNSIGNED NOT NULL,
  `event_type` VARCHAR(32) NOT NULL COMMENT 'TRIGGERED/AUTO_ACTION/MANUAL_ACTION/ACKNOWLEDGED/RESOLVED/COMMENT',
  `event_source` VARCHAR(16) NOT NULL COMMENT 'SYSTEM/MANUAL',
  `operator_id` BIGINT UNSIGNED DEFAULT NULL,
  `comment` VARCHAR(255) DEFAULT NULL,
  `event_payload` JSON DEFAULT NULL,
  `event_time` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_alert_timeline_alert_time` (`alert_id`, `event_time`),
  CONSTRAINT `fk_atv_alert` FOREIGN KEY (`alert_id`) REFERENCES `alerts` (`id`),
  CONSTRAINT `fk_atv_operator` FOREIGN KEY (`operator_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 12：能耗与资源消耗
-- ============================================================

CREATE TABLE `energy_consumption_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `record_type` VARCHAR(16) NOT NULL COMMENT 'ELECTRICITY/WATER/CO2_GAS',
  `consumption_value` DECIMAL(12,4) NOT NULL COMMENT '消耗量',
  `unit` VARCHAR(16) NOT NULL COMMENT 'kWh/m³/kg',
  `record_period_start` DATETIME(3) NOT NULL,
  `record_period_end` DATETIME(3) NOT NULL,
  `meter_reading_start` DECIMAL(12,4) DEFAULT NULL COMMENT '表头起始读数',
  `meter_reading_end` DECIMAL(12,4) DEFAULT NULL COMMENT '表头结束读数',
  `batch_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '分摊到批次',
  `recorded_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_energy_greenhouse_type_time` (`greenhouse_id`, `record_type`, `record_period_start`),
  KEY `idx_energy_batch` (`batch_id`),
  CONSTRAINT `fk_energy_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_energy_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_energy_user` FOREIGN KEY (`recorded_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `concentrate_usage_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inventory_id` BIGINT UNSIGNED NOT NULL,
  `solution_change_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的换液操作',
  `tank_id` BIGINT UNSIGNED DEFAULT NULL,
  `volume_used_ml` DECIMAL(10,2) NOT NULL,
  `used_by` BIGINT UNSIGNED DEFAULT NULL,
  `used_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_concentrate_usage_inventory` (`inventory_id`, `used_at`),
  CONSTRAINT `fk_cu_inventory` FOREIGN KEY (`inventory_id`) REFERENCES `nutrient_concentrate_inventory` (`id`),
  CONSTRAINT `fk_cu_solution_change` FOREIGN KEY (`solution_change_id`) REFERENCES `solution_change_events` (`id`),
  CONSTRAINT `fk_cu_tank` FOREIGN KEY (`tank_id`) REFERENCES `nutrient_tanks` (`id`),
  CONSTRAINT `fk_cu_user` FOREIGN KEY (`used_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 13：病虫害与植保
-- ============================================================

CREATE TABLE `pest_disease_observations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL,
  `batch_id` BIGINT UNSIGNED DEFAULT NULL,
  `observed_at` DATETIME(3) NOT NULL,
  `pest_or_disease` VARCHAR(64) NOT NULL COMMENT '病虫害名称（如 白粉病/蚜虫/红蜘蛛）',
  `severity` VARCHAR(16) NOT NULL COMMENT 'LIGHT/MODERATE/SEVERE',
  `affected_area_pct` DECIMAL(5,2) DEFAULT NULL COMMENT '受影响面积百分比',
  `affected_plant_count` INT UNSIGNED DEFAULT NULL,
  `symptoms` VARCHAR(255) DEFAULT NULL COMMENT '症状描述',
  `photo_urls` JSON DEFAULT NULL COMMENT '照片链接',
  `observed_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pd_obs_greenhouse_time` (`greenhouse_id`, `observed_at`),
  KEY `idx_pd_obs_batch` (`batch_id`),
  CONSTRAINT `fk_pd_obs_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_pd_obs_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`),
  CONSTRAINT `fk_pd_obs_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_pd_obs_user` FOREIGN KEY (`observed_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `treatment_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `observation_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的病虫害观察',
  `greenhouse_id` BIGINT UNSIGNED NOT NULL,
  `growing_zone_id` BIGINT UNSIGNED DEFAULT NULL,
  `batch_id` BIGINT UNSIGNED DEFAULT NULL,
  `treatment_type` VARCHAR(16) NOT NULL COMMENT 'CHEMICAL/BIOLOGICAL/PHYSICAL',
  `product_name` VARCHAR(128) NOT NULL COMMENT '药剂/天敌名称',
  `active_ingredient` VARCHAR(128) DEFAULT NULL COMMENT '有效成分',
  `dosage` VARCHAR(64) NOT NULL COMMENT '用量描述',
  `application_method` VARCHAR(32) NOT NULL COMMENT '施用方式 SPRAY/DRENCH/FOG/RELEASE',
  `safety_interval_days` INT UNSIGNED DEFAULT NULL COMMENT '安全间隔期（天）',
  `reentry_interval_hours` INT UNSIGNED DEFAULT NULL COMMENT '再进入间隔（小时）',
  `treated_at` DATETIME(3) NOT NULL,
  `treated_by` BIGINT UNSIGNED DEFAULT NULL,
  `note` VARCHAR(255) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_treatment_greenhouse_time` (`greenhouse_id`, `treated_at`),
  KEY `idx_treatment_batch_time` (`batch_id`, `treated_at`),
  CONSTRAINT `fk_treatment_obs` FOREIGN KEY (`observation_id`) REFERENCES `pest_disease_observations` (`id`),
  CONSTRAINT `fk_treatment_greenhouse` FOREIGN KEY (`greenhouse_id`) REFERENCES `greenhouses` (`id`),
  CONSTRAINT `fk_treatment_zone` FOREIGN KEY (`growing_zone_id`) REFERENCES `growing_zones` (`id`),
  CONSTRAINT `fk_treatment_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`),
  CONSTRAINT `fk_treatment_user` FOREIGN KEY (`treated_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 14：批次复盘
-- ============================================================

CREATE TABLE `batch_review_snapshots` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `snapshot_type` VARCHAR(16) NOT NULL DEFAULT 'DAILY' COMMENT 'DAILY/WEEKLY/STAGE_SUMMARY',
  `window_start` DATETIME(3) NOT NULL,
  `window_end` DATETIME(3) NOT NULL,
  `summary` JSON NOT NULL COMMENT '所有指标的平均值/最大值/最小值/告警次数/控制次数',
  `generated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_review_snapshots_window` (`batch_id`, `snapshot_type`, `window_start`, `window_end`),
  CONSTRAINT `fk_review_batch` FOREIGN KEY (`batch_id`) REFERENCES `crop_batches` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 域 15：用户与审计
-- ============================================================

CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(32) NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `nickname` VARCHAR(64) DEFAULT NULL,
  `phone` VARCHAR(32) DEFAULT NULL,
  `email` VARCHAR(64) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(32) NOT NULL,
  `description` VARCHAR(64) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles` (`user_id`, `role_id`),
  CONSTRAINT `fk_ur_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_ur_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `action` VARCHAR(64) NOT NULL,
  `target_type` VARCHAR(64) NOT NULL,
  `target_id` BIGINT UNSIGNED DEFAULT NULL,
  `detail` JSON DEFAULT NULL,
  `request_id` VARCHAR(64) DEFAULT NULL,
  `before_data` JSON DEFAULT NULL,
  `after_data` JSON DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_audit_user_time` (`user_id`, `created_at`),
  CONSTRAINT `fk_audit_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `notification_channels` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `channel_type` VARCHAR(16) NOT NULL COMMENT 'EMAIL/SMS/WEBHOOK/IN_APP',
  `name` VARCHAR(64) NOT NULL,
  `config` JSON NOT NULL,
  `min_alert_level` VARCHAR(16) NOT NULL DEFAULT 'WARN',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_nc_user_id` (`user_id`),
  KEY `idx_nc_enabled` (`enabled`),
  CONSTRAINT `fk_nc_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- SEED DATA
-- ============================================================

-- ── 角色 ──
INSERT INTO `roles` (`name`, `description`) VALUES
('ADMIN', '系统管理员'),
('OPERATOR', '操作员'),
('VIEWER', '查看者');

-- ── 管理员用户（密码: admin123） ──
INSERT INTO `users` (`username`, `password_hash`, `nickname`, `status`) VALUES
('admin', '$2a$10$BxC8C9oEQJ8QRjlhDLdXJ.adytUSppEPwIfJWLThOls6OEGfTuHcO', '默认管理员', 'ENABLED');

INSERT INTO `user_roles` (`user_id`, `role_id`)
SELECT u.id, r.id FROM `users` u, `roles` r
WHERE u.username = 'admin' AND r.name = 'ADMIN';

-- ── 指标定义 ──
INSERT INTO `metric_definitions` (`code`, `name`, `unit`, `precision_digits`, `normal_range_min`, `normal_range_max`, `is_core`) VALUES
('TEMP',     '温度',      '°C',     1, 18.0, 26.0, 1),
('HUMIDITY', '湿度',      '%',      1, 50.0, 80.0, 1),
('PH',       '酸碱度',    'pH',     1, 5.5,  6.5,  1),
('EC',       '电导率',    'mS/cm',  1, 1.2,  2.0,  1),
('DO',       '溶解氧',    'mg/L',   1, 5.0,  8.0,  1),
('WATER_TEMP','水温',     '°C',     1, 18.0, 24.0, 0),
('CO2',      '二氧化碳',  'ppm',    0, 400.0, 1200.0, 1),
('LIGHT',    '光照',      'lx',     0, 10000.0, 60000.0, 1);

-- ── 生长阶段 ──
INSERT INTO `growth_stages` (`code`, `name`, `sort_order`, `default_duration_days`) VALUES
('SEEDLING',       '苗期',     1, 14),
('VEGETATIVE',     '营养生长期', 2, 21),
('TRANSITION',     '转色期',   3, 14),
('FRUITING',       '结果期',   4, 30),
('HARVEST',        '采收期',   5, 14);

-- ── 作物品种 ──
INSERT INTO `crop_varieties` (`code`, `name`, `default_cycle_days`) VALUES
('LETTUCE_BUTTER',  '奶油生菜', 45),
('LETTUCE_LOOSE',   '散叶生菜', 40),
('TOMATO_CHERRY',   '樱桃番茄', 120),
('CUCUMBER',        '黄瓜',      90),
('STRAWBERRY',      '草莓',      150);
