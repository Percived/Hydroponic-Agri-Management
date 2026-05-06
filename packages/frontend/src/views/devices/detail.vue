<template>
    <div class="device-detail-page">
      <div class="page-header">
        <el-button @click="goBack" :icon="ArrowLeft">返回列表</el-button>
        <h1 class="page-title">设备详情 - {{ device?.device_code || '加载中...' }}</h1>
        <el-tag v-if="device" :type="deviceType === 'sensor' ? 'primary' : 'warning'" size="large">
          {{ deviceType === 'sensor' ? '传感器设备' : '执行器设备' }}
        </el-tag>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="8" animated />
      </div>

      <!-- 错误状态 -->
      <div v-else-if="errorMsg" class="error-container">
        <el-result icon="error" :title="errorMsg" sub-title="请检查网络连接或返回列表重试">
          <template #extra>
            <el-button type="primary" @click="loadData">重新加载</el-button>
            <el-button @click="goBack">返回列表</el-button>
          </template>
        </el-result>
      </div>

      <!-- 设备数据 -->
      <template v-else-if="device">
        <el-tabs v-model="activeTab" type="border-card">
          <!-- Tab 1: 基本信息 -->
          <el-tab-pane label="基本信息" name="info">
            <el-card class="info-card">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="设备编码">{{ device.device_code }}</el-descriptions-item>
                <el-descriptions-item label="设备名称">{{ device.name }}</el-descriptions-item>
                <el-descriptions-item label="设备型号">{{ device.model || '-' }}</el-descriptions-item>
                <el-descriptions-item label="固件版本">{{ device.firmware_version || '-' }}</el-descriptions-item>
                <el-descriptions-item label="通信协议">{{ device.protocol }}</el-descriptions-item>
                <el-descriptions-item label="所属温室">{{ getGreenhouseName(device.greenhouse_id) }}</el-descriptions-item>
                <el-descriptions-item label="种植区">{{ getZoneName(device.growing_zone_id) }}</el-descriptions-item>
                <el-descriptions-item label="在线状态">
                  <el-tag :type="device.status === 'ONLINE' ? 'success' : device.status === 'FAULT' ? 'warning' : 'danger'">
                    {{ device.status === 'ONLINE' ? '在线' : device.status === 'FAULT' ? '故障' : '离线' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="最后上报">
                  {{ device.last_seen_at ? formatDate(device.last_seen_at) : '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ formatDate(device.created_at) }}</el-descriptions-item>
                <el-descriptions-item label="更新时间">{{ formatDate(device.updated_at) }}</el-descriptions-item>
              </el-descriptions>
            </el-card>

            <!-- 通道列表 -->
            <el-card v-if="channels.length > 0" class="channels-card">
              <template #header>
                <span>{{ deviceType === 'sensor' ? '传感器通道' : '执行器通道' }}</span>
              </template>
              <el-table :data="channels" stripe size="small">
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="channel_code" label="通道编码" width="130" />
                <template v-if="deviceType === 'sensor'">
                  <el-table-column prop="metric_code" label="指标代码" width="100">
                    <template #default="{ row }">{{ getMetricName(row.metric_code) }}</template>
                  </el-table-column>
                  <el-table-column prop="unit" label="单位" width="80" />
                  <el-table-column prop="precision_digits" label="精度" width="70" />
                  <el-table-column label="量程" width="150">
                    <template #default="{ row }">{{ row.range_min ?? '-' }} ~ {{ row.range_max ?? '-' }}</template>
                  </el-table-column>
                  <el-table-column prop="sampling_interval_sec" label="采样间隔(s)" width="110" />
                </template>
                <template v-else>
                  <el-table-column prop="actuator_type" label="类型" width="100" />
                  <el-table-column prop="current_state" label="当前状态" width="100" />
                  <el-table-column prop="rated_power_watt" label="额定功率(W)" width="120">
                    <template #default="{ row }">{{ row.rated_power_watt ?? '-' }}</template>
                  </el-table-column>
                </template>
                <el-table-column prop="enabled" label="启用" width="70">
                  <template #default="{ row }">
                    <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
                      {{ row.enabled === 1 ? '是' : '否' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="last_reported_at" label="最后上报" width="160">
                  <template #default="{ row }">{{ formatDate(row.last_reported_at) }}</template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <!-- Tab 2: 遥测数据（仅传感器设备） -->
          <el-tab-pane v-if="deviceType === 'sensor'" label="遥测数据" name="telemetry">
            <div v-if="telemetryLoading" class="loading-placeholder">
              <el-skeleton :rows="3" animated />
            </div>
            <div v-else-if="telemetryByChannel.length === 0" class="empty-placeholder">
              <el-empty description="暂无遥测数据" />
            </div>
            <div v-else class="telemetry-grid">
              <div v-for="item in telemetryByChannel" :key="`${item.channel_id}-${item.metric_code}`" class="telemetry-item">
                <div class="metric-name">{{ getMetricName(item.metric_code) }}</div>
                <div class="metric-value">
                  {{ formatNumber(item.value) }}
                  <span class="metric-unit">{{ getChannelUnit(item.channel_id) || '' }}</span>
                </div>
                <div class="metric-meta">
                  <el-tag :type="item.quality_flag === 'normal' ? 'success' : 'danger'" size="small">
                    {{ item.quality_flag === 'normal' ? '正常' : item.quality_flag }}
                  </el-tag>
                  <span class="metric-time">{{ formatDate(item.collected_at, 'HH:mm:ss') }}</span>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </template>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { deviceApi, telemetryApi, greenhouseApi } from '@/api'
import { formatDate, formatNumber, getMetricName } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { SensorDevice, ActuatorDevice, SensorChannel, ActuatorChannel, TelemetryRecord, Greenhouse, GrowingZone } from '@/types'

const route = useRoute()
const router = useRouter()

const deviceId = computed(() => Number(route.params.id))

const loading = ref(false)
const errorMsg = ref('')
const activeTab = ref('info')
const telemetryLoading = ref(false)

// Determine device type (sensor or actuator) from route query or by trial
const deviceType = ref<'sensor' | 'actuator'>((route.query.type as 'sensor' | 'actuator') || 'sensor')

type AnyDevice = SensorDevice | ActuatorDevice
const device = ref<AnyDevice | null>(null)
const channels = ref<SensorChannel[] | ActuatorChannel[]>([])
const telemetryByChannel = ref<Array<TelemetryRecord & { channel_id: number }>>([])
const greenhouses = ref<Greenhouse[]>([])
const growingZones = ref<GrowingZone[]>([])

// Channel unit lookup map
const channelUnitMap = computed(() => {
  const map = new Map<number, string>()
  for (const ch of channels.value) {
    if ('unit' in ch) {
      map.set((ch as SensorChannel).id, (ch as SensorChannel).unit)
    }
  }
  return map
})

function getChannelUnit(channelId: number): string {
  return channelUnitMap.value.get(channelId) || ''
}

function getGreenhouseName(id: number): string {
  const gh = greenhouses.value.find(g => g.id === id)
  return gh?.name || '-'
}

function getZoneName(id: number | undefined | null): string {
  if (!id) return '-'
  const zone = growingZones.value.find(z => z.id === id)
  return zone?.name || '-'
}

function goBack() { router.push('/devices') }

async function loadData() {
  if (!deviceId.value) { errorMsg.value = '无效的设备 ID'; return }
  loading.value = true
  errorMsg.value = ''

  try {
    if (deviceType.value === 'sensor') {
      device.value = await deviceApi.getSensorDevice(deviceId.value)
      const chData = await deviceApi.getSensorChannels({
        sensor_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE
      })
      channels.value = chData.items
    } else {
      device.value = await deviceApi.getActuatorDevice(deviceId.value)
      const chData = await deviceApi.getActuatorChannels({
        actuator_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE
      })
      channels.value = chData.items
    }
  } catch (e: any) {
    // If first attempt fails, try the other device type
    if (deviceType.value === 'sensor') {
      try {
        deviceType.value = 'actuator'
        device.value = await deviceApi.getActuatorDevice(deviceId.value)
        const chData = await deviceApi.getActuatorChannels({
          actuator_device_id: deviceId.value,
          page_size: LARGE_PAGE_SIZE
        })
        channels.value = chData.items
      } catch {
        errorMsg.value = e?.message || '加载设备信息失败'
      }
    } else {
      errorMsg.value = e?.message || '加载设备信息失败'
    }
    return
  } finally {
    loading.value = false
  }

  // Load greenhouses and zones for display
  loadGreenhouses()
  if (device.value?.greenhouse_id) {
    loadGrowingZones(device.value.greenhouse_id)
  }

  // Load telemetry for sensor channels
  if (deviceType.value === 'sensor') {
    fetchTelemetry()
  }
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { /* ignore */ }
}

async function loadGrowingZones(greenhouseId: number) {
  try {
    const data = await greenhouseApi.getGrowingZones({ greenhouse_id: greenhouseId, page_size: LARGE_PAGE_SIZE })
    growingZones.value = data.items
  } catch { /* ignore */ }
}

async function fetchTelemetry() {
  const sensorChs = channels.value as SensorChannel[]
  if (sensorChs.length === 0) return

  telemetryLoading.value = true
  try {
    const results: Array<TelemetryRecord & { channel_id: number }> = []
    for (const ch of sensorChs) {
      try {
        const record = await telemetryApi.getChannelLatest(ch.id)
        results.push({ ...record, channel_id: ch.id })
      } catch {
        // Skip channels with no data
      }
    }
    telemetryByChannel.value = results
  } finally {
    telemetryLoading.value = false
  }
}

onMounted(() => { loadData() })
</script>

<style scoped lang="scss">
.device-detail-page {
  .page-header {
    display: flex; align-items: center; gap: 16px; margin-bottom: 20px;
  }
  .page-title {
    font-size: 22px; font-weight: 700; color: var(--color-text-primary); margin: 0; flex: 1;
  }
  .loading-container, .error-container {
    padding: 40px; background: var(--bg-card); border-radius: var(--radius-md);
  }
  .info-card, .channels-card {
    margin-bottom: 16px;
  }
  .loading-placeholder, .empty-placeholder {
    padding: 20px; text-align: center; color: var(--color-text-secondary);
  }
  .telemetry-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 16px;
  }
  .telemetry-item {
    padding: 16px; background: var(--color-primary-bg-light); border-radius: var(--radius-md); border: 1px solid var(--border-color-light);
  }
  .metric-name { font-size: 14px; color: var(--color-text-regular); margin-bottom: 8px; }
  .metric-value { font-size: 24px; font-weight: 600; color: var(--color-text-primary); margin-bottom: 8px; }
  .metric-unit { font-size: 14px; font-weight: normal; color: var(--color-text-secondary); }
  .metric-meta { display: flex; align-items: center; gap: 8px; }
  .metric-time { font-size: 12px; color: var(--color-text-secondary); }
}
</style>
