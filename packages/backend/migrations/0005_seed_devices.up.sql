-- ============================================================
-- 种子数据：温室 + 设备分组 + 设备
-- 2 个温室 / 6 个分组 / 19 个设备（12 传感器 + 7 执行器）
-- 所有 ID 从 10 开始，避免与已有数据冲突
-- ============================================================

-- ═══ 温室 ═══
INSERT INTO `greenhouses` (`id`, `name`, `location`, `description`) VALUES
(10, '1号温室', 'A区-东侧', '叶菜水培种植区，主要种植生菜、菠菜'),
(11, '2号温室', 'B区-西侧', '草莓水培种植区，品种为章姬草莓');

-- ═══ 设备分组 ═══
INSERT INTO `device_groups` (`id`, `greenhouse_id`, `name`, `description`) VALUES
(10, 10, '营养液监控组',   '1号温室营养液循环系统的传感器组'),
(11, 10, '环境监测组',     '1号温室空气环境参数监测'),
(12, 10, '执行设备组',     '1号温室执行器设备（泵、风机、阀门）'),
(13, 11, '营养液监控组',   '2号温室营养液循环系统的传感器组'),
(14, 11, '环境监测组',     '2号温室空气环境参数监测'),
(15, 11, '执行设备组',     '2号温室执行器设备（泵、风机）');

-- ═══ 设备 ═══
-- 1号温室 - 营养液监控组 (group_id=10)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(10, 'GH1-TEMP-01', '营养液温度传感器', 'SENSOR', 'TEMP',     10, 10, 'ENABLED', 'MQTT', 30),
(11, 'GH1-PH-01',   '营养液pH传感器',   'SENSOR', 'PH',       10, 10, 'ENABLED', 'MQTT', 30),
(12, 'GH1-EC-01',   '营养液EC传感器',   'SENSOR', 'EC',       10, 10, 'ENABLED', 'MQTT', 30);

-- 1号温室 - 环境监测组 (group_id=11)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(13, 'GH1-TEMP-02', '环境温度传感器',   'SENSOR', 'TEMP',     10, 11, 'ENABLED', 'MQTT', 30),
(14, 'GH1-HUM-01',  '环境湿度传感器',   'SENSOR', 'HUMIDITY', 10, 11, 'ENABLED', 'MQTT', 60),
(15, 'GH1-CO2-01',  '二氧化碳传感器',   'SENSOR', 'CO2',      10, 11, 'ENABLED', 'MQTT', 60),
(16, 'GH1-LIGHT-01','光照强度传感器',   'SENSOR', 'LIGHT',    10, 11, 'ENABLED', 'MQTT', 60);

-- 1号温室 - 执行设备组 (group_id=12)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(17, 'GH1-PUMP-01', '营养液循环泵',     'ACTUATOR', 'PUMP',  10, 12, 'ENABLED', 'MQTT', 0),
(18, 'GH1-FAN-01',  '通风风机',         'ACTUATOR', 'FAN',   10, 12, 'ENABLED', 'MQTT', 0),
(19, 'GH1-VALVE-01','供液电磁阀',       'ACTUATOR', 'VALVE', 10, 12, 'ENABLED', 'MQTT', 0);

-- 2号温室 - 营养液监控组 (group_id=13)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(20, 'GH2-TEMP-01', '营养液温度传感器', 'SENSOR', 'TEMP',     11, 13, 'ENABLED', 'MQTT', 30),
(21, 'GH2-PH-01',   '营养液pH传感器',   'SENSOR', 'PH',       11, 13, 'ENABLED', 'MQTT', 30),
(22, 'GH2-EC-01',   '营养液EC传感器',   'SENSOR', 'EC',       11, 13, 'ENABLED', 'MQTT', 30);

-- 2号温室 - 环境监测组 (group_id=14)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(23, 'GH2-TEMP-02', '环境温度传感器',   'SENSOR', 'TEMP',     11, 14, 'ENABLED', 'MQTT', 30),
(24, 'GH2-HUM-01',  '环境湿度传感器',   'SENSOR', 'HUMIDITY', 11, 14, 'ENABLED', 'MQTT', 60),
(25, 'GH2-CO2-01',  '二氧化碳传感器',   'SENSOR', 'CO2',      11, 14, 'ENABLED', 'MQTT', 60),
(26, 'GH2-LIGHT-01','光照强度传感器',   'SENSOR', 'LIGHT',    11, 14, 'ENABLED', 'MQTT', 60);

-- 2号温室 - 执行设备组 (group_id=15)
INSERT INTO `devices` (`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `sampling_interval_sec`) VALUES
(27, 'GH2-PUMP-01', '营养液循环泵',     'ACTUATOR', 'PUMP',  11, 15, 'ENABLED', 'MQTT', 0),
(28, 'GH2-FAN-01',  '通风风机',         'ACTUATOR', 'FAN',   11, 15, 'ENABLED', 'MQTT', 0);
