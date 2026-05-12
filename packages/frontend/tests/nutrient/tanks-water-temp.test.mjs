import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentFileDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(currentFileDir, '..', '..')
const tanksViewPath = resolve(projectRoot, 'src', 'views', 'nutrient', 'tanks.vue')

test('营养液槽页面只允许为温度槽位选择水温通道', () => {
  const content = readFileSync(tanksViewPath, 'utf8')

  assert.match(content, /label="水温传感器通道"/)
  assert.match(content, /metric_code === 'WATER_TEMP'/)
  assert.doesNotMatch(content, /metric_code === 'TEMP'/)
  assert.doesNotMatch(content, /metric_code === 'TEMPERATURE'/)
})

test('营养液槽页面更新时会显式发送 null 以解绑传感器通道', () => {
  const content = readFileSync(tanksViewPath, 'utf8')

  assert.match(content, /payload\.ec_sensor_channel_id = formData\.ec_sensor_channel_id \?\? null/)
  assert.match(content, /payload\.ph_sensor_channel_id = formData\.ph_sensor_channel_id \?\? null/)
  assert.match(content, /payload\.level_sensor_channel_id = formData\.level_sensor_channel_id \?\? null/)
  assert.match(content, /payload\.temp_sensor_channel_id = formData\.temp_sensor_channel_id \?\? null/)
})
