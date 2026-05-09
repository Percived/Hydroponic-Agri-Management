// 执行器类型标签映射（单一真实来源）
export const ACTUATOR_TYPE_LABELS: Record<string, string> = {
  PUMP: '水泵',
  AERATOR: '曝气泵',
  FAN: '风机',
  VALVE: '电磁阀',
  SHADE: '遮阳帘',
  LED: '补光灯',
  HEATER: '加热器',
  CO2_GEN: 'CO2发生器',
  FOGGER: '雾化器',
  DOSING_PUMP: '计量泵',
  CHILLER: '冷水机',
  STIRRER: '搅拌器',
  DEHUMIDIFIER: '除湿机',
  DAMPER: '电动风阀',
  UV_STERILIZER: '紫外线杀菌灯',
  OZONE_GENERATOR: '臭氧发生器',
  FILTER: '过滤器',
  RO_SYSTEM: 'RO反渗透系统',
  TOP_UP_VALVE: '自动补水阀',
  ALARM: '声光报警器',
  CALIBRATION_VALVE: '校准切换阀',
}

// 供下拉框使用
export const actuatorTypeOptions = Object.entries(ACTUATOR_TYPE_LABELS).map(
  ([value, label]) => ({ label, value })
)
