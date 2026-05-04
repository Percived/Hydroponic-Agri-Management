-- ============================================================
-- 种子数据：自动控制规则
-- 覆盖温度、湿度、pH、EC、CO2、光照六类指标
-- 1号温室 12 条 + 2号温室 10 条 = 共 22 条规则
-- 每条规则使用回差（hysteresis）设计，避免阈值附近反复触发
-- ============================================================

INSERT INTO `control_rules` (`name`, `metric_id`, `operator`, `threshold`, `action`, `target_device_id`, `enabled`, `created_by`)
SELECT rules.rule_name, m.id, rules.operator, rules.threshold, rules.action, rules.target_device_id, rules.enabled, 1
FROM (
  -- ═══════════════ 1号温室（叶菜水培）- 温度 ═══════════════
  SELECT 'GH1-高温开风机' AS rule_name, 'TEMP' AS metric_code, '>' AS operator, 30.0 AS threshold,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')) AS action,
    18 AS target_device_id, 1 AS enabled
  UNION ALL
  SELECT 'GH1-正常关风机', 'TEMP', '<=', 28.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    18, 1

  -- ═══════════════ 1号温室 - 湿度 ═══════════════
  UNION ALL
  SELECT 'GH1-高湿开风机', 'HUMIDITY', '>', 85.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    18, 1
  UNION ALL
  SELECT 'GH1-正常关风机除湿', 'HUMIDITY', '<=', 75.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    18, 1

  -- ═══════════════ 1号温室 - pH ═══════════════
  UNION ALL
  SELECT 'GH1-pH过高开循环泵', 'PH', '>', 6.5,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    17, 1
  UNION ALL
  SELECT 'GH1-pH正常停循环泵', 'PH', '<=', 6.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    17, 1

  -- ═══════════════ 1号温室 - 电导率（营养液浓度） ═══════════════
  UNION ALL
  SELECT 'GH1-浓度低开供液阀', 'EC', '<', 1.5,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    19, 1
  UNION ALL
  SELECT 'GH1-浓度正常关供液阀', 'EC', '>=', 2.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    19, 1

  -- ═══════════════ 1号温室 - CO2 ═══════════════
  UNION ALL
  SELECT 'GH1-CO2低开风机换气', 'CO2', '<', 400.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    18, 1
  UNION ALL
  SELECT 'GH1-CO2正常关风机', 'CO2', '>=', 600.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    18, 1

  -- ═══════════════ 1号温室 - 光照 ═══════════════
  UNION ALL
  SELECT 'GH1-强光开风机降温', 'LIGHT', '>', 80000.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    18, 1
  UNION ALL
  SELECT 'GH1-光照正常关风机', 'LIGHT', '<=', 60000.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    18, 1

  -- ═══════════════ 2号温室（草莓水培）- 温度 ═══════════════
  UNION ALL
  SELECT 'GH2-高温开风机', 'TEMP', '>', 30.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    28, 1
  UNION ALL
  SELECT 'GH2-正常关风机', 'TEMP', '<=', 28.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    28, 1

  -- ═══════════════ 2号温室 - 湿度 ═══════════════
  UNION ALL
  SELECT 'GH2-高湿开风机', 'HUMIDITY', '>', 85.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    28, 1
  UNION ALL
  SELECT 'GH2-正常关风机除湿', 'HUMIDITY', '<=', 75.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    28, 1

  -- ═══════════════ 2号温室 - pH ═══════════════
  UNION ALL
  SELECT 'GH2-pH过高开循环泵', 'PH', '>', 6.5,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    27, 1
  UNION ALL
  SELECT 'GH2-pH正常停循环泵', 'PH', '<=', 6.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    27, 1

  -- ═══════════════ 2号温室 - 电导率 ═══════════════
  UNION ALL
  SELECT 'GH2-浓度低开循环泵', 'EC', '<', 1.2,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    27, 1
  UNION ALL
  SELECT 'GH2-浓度正常停循环泵', 'EC', '>=', 1.8,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    27, 1

  -- ═══════════════ 2号温室 - CO2 ═══════════════
  UNION ALL
  SELECT 'GH2-CO2低开风机换气', 'CO2', '<', 400.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'ON')),
    28, 1
  UNION ALL
  SELECT 'GH2-CO2正常关风机', 'CO2', '>=', 600.0,
    JSON_OBJECT('command_type', 'SWITCH', 'payload', JSON_OBJECT('state', 'OFF')),
    28, 1
) AS rules
JOIN `metrics` m ON m.code = rules.metric_code;
