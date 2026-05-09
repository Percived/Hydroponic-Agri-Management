-- ============================================================
-- 执行器设备种子数据
-- 用于测试执行器控制、策略联动、告警触发等完整流程
-- 依赖：需先执行 seed_devices.sql（温室和种植区数据）
-- ============================================================

-- ── 执行器设备 ──

-- 1. DWC区继电器控制模块（绑定 ZONE-DWC-01）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, z.id, 'GH1-RELAY-DWC-01', 'DWC区继电器控制模块',
       'ESP32-RELAY-8CH', '3.1.0', 'ONLINE', 'MQTT', '{"channels":8,"rated_input":"12V DC"}'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01';

-- DWC区通道：循环水泵 / 曝气泵 / 加热棒 / 补液阀
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH1-PUMP',    'PUMP',    'ON',  150.00, 1, '{"label":"循环水泵","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-DWC-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH2-AERATOR', 'AERATOR', 'ON',   50.00, 1, '{"label":"曝气泵","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-DWC-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH3-HEATER',  'HEATER',  'OFF', 500.00, 1, '{"label":"加热棒","control":"RELAY","temp_range":"18-28°C"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-DWC-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH4-VALVE',   'VALVE',   'OFF',   0.00, 1, '{"label":"补液电磁阀","control":"RELAY","type":"NC"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-DWC-01';


-- 2. NFT区继电器控制模块（绑定 ZONE-NFT-01）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, z.id, 'GH1-RELAY-NFT-01', 'NFT区继电器控制模块',
       'ESP32-RELAY-4CH', '3.0.0', 'ONLINE', 'MQTT', '{"channels":4,"rated_input":"12V DC"}'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-NFT-01';

-- NFT区通道：循环水泵 / 供液阀 / 加热棒
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH1-PUMP',   'PUMP',   'ON',  100.00, 1, '{"label":"循环水泵","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-NFT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH2-VALVE',  'VALVE',  'OFF',   0.00, 1, '{"label":"供液电磁阀","control":"RELAY","type":"NC"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-NFT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH3-HEATER', 'HEATER', 'OFF', 300.00, 1, '{"label":"加热棒","control":"RELAY","temp_range":"18-28°C"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-NFT-01';


-- 3. 育苗区继电器控制模块（绑定 ZONE-SEED-01）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, z.id, 'GH1-RELAY-SEED-01', '育苗区继电器控制模块',
       'ESP32-RELAY-4CH', '3.0.0', 'ONLINE', 'MQTT', '{"channels":4,"rated_input":"12V DC"}'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-SEED-01';

-- 育苗区通道：循环水泵 / 曝气泵 / 加热灯 / 雾化器
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH1-PUMP',    'PUMP',    'ON',   80.00, 1, '{"label":"循环水泵","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-SEED-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH2-AERATOR', 'AERATOR', 'ON',   30.00, 1, '{"label":"曝气泵","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-SEED-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH3-LED',     'LED',     'OFF', 200.00, 1, '{"label":"育苗补光灯","control":"PWM","spectrum":"full"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-SEED-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH4-FOGGER',  'FOGGER',  'OFF',  60.00, 1, '{"label":"超声波雾化器","control":"RELAY","flow_rate":"300ml/h"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-SEED-01';


-- 4. 温室环境控制主控（温室级，不绑区域 — 含降温/通风/遮阳/CO2/加湿）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, NULL, 'GH1-ENV-CTRL-01', '温室环境控制主控',
       'ESP32-ENV-CTRL-PRO', '4.0.0', 'ONLINE', 'MQTT', '{"channels":8,"rated_input":"24V DC","has_pwm":true}'
FROM `greenhouses` g WHERE g.code = 'GH-001';

-- 主控通道
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FAN-EXHAUST-1', 'FAN',    'OFF', 200.00, 1, '{"label":"排风扇-南","control":"RELAY","airflow":"5000m³/h"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FAN-EXHAUST-2', 'FAN',    'OFF', 200.00, 1, '{"label":"排风扇-北","control":"RELAY","airflow":"5000m³/h"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FAN-CIRC',       'FAN',    'ON',  80.00, 1, '{"label":"内循环风机","control":"PWM","speed":"0-100%"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'SHADE-TOP',      'SHADE',  'OFF', 100.00, 1, '{"label":"顶遮阳帘","control":"RELAY_REV","close_pct":0}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'LED-SUP',        'LED',    'OFF', 600.00, 1, '{"label":"顶部补光灯组","control":"PWM","spectrum":"full","ppfd":"800μmol/m²/s"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CO2-GEN',        'CO2_GEN','OFF',  50.00, 1, '{"label":"CO2发生器","control":"RELAY","target_ppm":800}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FOGGER-1',       'FOGGER', 'OFF',  80.00, 1, '{"label":"高压雾化加湿器","control":"RELAY","flow_rate":"5L/h"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-01';


-- 5. 温室环境控制备控（温室级 — 冗余/备用，用于测试离线/故障场景）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, NULL, 'GH1-ENV-CTRL-02', '温室环境控制备控',
       'ESP32-ENV-CTRL-LITE', '3.5.0', 'OFFLINE', 'MQTT', '{"channels":4,"rated_input":"24V DC","role":"backup"}'
FROM `greenhouses` g WHERE g.code = 'GH-001';

-- 备控通道（部分 disabled，用于测试 enabled=false 场景）
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FAN-BACKUP', 'FAN',   'OFF', 150.00, 1, '{"label":"备用排风扇","control":"RELAY"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-02';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'SHADE-SIDE', 'SHADE', 'OFF',  80.00, 1, '{"label":"侧遮阳帘","control":"RELAY_REV"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-02';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'HEATER-AUX', 'HEATER','OFF', 800.00, 0, '{"label":"辅助加热器","control":"RELAY","reason":"季节性停用"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-CTRL-02';


-- 6. 故障设备（用于测试 FAULT 状态和告警联动）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, NULL, 'GH1-RELAY-FAULT-01', '故障模拟继电器',
       'ESP32-RELAY-2CH', '1.0.0', 'FAULT', 'MQTT', '{"error":"CH1 过流保护触发","last_error_at":"2026-05-09T08:30:00+08:00"}'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH1-PUMP',   'PUMP',   'OFF', 200.00, 0, '{"label":"故障水泵","control":"RELAY","fault":"overcurrent"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-FAULT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CH2-VALVE',  'VALVE',  'OFF',   0.00, 0, '{"label":"故障电磁阀","control":"RELAY","fault":"stuck_closed"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-RELAY-FAULT-01';


-- 7. 营养液自动化控制模块（绑定 ZONE-DWC-01）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, z.id, 'GH1-NUTRI-AUTO-01', '营养液自动化控制模块',
       'ESP32-NUTRI-CTRL', '3.2.0', 'ONLINE', 'MQTT', '{"channels":3,"rated_input":"12V DC","control":"RELAY/PULSE"}'
FROM `greenhouses` g
JOIN `growing_zones` z ON z.greenhouse_id = g.id
WHERE g.code = 'GH-001' AND z.code = 'ZONE-DWC-01';

INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'DOSING-PUMP-A', 'DOSING_PUMP', 'OFF', 25.00, 1, '{"label":"A液计量泵","control":"PULSE","flow_rate":"0.1-100ml/min"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-NUTRI-AUTO-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CHILLER-1', 'CHILLER', 'OFF', 500.00, 1, '{"label":"营养液冷水机","control":"RELAY","temp_range":"5-25°C"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-NUTRI-AUTO-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'STIRRER-1', 'STIRRER', 'ON', 120.00, 1, '{"label":"母液搅拌器","control":"RELAY","rpm":"60-300"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-NUTRI-AUTO-01';


-- 8. 水处理与消毒模块（温室级）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, NULL, 'GH1-WATER-TRT-01', '水处理与消毒模块',
       'ESP32-WATER-CTRL', '3.0.0', 'ONLINE', 'MQTT', '{"channels":5,"rated_input":"24V DC"}'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'UV-STER-1', 'UV_STERILIZER', 'ON', 40.00, 1, '{"label":"紫外线杀菌灯","control":"RELAY","wavelength":"254nm"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-WATER-TRT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'OZONE-GEN-1', 'OZONE_GENERATOR', 'OFF', 100.00, 1, '{"label":"臭氧发生器","control":"RELAY","output":"5g/h"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-WATER-TRT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'FILTER-1', 'FILTER', 'ON', 80.00, 1, '{"label":"多介质过滤器","control":"RELAY","mesh":"120","backwash":"true"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-WATER-TRT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'RO-SYS-1', 'RO_SYSTEM', 'OFF', 300.00, 1, '{"label":"RO反渗透系统","control":"RELAY","tds_removal":"95%"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-WATER-TRT-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'TOP-UP-1', 'TOP_UP_VALVE', 'OFF', 0.00, 1, '{"label":"自动补水阀","control":"RELAY","trigger":"low_level"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-WATER-TRT-01';


-- 9. 环境与安全辅助模块（温室级）
INSERT INTO `actuator_devices` (`greenhouse_id`, `growing_zone_id`, `device_code`, `name`, `model`, `firmware_version`, `status`, `protocol`, `metadata`)
SELECT g.id, NULL, 'GH1-ENV-AUX-01', '环境与安全辅助模块',
       'ESP32-AUX-CTRL', '3.1.0', 'ONLINE', 'MQTT', '{"channels":4,"rated_input":"12V DC"}'
FROM `greenhouses` g WHERE g.code = 'GH-001';

INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'DEHUM-1', 'DEHUMIDIFIER', 'OFF', 400.00, 1, '{"label":"工业除湿机","control":"RELAY","capacity":"20L/day"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-AUX-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'DAMPER-IN', 'DAMPER', 'OFF', 10.00, 1, '{"label":"进风口电动风阀","control":"SERVO","angle":"0-90°"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-AUX-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'ALARM-SYS', 'ALARM', 'OFF', 15.00, 1, '{"label":"声光报警器","control":"RELAY","db":"120dB","flash":"true"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-AUX-01';
INSERT INTO `actuator_channels` (`actuator_device_id`, `channel_code`, `actuator_type`, `current_state`, `rated_power_watt`, `enabled`, `metadata`)
SELECT d.id, 'CALIB-VALVE', 'CALIBRATION_VALVE', 'OFF', 5.00, 1, '{"label":"校准切换阀","control":"SOLENOID","ports":"3-way"}'
FROM `actuator_devices` d WHERE d.device_code = 'GH1-ENV-AUX-01';


-- ============================================================
-- 汇总
-- ============================================================
-- 执行器设备: 9 台（7 ONLINE / 1 OFFLINE / 1 FAULT）
-- 执行器通道: 35 个
--
--   GH1-RELAY-DWC-01    4ch (PUMP, AERATOR, HEATER, VALVE)                 DWC区
--   GH1-RELAY-NFT-01    3ch (PUMP, VALVE, HEATER)                          NFT区
--   GH1-RELAY-SEED-01   4ch (PUMP, AERATOR, LED, FOGGER)                   育苗区
--   GH1-ENV-CTRL-01     7ch (FAN×3, SHADE, LED, CO2_GEN, FOGGER)          温室级
--   GH1-ENV-CTRL-02     3ch (FAN, SHADE, HEATER)                           温室级(备)
--   GH1-RELAY-FAULT-01  2ch (PUMP, VALVE)                                  温室级(故障)
--   GH1-NUTRI-AUTO-01   3ch (DOSING_PUMP, CHILLER, STIRRER)                DWC区
--   GH1-WATER-TRT-01    5ch (UV_STERILIZER, OZONE_GENERATOR, FILTER, RO_SYSTEM, TOP_UP_VALVE) 温室级
--   GH1-ENV-AUX-01      4ch (DEHUMIDIFIER, DAMPER, ALARM, CALIBRATION_VALVE) 温室级
--
-- 执行器类型分布:
--   PUMP:              4   (DWC/NFT/育苗循环泵 + 1故障)
--   AERATOR:           2   (DWC/育苗曝气泵)
--   HEATER:            3   (DWC/NFT/备控)
--   VALVE:             3   (DWC补液/NFT供液/故障)
--   FAN:               4   (排风扇×2/循环风机/备用风机)
--   SHADE:             2   (顶遮阳/侧遮阳)
--   LED:               2   (育苗补光/顶部补光)
--   CO2_GEN:           1   (CO2发生器)
--   FOGGER:            2   (育苗雾化/温室加湿)
--   DOSING_PUMP:       1   (营养液计量泵)
--   CHILLER:           1   (营养液冷水机)
--   STIRRER:           1   (母液搅拌器)
--   DEHUMIDIFIER:      1   (工业除湿机)
--   DAMPER:            1   (电动风阀)
--   UV_STERILIZER:     1   (紫外线杀菌灯)
--   OZONE_GENERATOR:   1   (臭氧发生器)
--   FILTER:            1   (多介质过滤器)
--   RO_SYSTEM:         1   (RO反渗透系统)
--   TOP_UP_VALVE:      1   (自动补水阀)
--   ALARM:             1   (声光报警器)
--   CALIBRATION_VALVE: 1   (校准切换阀)
--
-- 与传感器种子数据合并后:
--   温室:     1
--   种植区:   3
--   传感器:   5 设备, 16 通道
--   执行器:   9 设备, 35 通道
--   总计:    14 设备, 51 通道
-- ============================================================
