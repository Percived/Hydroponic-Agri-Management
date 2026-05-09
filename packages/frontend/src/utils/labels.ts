import type {
  Greenhouse,
  GrowingZone,
  NutrientTank,
  CropBatch,
  CropVariety,
  GrowthStage,
  SensorDevice,
  SensorChannel,
  ActuatorDevice,
  ActuatorChannel
} from '@/types'

export function fallbackIdLabel(prefix: string, id?: number | null): string {
  if (id == null) return `${prefix}#-`
  return `${prefix}#${id}`
}

export function buildIdLabelMap<T>(
  items: T[],
  getId: (t: T) => number,
  getLabel: (t: T) => string,
  fallbackPrefix: string
): Record<number, string> {
  const map: Record<number, string> = {}
  for (const it of items) {
    const id = getId(it)
    const label = (getLabel(it) || '').trim()
    map[id] = label ? label : fallbackIdLabel(fallbackPrefix, id)
  }
  return map
}

export function pickNameCodeLabel(name?: string, code?: string): string {
  const n = (name || '').trim()
  const c = (code || '').trim()
  if (n && c && n !== c) return `${n}/${c}`
  return n || c
}

export function greenhouseLabel(g: Greenhouse): string {
  return pickNameCodeLabel(g.name, g.code) || fallbackIdLabel('温室', g.id)
}

export function growingZoneLabel(z: GrowingZone): string {
  return pickNameCodeLabel(z.name, z.code) || fallbackIdLabel('种植区', z.id)
}

export function nutrientTankLabel(t: NutrientTank): string {
  const code = (t.code || '').trim()
  return code || fallbackIdLabel('液槽', t.id)
}

export function cropBatchLabel(b: CropBatch): string {
  const no = (b.batch_no || '').trim()
  return no || fallbackIdLabel('批次', b.id)
}

export function cropVarietyLabel(v: CropVariety): string {
  return pickNameCodeLabel(v.name, v.code) || fallbackIdLabel('作物', v.id)
}

export function growthStageLabel(s: GrowthStage): string {
  return pickNameCodeLabel(s.name, s.code) || fallbackIdLabel('阶段', s.id)
}

export function sensorDeviceLabel(d: SensorDevice): string {
  return pickNameCodeLabel(d.name, d.device_code) || fallbackIdLabel('设备', d.id)
}

export function sensorChannelLabel(ch: SensorChannel, sensorDeviceLabelById: Record<number, string>): string {
  const dev = sensorDeviceLabelById[ch.sensor_device_id] || fallbackIdLabel('设备', ch.sensor_device_id)
  return `${ch.channel_code} (${dev})`
}

export function actuatorDeviceLabel(d: ActuatorDevice): string {
  return pickNameCodeLabel(d.name, d.device_code) || fallbackIdLabel('设备', d.id)
}

export function actuatorChannelLabel(ch: ActuatorChannel, actuatorDeviceLabelById: Record<number, string>): string {
  const dev = actuatorDeviceLabelById[ch.actuator_device_id] || fallbackIdLabel('设备', ch.actuator_device_id)
  return `${ch.channel_code} (${dev})`
}
