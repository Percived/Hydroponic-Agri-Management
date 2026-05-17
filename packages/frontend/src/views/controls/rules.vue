<template>
  <div class="rules-page">
    <div class="page-header">
      <h1 class="page-title">策略管理</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增策略
      </el-button>
    </div>

    <div class="table-container">
      <el-table :data="policies" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="policy_code" label="策略编码" width="150" />
        <el-table-column prop="name" label="策略名称" min-width="160" />
        <el-table-column prop="policy_type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.policy_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="计划描述" min-width="170">
          <template #default="{ row }">
            <span>{{ formatScheduleSummary(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="90" />
        <el-table-column prop="retry_limit" label="重试" width="80" />
        <el-table-column prop="timeout_sec" label="超时(s)" width="90" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="isPolicyEnabled(row) ? 'success' : 'info'">{{ isPolicyEnabled(row) ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布状态" width="110">
          <template #default="{ row }">
            <el-tag :type="isPolicyPublished(row) ? 'success' : 'warning'">
              {{ isPolicyPublished(row) ? '已发布' : '未发布' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="warning" link :disabled="isPolicyPublished(row)" @click="handlePublish(row)">
              {{ isPolicyPublished(row) ? '已发布' : '发布' }}
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="760px">
      <el-form :model="formData" label-width="120px">
        <el-form-item label="策略编码" required>
          <el-input v-model="formData.policy_code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="策略名称" required>
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="所属温室" required>
              <el-select v-model="formData.greenhouse_id" placeholder="请选择温室" filterable style="width: 100%" @change="onGreenhouseChange">
                <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="种植区">
              <el-select v-model="formData.growing_zone_id" placeholder="可选" clearable style="width: 100%" @change="onGrowingZoneChange">
                <el-option v-for="zone in growingZones" :key="zone.id" :label="zone.name" :value="zone.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-input-number v-model="formData.priority" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="策略类型" required>
              <el-select v-model="formData.policy_type" style="width: 100%">
                <el-option label="阈值 THRESHOLD" value="THRESHOLD" />
                <el-option label="定时 SCHEDULE" value="SCHEDULE" />
                <el-option label="时长 DURATION" value="DURATION" :disabled="true" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="重试次数">
              <el-input-number v-model="formData.retry_limit" :min="0" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="超时(秒)">
              <el-input-number v-model="formData.timeout_sec" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="生效开始">
              <el-date-picker
                v-model="formData.effective_from"
                type="datetime"
                placeholder="可选"
                clearable
                style="width: 100%"
                format="YYYY-MM-DD HH:mm"
                :disabled-date="disabledPastDate"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="生效结束">
              <el-date-picker
                v-model="formData.effective_to"
                type="datetime"
                placeholder="可选"
                clearable
                style="width: 100%"
                format="YYYY-MM-DD HH:mm"
                :disabled-date="disabledInvalidToDate"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- DURATION 未实现提示 -->
        <el-alert
          v-if="formData.policy_type === 'DURATION'"
          title="时长策略后端尚未实现，请选择阈值或定时策略"
          type="warning"
          :closable="false"
          style="margin-bottom: 16px"
        />

        <template v-if="formData.policy_type === 'SCHEDULE'">
          <el-divider>执行计划</el-divider>
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="计划类型" required>
                <el-select v-model="formData.schedule_mode" style="width: 100%">
                  <el-option label="单次执行" value="ONCE" />
                  <el-option label="每日执行" value="DAILY" />
                  <el-option label="每周执行" value="WEEKLY" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="时区">
                <el-input v-model="formData.timezone" disabled />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item v-if="formData.schedule_mode === 'ONCE'" label="执行时间" required>
            <el-date-picker
              v-model="formData.run_once_at"
              type="datetime"
              placeholder="选择单次执行时间"
              style="width: 100%"
              format="YYYY-MM-DD HH:mm:ss"
            />
          </el-form-item>

          <el-form-item v-if="formData.schedule_mode === 'DAILY'" label="每日时刻" required>
            <el-time-picker
              v-model="formData.time_of_day"
              placeholder="选择每日执行时刻"
              style="width: 100%"
              format="HH:mm:ss"
              value-format="HH:mm:ss"
            />
          </el-form-item>

          <template v-if="formData.schedule_mode === 'WEEKLY'">
            <el-form-item label="每周时刻" required>
              <el-time-picker
                v-model="formData.time_of_day"
                placeholder="选择每周执行时刻"
                style="width: 100%"
                format="HH:mm:ss"
                value-format="HH:mm:ss"
              />
            </el-form-item>
            <el-form-item label="执行星期" required>
              <el-checkbox-group v-model="formData.weekdays">
                <el-checkbox-button v-for="item in weekdayOptions" :key="item.value" :label="item.value">
                  {{ item.label }}
                </el-checkbox-button>
              </el-checkbox-group>
            </el-form-item>
          </template>
        </template>

        <!-- THRESHOLD: 策略条件 -->
        <template v-if="formData.policy_type === 'THRESHOLD'">
          <el-divider>策略条件</el-divider>

          <div v-for="(cond, index) in conditions" :key="index" class="target-item">
            <div class="target-item-header">
              <span class="target-item-title">触发条件 {{ index + 1 }}</span>
              <el-button
                v-if="conditions.length > 1"
                type="danger"
                size="small"
                plain
                @click="removeCondition(index)"
              >
                删除
              </el-button>
            </div>
            <el-form-item label="指标代码" required>
              <el-select v-model="cond.metric_code" placeholder="选择指标" style="width: 100%">
                <el-option v-for="m in metrics" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
              </el-select>
            </el-form-item>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="运算符" required>
                  <el-select v-model="cond.operator" style="width: 100%">
                    <el-option label="大于 >" value=">" />
                    <el-option label="大于等于 >=" value=">=" />
                    <el-option label="小于 <" value="<" />
                    <el-option label="小于等于 <=" value="<=" />
                    <el-option label="等于 =" value="=" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="阈值" required>
                  <el-input-number v-model="cond.threshold_value" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="滞后值">
                  <el-input-number v-model="cond.hysteresis" :min="0" :precision="2" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="窗口(秒)">
                  <el-input-number v-model="cond.window_sec" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="持续(秒)">
                  <el-input-number v-model="cond.required_duration_sec" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="聚合方式">
                  <el-select v-model="cond.aggregation" placeholder="默认 last" clearable style="width: 100%">
                    <el-option label="最新 last" value="last" />
                    <el-option label="平均 avg" value="avg" />
                    <el-option label="最大 max" value="max" />
                    <el-option label="最小 min" value="min" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
          </div>
          <el-button type="primary" plain @click="addCondition" style="width: 100%; margin-bottom: 16px;">
            + 添加触发条件 (满足所有条件时触发)
          </el-button>
        </template>

        <el-divider>目标动作</el-divider>
        <div v-for="(t, index) in targets" :key="index" class="target-item">
          <div class="target-item-header">
            <span class="target-item-title">目标动作 {{ index + 1 }}</span>
            <el-button
              v-if="targets.length > 1"
              type="danger"
              size="small"
              plain
              @click="removeTarget(index)"
            >
              删除
            </el-button>
          </div>
          <el-form-item label="执行器通道" required>
            <el-select
              v-model="t.channelId"
              placeholder="选择执行器通道"
              filterable
              style="width: 100%"
              @change="(val: number | undefined) => onTargetChannelChange(index, val)"
            >
              <el-option
                v-for="ch in actuatorChannels"
                :key="ch.id"
                :label="`${ch.channel_code} (${actuatorTypeLabels[ch.actuator_type] || ch.actuator_type}) - ${ch.device_name || ''}`"
                :value="ch.id"
              />
            </el-select>
            <div v-if="t.channelId && getSelectedActuator(t.channelId)" class="metric-hints">
              <span class="metric-hints-label">可影响指标：</span>
              <template v-if="getMetricHints(t.channelId).length">
                <el-tag
                  v-for="m in getMetricHints(t.channelId)"
                  :key="m.code"
                  size="small"
                  type="info"
                  class="metric-hint-tag"
                >
                  {{ metricLabelMap[m.code] || m.code }}{{ m.unit ? ` (${m.unit})` : '' }}
                </el-tag>
              </template>
              <span v-else class="metric-hints-none">无明确指标影响</span>
            </div>
          </el-form-item>
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="命令类型" required>
                <el-select v-model="t.commandType" style="width: 100%" @change="onCommandTypeChange(t)">
                  <el-option label="开关" value="SWITCH" />
                  <el-option label="设置值" value="SET_VALUE" />
                  <el-option label="校准" value="CALIBRATE" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="执行顺序">
                <el-input-number v-model="t.executionOrder" :min="0" :max="99" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="命令负载" required>
            <!-- SWITCH: ON/OFF select -->
            <el-select
              v-if="t.commandType === 'SWITCH'"
              v-model="t.switchState"
              style="width: 100%"
              @change="syncStructuredToPayload(t)"
            >
              <el-option label="开启 ON" value="ON" />
              <el-option label="关闭 OFF" value="OFF" />
            </el-select>
            <!-- SET_VALUE: key (with suggestions) + value -->
            <el-row v-else-if="t.commandType === 'SET_VALUE'" :gutter="8">
              <el-col :span="8">
                <el-select
                  v-model="t.payloadKey"
                  filterable
                  allow-create
                  default-first-option
                  placeholder="参数名"
                  style="width: 100%"
                  @change="syncStructuredToPayload(t)"
                >
                  <el-option
                    v-for="p in getParamSuggestions(index)"
                    :key="p"
                    :label="p"
                    :value="p"
                  />
                </el-select>
              </el-col>
              <el-col :span="16">
                <el-input v-model="t.payloadValue" placeholder="参数值" @change="syncStructuredToPayload(t)" />
              </el-col>
            </el-row>
            <!-- CALIBRATE / fallback: raw JSON -->
            <el-input
              v-else
              v-model="t.payloadRaw"
              type="textarea"
              :rows="3"
              placeholder='{"param":"value"}'
            />
          </el-form-item>
        </div>
        <el-button type="primary" plain @click="addTarget" style="width: 100%">
          + 添加目标动作
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { deviceApi, policyApi, greenhouseApi, metricApi } from '@/api'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { ControlPolicy, ActuatorChannel, Greenhouse, GrowingZone, MetricDefinition } from '@/types'
import { populateMetricNames } from '@/utils/format'

// ── 执行器类型中文化 ──
const actuatorTypeLabels: Record<string, string> = {
  PUMP: '水泵', AERATOR: '曝气器', FAN: '风扇', VALVE: '阀门',
  SHADE: '遮阳帘', LED: '补光灯', HEATER: '加热器', CO2_GEN: 'CO2发生器',
  FOGGER: '雾化器', DOSING_PUMP: '计量泵', CHILLER: '冷水机', STIRRER: '搅拌器',
  DEHUMIDIFIER: '除湿机', DAMPER: '风阀', UV_STERILIZER: '紫外消毒器',
  OZONE_GENERATOR: '臭氧发生器', FILTER: '过滤器', RO_SYSTEM: '反渗透系统',
  TOP_UP_VALVE: '补水阀', ALARM: '报警器', CALIBRATION_VALVE: '校准阀'
}

const metricLabelMap: Record<string, string> = {
  TEMP: '温度', HUMIDITY: '湿度', CO2: 'CO2', LIGHT: '光照',
  PH: 'pH', EC: 'EC', DO: '溶氧', LEVEL: '液位',
  FLOW_RATE: '流量', TDS: 'TDS', TURBIDITY: '浊度', O3: '臭氧'
}

const actuatorMetricMap: Record<string, { code: string; unit: string }[]> = {
  FAN:       [{ code: 'TEMP', unit: '°C' }, { code: 'HUMIDITY', unit: '%RH' }, { code: 'CO2', unit: 'ppm' }],
  HEATER:    [{ code: 'TEMP', unit: '°C' }],
  CHILLER:   [{ code: 'TEMP', unit: '°C' }],
  DEHUMIDIFIER: [{ code: 'HUMIDITY', unit: '%RH' }],
  FOGGER:    [{ code: 'HUMIDITY', unit: '%RH' }, { code: 'TEMP', unit: '°C' }],
  LED:       [{ code: 'LIGHT', unit: 'lux' }],
  SHADE:     [{ code: 'LIGHT', unit: 'lux' }, { code: 'TEMP', unit: '°C' }],
  CO2_GEN:   [{ code: 'CO2', unit: 'ppm' }],
  DOSING_PUMP: [{ code: 'PH', unit: '' }, { code: 'EC', unit: 'mS/cm' }],
  PUMP:      [{ code: 'EC', unit: 'mS/cm' }, { code: 'PH', unit: '' }, { code: 'DO', unit: 'mg/L' }, { code: 'LEVEL', unit: 'cm' }],
  AERATOR:   [{ code: 'DO', unit: 'mg/L' }],
  VALVE:     [{ code: 'LEVEL', unit: 'cm' }, { code: 'FLOW_RATE', unit: 'L/min' }],
  TOP_UP_VALVE: [{ code: 'LEVEL', unit: 'cm' }],
  DAMPER:    [{ code: 'TEMP', unit: '°C' }, { code: 'HUMIDITY', unit: '%RH' }, { code: 'CO2', unit: 'ppm' }],
  STIRRER:   [{ code: 'EC', unit: 'mS/cm' }, { code: 'PH', unit: '' }, { code: 'TDS', unit: 'ppm' }],
  UV_STERILIZER: [{ code: 'TURBIDITY', unit: 'NTU' }],
  OZONE_GENERATOR: [{ code: 'O3', unit: 'ppm' }],
  FILTER:    [{ code: 'TURBIDITY', unit: 'NTU' }, { code: 'TDS', unit: 'ppm' }],
  RO_SYSTEM: [{ code: 'TDS', unit: 'ppm' }, { code: 'EC', unit: 'mS/cm' }]
}

// 执行器类型 → SET_VALUE 参数名建议
const actuatorParamSuggestions: Record<string, string[]> = {
  FAN:            ['speed', 'state'],
  HEATER:         ['target_temp', 'state'],
  CHILLER:        ['target_temp', 'state'],
  DEHUMIDIFIER:   ['target_humidity', 'state'],
  FOGGER:         ['interval_sec', 'duration_sec', 'state'],
  LED:            ['brightness', 'state', 'spectrum'],
  SHADE:          ['position', 'state'],
  CO2_GEN:        ['target_co2', 'state'],
  DOSING_PUMP:    ['ph_target', 'ec_target', 'dose_ml', 'state'],
  PUMP:           ['flow_rate', 'state'],
  AERATOR:        ['power', 'state'],
  VALVE:          ['position', 'state'],
  TOP_UP_VALVE:   ['state'],
  DAMPER:         ['position', 'state'],
  STIRRER:        ['speed', 'state'],
  UV_STERILIZER:  ['duration_sec', 'state'],
  OZONE_GENERATOR: ['target_o3', 'state'],
  FILTER:         ['state'],
  RO_SYSTEM:      ['state'],
  ALARM:          ['message', 'severity'],
  CALIBRATION_VALVE: ['state']
}

const weekdayOptions = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 7, label: '周日' }
]

function getParamSuggestions(index: number): string[] {
  const t = targets.value[index]
  if (!t?.channelId) return []
  const ch = getSelectedActuator(t.channelId)
  if (!ch) return []
  return actuatorParamSuggestions[ch.actuator_type] || []
}

// ── 多目标动作 ──
interface TargetItem {
  channelId: number | undefined
  commandType: string
  payloadRaw: string
  executionOrder: number
  // Structured payload helpers (synced with payloadRaw)
  switchState: string
  payloadKey: string
  payloadValue: string
}

function defaultTarget(): TargetItem {
  return {
    channelId: undefined,
    commandType: 'SWITCH',
    payloadRaw: '{"state":"ON"}',
    executionOrder: 1,
    switchState: 'ON',
    payloadKey: '',
    payloadValue: ''
  }
}

// ── 多触发条件 ──
interface ConditionItem {
  metric_code: string
  operator: string
  threshold_value: number
  hysteresis?: number
  window_sec?: number
  required_duration_sec?: number
  aggregation?: string
}

function defaultCondition(): ConditionItem {
  return {
    metric_code: 'TEMP',
    operator: '>',
    threshold_value: 30,
    hysteresis: undefined,
    window_sec: undefined,
    required_duration_sec: undefined,
    aggregation: undefined
  }
}

// Sync structured fields → payloadRaw based on commandType
function syncStructuredToPayload(t: TargetItem) {
  if (t.commandType === 'SWITCH') {
    t.payloadRaw = JSON.stringify({ state: t.switchState || 'ON' })
  } else if (t.commandType === 'SET_VALUE') {
    if (t.payloadKey) {
      // Try to parse payloadValue as number if possible
      const num = Number(t.payloadValue)
      const val: string | number = t.payloadValue !== '' && !isNaN(num) ? num : t.payloadValue
      t.payloadRaw = JSON.stringify({ [t.payloadKey]: val })
    } else {
      t.payloadRaw = '{}'
    }
  }
}

// Safely parse payloadRaw to a plain object, handling double-encoded strings
function safeParsePayload(raw: string): Record<string, unknown> {
  try {
    const parsed = JSON.parse(raw || '{}')
    if (typeof parsed === 'string') return JSON.parse(parsed)
    return typeof parsed === 'object' && parsed !== null ? parsed : {}
  } catch { return {} }
}

// Parse payloadRaw into structured fields (used when loading existing targets)
function parseTargetFromPayload(raw: string, commandType: string): Pick<TargetItem, 'payloadRaw' | 'switchState' | 'payloadKey' | 'payloadValue'> {
  const obj = safeParsePayload(raw)
  if (commandType === 'SWITCH') {
    return {
      payloadRaw: raw,
      switchState: (obj.state as string) || 'ON',
      payloadKey: '',
      payloadValue: ''
    }
  } else if (commandType === 'SET_VALUE') {
    const keys = Object.keys(obj)
    return {
      payloadRaw: raw,
      switchState: 'ON',
      payloadKey: keys[0] || '',
      payloadValue: keys[0] ? String(obj[keys[0]]) : ''
    }
  }
  return { payloadRaw: raw, switchState: 'ON', payloadKey: '', payloadValue: '' }
}

const loading = ref(false)
const submitLoading = ref(false)
const policies = ref<ControlPolicy[]>([])
const total = ref(0)
const pagination = reactive({ page: 1, pageSize: 20 })
const actuatorChannels = ref<ActuatorChannel[]>([])
const greenhouses = ref<Greenhouse[]>([])
const growingZones = ref<GrowingZone[]>([])
const metrics = ref<MetricDefinition[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingPolicyId = ref<number | null>(null)

const formData = reactive({
  policy_code: '',
  name: '',
  policy_type: 'THRESHOLD' as string,
  greenhouse_id: null as number | null,
  growing_zone_id: undefined as number | undefined,
  priority: 50,
  retry_limit: 3,
  timeout_sec: 30,
  effective_from: null as Date | null,
  effective_to: null as Date | null,
  schedule_mode: 'ONCE' as 'ONCE' | 'DAILY' | 'WEEKLY',
  run_once_at: null as Date | null,
  time_of_day: '' as string,
  weekdays: [] as number[],
  timezone: 'Asia/Shanghai'
})

const conditions = ref<ConditionItem[]>([defaultCondition()])

const targets = ref<TargetItem[]>([defaultTarget()])

// 动态标题
const dialogTitle = computed(() => {
  const tLabel = typeLabel(formData.policy_type)
  return isEdit.value ? `编辑策略 - ${tLabel}` : `新增策略 - ${tLabel}`
})

function typeLabel(t: string) {
  const map: Record<string, string> = { THRESHOLD: '阈值', SCHEDULE: '定时', DURATION: '时长' }
  return map[t] || t
}

function isPolicyEnabled(policy: ControlPolicy) {
  return policy.enabled === true || policy.enabled === 1
}

function isPolicyPublished(policy: ControlPolicy) {
  return Boolean(policy.published_at)
}

function encodeWeekdaysMask(weekdays: number[]): number | undefined {
  if (!weekdays.length) return undefined
  return weekdays.reduce((mask, day) => {
    if (day === 7) return mask | (1 << 6)
    return mask | (1 << (day - 1))
  }, 0)
}

function decodeWeekdaysMask(mask?: number): number[] {
  if (!mask) return []
  return weekdayOptions
    .map(item => item.value)
    .filter(day => (day === 7 ? (mask & (1 << 6)) !== 0 : (mask & (1 << (day - 1))) !== 0))
}

function formatScheduleSummary(policy: ControlPolicy): string {
  if (policy.policy_type !== 'SCHEDULE') return '-'
  if (!policy.schedule_mode) return '计划未配置'
  if (policy.schedule_mode === 'ONCE') {
    return policy.run_once_at ? `${formatRunOnceAt(policy.run_once_at, policy.timezone)} 单次` : '单次时间未配置'
  }
  if (policy.schedule_mode === 'DAILY') {
    return policy.time_of_day ? `每日 ${policy.time_of_day}` : '每日时刻未配置'
  }
  if (policy.schedule_mode === 'WEEKLY') {
    const weekdays = decodeWeekdaysMask(policy.weekdays_mask)
      .map(day => weekdayOptions.find(item => item.value === day)?.label)
      .filter(Boolean)
      .join('/')
    return weekdays && policy.time_of_day ? `${weekdays} ${policy.time_of_day}` : '每周计划未配置'
  }
  return '计划未配置'
}

function formatRunOnceAt(value: string, timezone?: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const formatter = new Intl.DateTimeFormat('zh-CN', {
    timeZone: timezone || 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
  return formatter.format(date).replace(/\//g, '-')
}

// ── 多目标动作 helpers ──
function addTarget() {
  const t = defaultTarget()
  t.executionOrder = targets.value.length + 1
  targets.value.push(t)
}

function removeTarget(index: number) {
  if (targets.value.length <= 1) return
  targets.value.splice(index, 1)
}

function addCondition() {
  conditions.value.push(defaultCondition())
}

function removeCondition(index: number) {
  if (conditions.value.length <= 1) return
  conditions.value.splice(index, 1)
}

function getSelectedActuator(channelId: number): ActuatorChannel | undefined {
  return actuatorChannels.value.find(ch => ch.id === channelId)
}

function getMetricHints(channelId: number): { code: string; unit: string }[] {
  const ch = getSelectedActuator(channelId)
  if (!ch) return []
  return actuatorMetricMap[ch.actuator_type] || []
}

function onTargetChannelChange(_index: number, _val: number | undefined) {
  // v-model already updates the target, no extra action needed
}

function onCommandTypeChange(t: TargetItem) {
  // Reset structured fields to defaults for the new command type, then sync to payload
  if (t.commandType === 'SWITCH') {
    t.switchState = 'ON'
    t.payloadKey = ''
    t.payloadValue = ''
    syncStructuredToPayload(t)
  } else if (t.commandType === 'SET_VALUE') {
    t.switchState = 'ON'
    t.payloadKey = ''
    t.payloadValue = ''
    t.payloadRaw = '{}'
  }
  // For CALIBRATE, keep existing payloadRaw as-is
}


async function fetchData() {
  loading.value = true
  try {
    const data = await policyApi.getPolicies({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    policies.value = data.items
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function loadActuatorChannels(greenhouseId?: number, growingZoneId?: number) {
  try {
    const params: Record<string, unknown> = { page_size: LARGE_PAGE_SIZE }
    if (greenhouseId) params.greenhouse_id = greenhouseId
    if (growingZoneId) params.growing_zone_id = growingZoneId
    const data = await deviceApi.getActuatorChannels(params)
    actuatorChannels.value = data.items
  } catch {
    actuatorChannels.value = []
  }
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { /* ignore */ }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    metrics.value = data.items
    populateMetricNames(data.items)
  } catch { /* ignore */ }
}

async function loadGrowingZones(greenhouseId?: number) {
  if (!greenhouseId) return
  try {
    const data = await greenhouseApi.getGrowingZones({ greenhouse_id: greenhouseId, page_size: LARGE_PAGE_SIZE })
    growingZones.value = data.items
  } catch {
    growingZones.value = []
  }
}

function disabledPastDate(_date: Date): boolean {
  return false
}

function disabledInvalidToDate(date: Date): boolean {
  if (date.getTime() < Date.now() - 60 * 1000) return true
  if (formData.effective_from) {
    return date.getTime() <= formData.effective_from.getTime()
  }
  return false
}

function onGreenhouseChange(greenhouseId: number | null) {
  formData.growing_zone_id = undefined
  growingZones.value = []
  if (greenhouseId) {
    loadGrowingZones(greenhouseId)
    loadActuatorChannels(greenhouseId)
  } else {
    actuatorChannels.value = []
  }
}

function onGrowingZoneChange(growingZoneId: number | undefined) {
  if (growingZoneId && formData.greenhouse_id) {
    loadActuatorChannels(formData.greenhouse_id, growingZoneId)
  } else if (formData.greenhouse_id) {
    loadActuatorChannels(formData.greenhouse_id)
  } else {
    actuatorChannels.value = []
  }
}

function openCreateDialog() {
  isEdit.value = false
  editingPolicyId.value = null
  formData.policy_code = `POL-${Date.now().toString().slice(-6)}`
  formData.name = ''
  formData.policy_type = 'THRESHOLD'
  formData.greenhouse_id = null
  formData.growing_zone_id = undefined
  formData.priority = 50
  formData.retry_limit = 3
  formData.timeout_sec = 30
  formData.effective_from = null
  formData.effective_to = null
  formData.schedule_mode = 'ONCE'
  formData.run_once_at = null
  formData.time_of_day = ''
  formData.weekdays = []
  formData.timezone = 'Asia/Shanghai'
  conditions.value = [defaultCondition()]
  targets.value = [defaultTarget()]
  growingZones.value = []
  actuatorChannels.value = []
  loadActuatorChannels()
  dialogVisible.value = true
}

watch(() => formData.policy_type, (val) => {
  if (val !== 'SCHEDULE') {
    formData.schedule_mode = 'ONCE'
    formData.run_once_at = null
    formData.time_of_day = ''
    formData.weekdays = []
    formData.timezone = 'Asia/Shanghai'
  }
})

async function openEditDialog(policy: ControlPolicy) {
  isEdit.value = true
  editingPolicyId.value = policy.id
  formData.policy_code = policy.policy_code
  formData.name = policy.name
  formData.policy_type = policy.policy_type
  formData.greenhouse_id = policy.greenhouse_id
  formData.growing_zone_id = policy.growing_zone_id
  formData.priority = policy.priority || 50
  formData.retry_limit = policy.retry_limit || 3
  formData.timeout_sec = policy.timeout_sec || 30
  formData.effective_from = policy.effective_from ? new Date(policy.effective_from) : null
  formData.effective_to = policy.effective_to ? new Date(policy.effective_to) : null
  formData.schedule_mode = policy.schedule_mode || 'ONCE'
  formData.run_once_at = policy.run_once_at ? new Date(policy.run_once_at) : null
  formData.time_of_day = policy.time_of_day || ''
  formData.weekdays = decodeWeekdaysMask(policy.weekdays_mask)
  formData.timezone = policy.timezone || 'Asia/Shanghai'

  // Load conditions only for threshold policies.
  if (policy.policy_type === 'THRESHOLD') {
    try {
      const condResult = await policyApi.getPolicyConditions(policy.id)
      if (condResult.items && condResult.items.length > 0) {
        conditions.value = condResult.items.map((c: any) => ({
          metric_code: c.metric_code,
          operator: c.operator,
          threshold_value: c.threshold_value,
          hysteresis: c.hysteresis,
          window_sec: c.window_sec,
          required_duration_sec: c.required_duration_sec,
          aggregation: c.aggregation
        }))
      } else {
        conditions.value = [defaultCondition()]
      }
    } catch {
      conditions.value = [defaultCondition()]
    }
  } else {
    conditions.value = [defaultCondition()]
  }

  // Load targets
  try {
    const tgtResult = await policyApi.getPolicyTargets(policy.id)
    if (tgtResult.items && tgtResult.items.length > 0) {
      targets.value = tgtResult.items.map(t => {
        const raw = typeof t.command_payload === 'string' ? t.command_payload : JSON.stringify(t.command_payload)
        return {
          channelId: t.actuator_channel_id,
          commandType: t.command_type,
          executionOrder: t.execution_order,
          ...parseTargetFromPayload(raw, t.command_type)
        }
      })
    } else {
      targets.value = [defaultTarget()]
    }
  } catch {
    targets.value = [defaultTarget()]
  }

  if (policy.greenhouse_id) {
    loadGrowingZones(policy.greenhouse_id)
    loadActuatorChannels(policy.greenhouse_id, policy.growing_zone_id)
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const now = Date.now()

  if (formData.effective_from && formData.effective_from.getTime() < now) {
    ElMessage.warning('生效起始时间不能早于当前时间')
    return
  }
  if (formData.effective_to) {
    if (formData.effective_to.getTime() < now) {
      ElMessage.warning('失效时间不能早于当前时间')
      return
    }
    if (formData.effective_from && formData.effective_to.getTime() <= formData.effective_from.getTime()) {
      ElMessage.warning('失效时间必须晚于生效起始时间')
      return
    }
  }

  if (formData.policy_type === 'SCHEDULE') {
    if (!formData.schedule_mode) {
      ElMessage.warning('请选择计划类型')
      return
    }
    if (formData.schedule_mode === 'ONCE' && !formData.run_once_at) {
      ElMessage.warning('请设置单次执行时间')
      return
    }
    if ((formData.schedule_mode === 'DAILY' || formData.schedule_mode === 'WEEKLY') && !formData.time_of_day) {
      ElMessage.warning('请设置执行时刻')
      return
    }
    if (formData.schedule_mode === 'WEEKLY' && !formData.weekdays.length) {
      ElMessage.warning('请至少选择一个执行星期')
      return
    }
  }

  submitLoading.value = true
  try {
    const useCond = formData.policy_type === 'THRESHOLD'
    const finalConditions = useCond ? conditions.value : []
    const finalTargets = targets.value.filter(t => t.channelId)

    if (isEdit.value && editingPolicyId.value) {
      const pid = editingPolicyId.value
      const payload: any = {
        name: formData.name,
        policy_type: formData.policy_type,
        growing_zone_id: formData.growing_zone_id,
        priority: formData.priority,
        retry_limit: formData.retry_limit,
        timeout_sec: formData.timeout_sec,
        effective_from: formData.effective_from?.toISOString() || null,
        effective_to: formData.effective_to?.toISOString() || null,
        schedule_mode: formData.policy_type === 'SCHEDULE' ? formData.schedule_mode : undefined,
        run_once_at: formData.policy_type === 'SCHEDULE' && formData.schedule_mode === 'ONCE' ? formData.run_once_at?.toISOString() || null : undefined,
        time_of_day: formData.policy_type === 'SCHEDULE' && formData.schedule_mode !== 'ONCE' ? formData.time_of_day : undefined,
        weekdays_mask: formData.policy_type === 'SCHEDULE' && formData.schedule_mode === 'WEEKLY' ? encodeWeekdaysMask(formData.weekdays) : undefined,
        timezone: formData.policy_type === 'SCHEDULE' ? formData.timezone : undefined
      }
      // Remove null fields to avoid overwriting with null
      if (payload.effective_from === null) delete payload.effective_from
      if (payload.effective_to === null) delete payload.effective_to
      if (payload.run_once_at === null) delete payload.run_once_at
      await policyApi.updatePolicy(pid, payload as any)

      // Sync Conditions
      try {
        const existingConds = await policyApi.getPolicyConditions(pid).then(res => res.items || [])
        for (const ec of existingConds) {
          await policyApi.deletePolicyCondition(pid, ec.id)
        }
      } catch { /* ignore */ }
      for (const c of finalConditions) {
        const cond = { ...c }
        if (!cond.aggregation) delete cond.aggregation
        await policyApi.createPolicyCondition(pid, cond as any)
      }

      // Sync Targets
      try {
        const existingTargets = await policyApi.getPolicyTargets(pid).then(res => res.items || [])
        for (const et of existingTargets) {
          await policyApi.deletePolicyTarget(pid, et.id)
        }
      } catch { /* ignore */ }
      for (let i = 0; i < finalTargets.length; i++) {
        const t = finalTargets[i]
        await policyApi.createPolicyTarget(pid, {
          actuator_channel_id: t.channelId!,
          command_type: t.commandType,
          command_payload: safeParsePayload(t.payloadRaw),
          execution_order: t.executionOrder ?? i
        })
      }

      ElMessage.success('策略更新成功，当前为未发布状态，请重新发布后生效')
    } else {
      // Create new policy
      const createPayload = {
        policy_code: formData.policy_code,
        name: formData.name,
        policy_type: formData.policy_type,
        greenhouse_id: formData.greenhouse_id!,
        growing_zone_id: formData.growing_zone_id,
        priority: formData.priority,
        retry_limit: formData.retry_limit,
        timeout_sec: formData.timeout_sec,
        effective_from: formData.effective_from?.toISOString() || undefined,
        effective_to: formData.effective_to?.toISOString() || undefined,
        schedule_mode: formData.policy_type === 'SCHEDULE' ? formData.schedule_mode : undefined,
        run_once_at: formData.policy_type === 'SCHEDULE' && formData.schedule_mode === 'ONCE' ? formData.run_once_at?.toISOString() || undefined : undefined,
        time_of_day: formData.policy_type === 'SCHEDULE' && formData.schedule_mode !== 'ONCE' ? formData.time_of_day : undefined,
        weekdays_mask: formData.policy_type === 'SCHEDULE' && formData.schedule_mode === 'WEEKLY' ? encodeWeekdaysMask(formData.weekdays) : undefined,
        timezone: formData.policy_type === 'SCHEDULE' ? formData.timezone : undefined
      }
      const res = await policyApi.createPolicy(createPayload as any)
      const pid = res.id

      // Create conditions
      for (const c of finalConditions) {
        const cond = { ...c }
        if (!cond.aggregation) delete cond.aggregation
        await policyApi.createPolicyCondition(pid, cond as any)
      }

      // Create targets
      for (let i = 0; i < finalTargets.length; i++) {
        const t = finalTargets[i]
        await policyApi.createPolicyTarget(pid, {
          actuator_channel_id: t.channelId!,
          command_type: t.commandType,
          command_payload: safeParsePayload(t.payloadRaw),
          execution_order: t.executionOrder ?? i
        })
      }

      ElMessage.success('策略创建成功，请发布后生效')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitLoading.value = false
  }
}

async function handlePublish(policy: ControlPolicy) {
  if (isPolicyPublished(policy)) return
  try {
    await policyApi.publishPolicy(policy.id)
    ElMessage.success('策略已发布')
    fetchData()
  } catch {
    // error handled
  }
}

async function handleDelete(policy: ControlPolicy) {
  await ElMessageBox.confirm(`确认删除策略「${policy.name}」？`, '提示', { type: 'warning' })
  await policyApi.deletePolicy(policy.id)
  ElMessage.success('策略已删除')
  fetchData()
}

onMounted(() => {
  fetchData()
  loadActuatorChannels()
  loadGreenhouses()
  loadMetrics()
})
</script>

<style scoped lang="scss">
.rules-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .page-title {
    font-size: 22px;
    font-weight: 700;
    margin: 0;
  }

  .table-container {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--spacing-md);
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }
}

.target-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 16px 4px;
  margin-bottom: 12px;
  background: var(--bg-subtle, #fafafa);

  .target-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    .target-item-title {
      font-weight: 600;
      font-size: 14px;
      color: var(--text-primary);
    }
  }
}

.metric-hints {
  margin-top: 8px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;

  .metric-hints-label {
    font-size: 12px;
    color: #909399;
  }

  .metric-hint-tag {
    margin: 0;
  }

  .metric-hints-none {
    font-size: 12px;
    color: #c0c4cc;
  }
}
</style>
