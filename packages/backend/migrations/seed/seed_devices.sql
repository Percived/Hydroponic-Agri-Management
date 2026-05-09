-- ============================================================
-- 传感器设备种子数据
-- 用于测试采集器从注册到上报的完整流程
-- ============================================================

-- ── 1. 温室 ──
INSERT INTO `greenhouses` (`code`, `name`, `location`, `area_sqm`, `description`, `status`) VALUES
('GH-001', '1号水培温室', 'A区-东侧', 500.00, '主生产温室，配备环境自动调控系统', 'ENABLED');

-- ── 2. 种植区（关联温室） ──
INSERT INTO `growing_zones` (`greenhouse_id`, `code`, `name`, `system_type`, `tank_volume_liter`, `planting_density_per_sqm`, `status`)
SELECT g.id, 'ZONE-DWC-01', 'DWC深水栽培区', 'DWC', 2000.00, 25.00, 'ENABLED'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `growing_zones` (`greenhouse_id`, `code`, `name`, `system_type`, `tank_volume_liter`, `planting_density_per_sqm`, `status`)
SELECT g.id, 'ZONE-NFT-01', 'NFT管道栽培区', 'NFT', 500.00, 30.00, 'ENABLED'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `growing_zones` (`greenhouse_id`, `code`, `name`, `system_type`, `tank_volume_liter`, `planting_density_per_sqm`, `status`)
SELECT g.id, 'ZONE-SEED-01', '育苗区', 'DWC', 500.00, 40.00, 'ENABLED'
FROM `greenhouses` g WHERE g.code = 'GH-001';


-- ── 3. 传感器设备 ──

-- 3.1 温室环境综合传感器 A（温室级，不绑定区域）
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, NULL, 'GH1-ENV-01', '温室环境监测节点A',
       'ESP32-ENV-PRO', '2.1.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g WHERE g.code = 'GH-001';

-- 4 个通道：温度 / 湿度 / CO2 / 光照
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'TEMP',      'TEMP',     '°C',    1,  -10.0,  50.0,   10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'HUMIDITY',  'HUMIDITY', '%',      1,    0.0, 100.0,   10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'CO2',       'CO2',      'ppm',    0,    0.0, 5000.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'LIGHT',     'LIGHT',    'lx',     0,    0.0, 100000.0, 30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-01';

-- 3.2 温室环境监测节点B（冗余传感器，温室级）
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, NULL, 'GH1-ENV-02', '温室环境监测节点B',
       'ESP32-ENV-LITE', '2.0.1', 'ONLINE', 'MQTT'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'TEMP',     'TEMP',     '°C',  1, -10.0, 50.0,  10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-02';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'HUMIDITY', 'HUMIDITY', '%',   1,   0.0, 100.0, 10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-ENV-02';

-- 3.3 DWC区水质传感器
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, z.id, 'GH1-WQ-DWC-01', 'DWC区水质监测节点',
       'ESP32-WQ-PRO', '3.0.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'PH',        'PH',        'pH',    1,  0.0, 14.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'EC',        'EC',        'mS/cm', 1,  0.0, 10.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'DO',        'DO',        'mg/L',  1,  0.0, 20.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'WATER_TEMP','WATER_TEMP','°C',    1,  0.0, 40.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-DWC-01';

-- 3.4 NFT区水质传感器
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, z.id, 'GH1-WQ-NFT-01', 'NFT区水质监测节点',
       'ESP32-WQ-LITE', '2.5.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-NFT-01';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'PH',        'PH',        'pH',    1,  0.0, 14.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-NFT-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'EC',        'EC',        'mS/cm', 1,  0.0, 10.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-NFT-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'WATER_TEMP','WATER_TEMP','°C',    1,  0.0, 40.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-NFT-01';

-- 3.5 育苗区水质传感器
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, z.id, 'GH1-WQ-SEED-01', '育苗区水质监测节点',
       'ESP32-WQ-LITE', '2.5.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-SEED-01';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'PH',        'PH',        'pH',    1,  0.0, 14.0,  60, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-SEED-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'EC',        'EC',        'mS/cm', 1,  0.0, 10.0,  60, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-SEED-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'WATER_TEMP','WATER_TEMP','°C',    1,  0.0, 40.0,  60, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WQ-SEED-01';

-- 3.6 DWC区水处理监测传感器
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, z.id, 'GH1-WT-DWC-01', 'DWC区水处理监测节点',
       'ESP32-WT-PRO', '1.0.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'LEVEL',     'LEVEL',     'cm',    1,   0.0, 200.0,  10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WT-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'TDS',       'TDS',       'ppm',   0,   0.0, 5000.0, 30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WT-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'TURBIDITY', 'TURBIDITY', 'NTU',   1,   0.0, 100.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WT-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'FLOW_RATE', 'FLOW_RATE', 'L/min', 1,   0.0, 50.0,   10, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-WT-DWC-01';

-- 3.7 DWC区安全监测传感器
INSERT INTO `sensor_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`)
SELECT g.id, z.id, 'GH1-SAFE-DWC-01', 'DWC区安全监测节点',
       'ESP32-SAFE-PRO', '1.0.0', 'ONLINE', 'MQTT'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01';

INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'ORP',       'ORP',       'mV',    0,   0.0, 800.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-SAFE-DWC-01';
INSERT INTO `sensor_channels` (`sensor_device_id`, `channel_code`, `metric_code`, `unit`, `precision_digits`, `range_min`, `range_max`, `sampling_interval_sec`, `enabled`)
SELECT d.id, 'O3',        'O3',        'ppb',   1,   0.0, 200.0,  30, 1 FROM `sensor_devices` d WHERE d.device_code = 'GH1-SAFE-DWC-01';

-- ============================================================
-- 汇总
-- ============================================================
-- 温室:   1 (GH-001)
-- 种植区: 3 (ZONE-DWC-01, ZONE-NFT-01, ZONE-SEED-01)
-- 传感器: 7 设备, 22 通道
--   GH1-ENV-01:      4ch (TEMP, HUMIDITY, CO2, LIGHT)        温室级
--   GH1-ENV-02:      2ch (TEMP, HUMIDITY)                     温室级
--   GH1-WQ-DWC-01:   4ch (PH, EC, DO, WATER_TEMP)            DWC区
--   GH1-WQ-NFT-01:   3ch (PH, EC, WATER_TEMP)                NFT区
--   GH1-WQ-SEED-01:  3ch (PH, EC, WATER_TEMP)                育苗区
--   GH1-WT-DWC-01:   4ch (LEVEL, TDS, TURBIDITY, FLOW_RATE)  DWC区水处理
--   GH1-SAFE-DWC-01: 2ch (ORP, O3)                           DWC区安全
-- ============================================================
