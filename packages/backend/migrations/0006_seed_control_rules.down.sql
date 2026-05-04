-- 清除种子控制规则（按名称模式匹配）
DELETE FROM `control_rules`
WHERE `name` LIKE 'GH1-%' OR `name` LIKE 'GH2-%';
