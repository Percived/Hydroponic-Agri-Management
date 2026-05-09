-- ============================================================
-- 控制策略种子数据
-- 依赖：需先执行 seed_devices.sql + seed_actuators.sql
-- 覆盖：温室环境控制 / 水肥管理 / 定时任务 / 安全报警
-- ============================================================

-- ============================================================
-- 一、THRESHOLD（阈值触发）策略 — 温室级环境控制
-- ============================================================

-- 1. 高温通风降温：TEMP > 30°C → 排风扇×2 + 遮阳帘
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-001', '高温通风降温', 'THRESHOLD', g.id, NULL, 10, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'TEMP', '>', 30.0, 1.0, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-001';

-- 排风扇-南 ON
INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-001' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'FAN-EXHAUST-1';

-- 排风扇-北 ON
INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-001' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'FAN-EXHAUST-2';

-- 遮阳帘 ON
INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 3, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-001' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'SHADE-TOP';


-- 2. 低温保护：TEMP < 15°C（持续120s）→ 关闭风阀 + 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-002', '低温保护', 'THRESHOLD', g.id, NULL, 10, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'TEMP', '<', 15.0, 1.0, 120, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-002';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-002' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'DAMPER-IN';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-002' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- 3. 高湿除湿：HUMIDITY > 85% → 除湿机 + 内循环风机
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-003', '高湿除湿', 'THRESHOLD', g.id, NULL, 15, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'HUMIDITY', '>', 85.0, 2.0, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-003';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-003' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'DEHUM-1';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-003' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'FAN-CIRC';


-- 4. 低湿加湿：HUMIDITY < 60% → 雾化加湿器 80%功率
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-004', '低湿加湿', 'THRESHOLD', g.id, NULL, 15, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'HUMIDITY', '<', 60.0, 2.0, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-004';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SET', '{"state":"ON","value":80}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-004' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'FOGGER-1';


-- 5. CO2补充：CO2 < 400ppm → CO2发生器
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-005', 'CO2补充', 'THRESHOLD', g.id, NULL, 20, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'CO2', '<', 400.0, 20.0, 60, 'min', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-005';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-005' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'CO2-GEN';


-- 6. 强光遮阳：LIGHT > 80000 lux → 遮阳帘
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-GH-006', '强光遮阳', 'THRESHOLD', g.id, NULL, 10, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'LIGHT', '>', 80000.0, 5000.0, 30, 'max', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-GH-006';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-GH-006' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'SHADE-TOP';


-- ============================================================
-- 二、THRESHOLD 策略 — DWC区水肥管理
-- ============================================================

-- 7. DWC水温过高：WATER_TEMP > 25°C → 冷水机
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-001', 'DWC水温过高降温', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'WATER_TEMP', '>', 25.0, 0.5, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-001';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-001' AND ad.device_code = 'GH1-NUTRI-AUTO-01' AND ac.channel_code = 'CHILLER-1';


-- 8. DWC水温过低：WATER_TEMP < 18°C → 加热棒
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-002', 'DWC水温过低加热', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'WATER_TEMP', '<', 18.0, 0.5, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-002';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-002' AND ad.device_code = 'GH1-RELAY-DWC-01' AND ac.channel_code = 'CH3-HEATER';


-- 9. DWC pH偏高：PH > 6.5 → 计量泵（加酸）
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-003', 'DWC pH偏高加酸', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'PH', '>', 6.5, 0.1, 60, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-003';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'ACTIVATE', '{"state":"ON","duration_sec":10,"type":"ACID"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-003' AND ad.device_code = 'GH1-NUTRI-AUTO-01' AND ac.channel_code = 'DOSING-PUMP-A';


-- 10. DWC pH偏低：PH < 5.5 → 计量泵（加碱）+ 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-004', 'DWC pH偏低加碱', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'PH', '<', 5.5, 0.1, 60, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-004';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'ACTIVATE', '{"state":"ON","duration_sec":10,"type":"ALKALI"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-004' AND ad.device_code = 'GH1-NUTRI-AUTO-01' AND ac.channel_code = 'DOSING-PUMP-A';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-004' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- 11. DWC EC偏高：EC > 2.5 mS/cm → 补水阀（稀释）
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-005', 'DWC EC偏高稀释', 'THRESHOLD', g.id, z.id, 8, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'EC', '>', 2.5, 0.1, 120, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-005';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-005' AND ad.device_code = 'GH1-WATER-TRT-01' AND ac.channel_code = 'TOP-UP-1';


-- 12. DWC 溶氧不足：DO < 4.0 mg/L → 曝气泵
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-006', 'DWC溶氧不足曝气', 'THRESHOLD', g.id, z.id, 3, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'DO', '<', 4.0, 0.5, 60, 'min', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-006';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-006' AND ad.device_code = 'GH1-RELAY-DWC-01' AND ac.channel_code = 'CH2-AERATOR';


-- 13. DWC 液位过低：LEVEL < 20% → 补水阀 + 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-007', 'DWC液位过低补水', 'THRESHOLD', g.id, z.id, 2, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'LEVEL', '<', 20.0, 2.0, 30, 'last', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-007';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-007' AND ad.device_code = 'GH1-WATER-TRT-01' AND ac.channel_code = 'TOP-UP-1';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-007' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- 14. DWC 水质安全 — TDS偏高：TDS > 1500 ppm → 补水稀释 + 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-DWC-008', 'DWC TDS偏高稀释', 'THRESHOLD', g.id, z.id, 10, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'TDS', '>', 1500.0, 50.0, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-DWC-008';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-008' AND ad.device_code = 'GH1-WATER-TRT-01' AND ac.channel_code = 'TOP-UP-1';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 2, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-DWC-008' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- ============================================================
-- 三、THRESHOLD 策略 — NFT区监控报警
-- ============================================================

-- 15. NFT水温过高：WATER_TEMP > 28°C → 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-NFT-001', 'NFT水温过高报警', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-NFT-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'WATER_TEMP', '>', 28.0, 0.5, 60, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-NFT-001';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-NFT-001' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- 16. NFT pH偏高：PH > 7.0 → 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-NFT-002', 'NFT pH偏高报警', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-NFT-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'PH', '>', 7.0, 0.1, 60, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-NFT-002';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-NFT-002' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- 17. NFT pH偏低：PH < 5.0 → 报警
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-NFT-003', 'NFT pH偏低报警', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-NFT-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `required_duration_sec`, `aggregation`, `enabled`)
SELECT p.id, 'PH', '<', 5.0, 0.1, 60, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-NFT-003';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-NFT-003' AND ad.device_code = 'GH1-ENV-AUX-01' AND ac.channel_code = 'ALARM-SYS';


-- ============================================================
-- 四、THRESHOLD 策略 — 育苗区
-- ============================================================

-- 18. 育苗区水温过低：WATER_TEMP < 20°C → LED补光灯（供暖）
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-SEED-001', '育苗区水温过低加温', 'THRESHOLD', g.id, z.id, 5, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-SEED-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'WATER_TEMP', '<', 20.0, 0.5, 60, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-SEED-001';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SET', '{"state":"ON","value":100}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-SEED-001' AND ad.device_code = 'GH1-RELAY-SEED-01' AND ac.channel_code = 'CH3-LED';


-- 19. 育苗区湿度过低：HUMIDITY < 70% → 雾化器
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'TH-SEED-002', '育苗区湿度过低加湿', 'THRESHOLD', g.id, z.id, 8, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `growing_zones` z, `users` u
WHERE g.code = 'GH-001' AND z.code = 'ZONE-SEED-01' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `hysteresis`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'HUMIDITY', '<', 70.0, 2.0, 120, 'avg', 1
FROM `control_policies` p WHERE p.policy_code = 'TH-SEED-002';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'TH-SEED-002' AND ad.device_code = 'GH1-RELAY-SEED-01' AND ac.channel_code = 'CH4-FOGGER';


-- ============================================================
-- 五、SCHEDULE（定时）策略
-- ============================================================

-- 20. LED日间补光：每30s扫描, CO2 < 2000（总是满足）→ 顶部补光灯组
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'SC-GH-001', 'LED日间补光', 'SCHEDULE', g.id, NULL, 50, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'CO2', '<', 2000.0, 30, 'last', 1
FROM `control_policies` p WHERE p.policy_code = 'SC-GH-001';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SET', '{"state":"ON","value":80}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'SC-GH-001' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'LED-SUP';


-- 21. 循环风扇定时：每30s扫描, TEMP > 0（总是满足）→ 内循环风机 60%转速
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'SC-GH-002', '内循环风扇常开', 'SCHEDULE', g.id, NULL, 100, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'TEMP', '>', 0.0, 30, 'last', 1
FROM `control_policies` p WHERE p.policy_code = 'SC-GH-002';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SET', '{"state":"ON","value":60}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'SC-GH-002' AND ad.device_code = 'GH1-ENV-CTRL-01' AND ac.channel_code = 'FAN-CIRC';


-- ============================================================
-- 六、DURATION（持续时长）策略 — 示例
-- ============================================================

-- 22. RO反渗透定时冲洗：FLOW_RATE < 100（正常范围）→ RO系统 ON
INSERT INTO `control_policies` (`policy_code`, `name`, `policy_type`, `greenhouse_id`, `growing_zone_id`, `priority`, `retry_limit`, `timeout_sec`, `enabled`, `version`, `created_by`)
SELECT 'DR-WTR-001', 'RO系统定时运行', 'DURATION', g.id, NULL, 80, 3, 30, 1, 'v1', u.id
FROM `greenhouses` g, `users` u
WHERE g.code = 'GH-001' AND u.username = 'admin';

INSERT INTO `policy_conditions` (`policy_id`, `metric_code`, `operator`, `threshold_value`, `window_sec`, `aggregation`, `enabled`)
SELECT p.id, 'FLOW_RATE', '<', 100.0, 30, 'last', 1
FROM `control_policies` p WHERE p.policy_code = 'DR-WTR-001';

INSERT INTO `policy_targets` (`policy_id`, `actuator_channel_id`, `command_type`, `command_payload`, `execution_order`, `enabled`)
SELECT p.id, ac.id, 'SWITCH', '{"state":"ON"}', 1, 1
FROM `control_policies` p, `actuator_channels` ac
JOIN `actuator_devices` ad ON ad.id = ac.actuator_device_id
WHERE p.policy_code = 'DR-WTR-001' AND ad.device_code = 'GH1-WATER-TRT-01' AND ac.channel_code = 'RO-SYS-1';


-- ============================================================
-- 汇总
-- ============================================================
-- 控制策略: 22 条
--
-- 温室级 THRESHOLD:  6 条 (TH-GH-001 ~ TH-GH-006)
--   - 高温通风 / 低温保护 / 高湿除湿 / 低湿加湿 / CO2补充 / 强光遮阳
-- DWC区 THRESHOLD:  8 条 (TH-DWC-001 ~ TH-DWC-008)
--   - 水温高/低 / pH高/低 / EC高 / 溶氧低 / 液位低 / TDS高
-- NFT区 THRESHOLD:  3 条 (TH-NFT-001 ~ TH-NFT-003)
--   - 水温过高报警 / pH偏高报警 / pH偏低报警
-- 育苗区 THRESHOLD: 2 条 (TH-SEED-001 ~ TH-SEED-002)
--   - 水温过低加温 / 湿度过低加湿
-- 温室级 SCHEDULE:  2 条 (SC-GH-001 ~ SC-GH-002)
--   - LED日间补光 / 内循环风扇常开
-- 温室级 DURATION:  1 条 (DR-WTR-001)
--   - RO系统定时运行
--
-- 策略类型分布: THRESHOLD=19 / SCHEDULE=2 / DURATION=1
-- 条件总数: 23 条
-- 目标总数: 27 条
-- ============================================================
