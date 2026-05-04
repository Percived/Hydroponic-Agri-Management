DELETE FROM `metrics`
WHERE `code` IN ('TEMP', 'HUMIDITY', 'PH', 'EC', 'CO2', 'LIGHT');
