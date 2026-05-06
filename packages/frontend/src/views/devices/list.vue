<template>
    <div class="device-list-page">
      <div class="page-header">
        <h1 class="page-title">设备管理</h1>
        <div class="header-actions">
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>新增设备
          </el-button>
        </div>
      </div>

      <!-- 设备类型切换 -->
      <el-tabs v-model="activeDeviceType" @tab-change="onDeviceTypeChange">
        <el-tab-pane label="传感器设备" name="sensor" />
        <el-tab-pane label="执行器设备" name="actuator" />
      </el-tabs>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-select v-model="filters.status" placeholder="在线状态" clearable style="width: 120px">
          <el-option label="在线" value="ONLINE" />
          <el-option label="离线" value="OFFLINE" />
          <el-option label="故障" value="FAULT" />
        </el-select>
        <el-select v-model="filters.greenhouse_id" placeholder="所属温室" clearable style="width: 150px">
          <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
        </el-select>
        <el-select v-model="filters.zone_id" placeholder="种植区" clearable style="width: 150px">
          <el-option v-for="zone in growingZones" :key="zone.id" :label="zone.name" :value="zone.id" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索设备编码/名称" clearable style="width: 200px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-container">
        <el-table :data="devices" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="device_code" label="设备编码" width="130" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column v-if="activeDeviceType === 'actuator'" label="执行器通道数" width="120">
            <template #default="{ row }">{{ channelCountMap[row.id] ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="所属温室" width="120">
            <template #default="{ row }">{{ getGreenhouseName(row.greenhouse_id) }}</template>
          </el-table-column>
          <el-table-column label="种植区" width="120">
            <template #default="{ row }">{{ getZoneName(row.growing_zone_id) }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ONLINE' ? 'success' : row.status === 'FAULT' ? 'warning' : 'danger'">
                {{ row.status === 'ONLINE' ? '在线' : row.status === 'FAULT' ? '故障' : '离线' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="protocol" label="协议" width="80" />
          <el-table-column prop="last_seen_at" label="最后上报" width="160">
            <template #default="{ row }">{{ formatDate(row.last_seen_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="goDetail(row.id)">详情</el-button>
              <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
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

      <!-- 新增/编辑弹窗 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑设备' : '新增设备'" width="550px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
          <el-form-item label="设备编码" prop="device_code">
            <el-input v-model="formData.device_code" :disabled="isEdit" placeholder="请输入设备编码" maxlength="64" autocomplete="off" name="device_code" />
          </el-form-item>
          <el-form-item label="设备名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入设备名称" maxlength="64" autocomplete="off" name="name" />
          </el-form-item>
          <el-form-item label="设备型号" prop="model">
            <el-input v-model="formData.model" placeholder="请输入型号（选填）" maxlength="64" autocomplete="off" name="model" />
          </el-form-item>
          <el-form-item label="固件版本" prop="firmware_version">
            <el-input v-model="formData.firmware_version" placeholder="请输入固件版本（选填）" maxlength="32" autocomplete="off" name="firmware_version" />
          </el-form-item>
          <el-form-item label="通信协议" prop="protocol">
            <el-select v-model="formData.protocol" placeholder="请选择通信协议" style="width: 100%">
              <el-option label="MQTT" value="MQTT" />
              <el-option label="HTTP" value="HTTP" />
            </el-select>
          </el-form-item>
          <el-form-item label="所属温室" prop="greenhouse_id">
            <el-select v-model="formData.greenhouse_id" placeholder="请选择温室" clearable style="width: 100%">
              <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="种植区" prop="growing_zone_id">
            <el-select v-model="formData.growing_zone_id" placeholder="请选择种植区" clearable style="width: 100%">
              <el-option v-for="zone in filteredGrowingZones" :key="zone.id" :label="zone.name" :value="zone.id" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { deviceApi, greenhouseApi } from '@/api'
import { formatDate } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { SensorDevice, ActuatorDevice, Greenhouse, GrowingZone } from '@/types'

const router = useRouter()

type AnyDevice = SensorDevice | ActuatorDevice

const activeDeviceType = ref<'sensor' | 'actuator'>('sensor')
const devices = ref<AnyDevice[]>([])
const total = ref(0)
const loading = ref(false)
const greenhouses = ref<Greenhouse[]>([])
const growingZones = ref<GrowingZone[]>([])
const channelCountMap = ref<Record<number, number>>({})

const filters = reactive({
  status: '' as string,
  greenhouse_id: null as number | null,
  zone_id: null as number | null,
  keyword: '' as string
})

const pagination = reactive({ page: 1, pageSize: 20 })

// 创建设备弹窗
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive({
  device_code: '',
  name: '',
  model: '',
  firmware_version: '',
  protocol: 'MQTT',
  greenhouse_id: null as number | null,
  growing_zone_id: undefined as number | undefined
})

const formRules: FormRules = {
  device_code: [
    { required: true, message: '请输入设备编码', trigger: 'blur' },
    { min: 1, max: 64, message: '设备编码长度为 1-64 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入设备名称', trigger: 'blur' },
    { min: 1, max: 64, message: '设备名称长度为 1-64 个字符', trigger: 'blur' }
  ],
  greenhouse_id: [{ required: true, message: '请选择温室', trigger: 'change' }]
}

const filteredGrowingZones = computed(() => {
  if (!formData.greenhouse_id) return growingZones.value
  return growingZones.value.filter(z => z.greenhouse_id === formData.greenhouse_id)
})

function getGreenhouseName(id: number): string {
  const gh = greenhouses.value.find(g => g.id === id)
  return gh?.name || '-'
}

function getZoneName(id: number | undefined | null): string {
  if (!id) return '-'
  const zone = growingZones.value.find(z => z.id === id)
  return zone?.name || '-'
}

function goDetail(id: number) {
  router.push({ path: `/devices/${id}`, query: { type: activeDeviceType.value } })
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.status) params.status = filters.status
    if (filters.greenhouse_id) params.greenhouse_id = filters.greenhouse_id
    if (filters.zone_id) params.growing_zone_id = filters.zone_id
    if (filters.keyword) params.keyword = filters.keyword

    if (activeDeviceType.value === 'sensor') {
      const data = await deviceApi.getSensorDevices(params)
      devices.value = data.items
      total.value = data.total
    } else {
      const data = await deviceApi.getActuatorDevices(params)
      devices.value = data.items
      total.value = data.total
      // Load channel counts for actuator devices
      loadActuatorChannelCounts(data.items as ActuatorDevice[])
    }
  } catch {
    // error handled
  } finally {
    loading.value = false
  }
}

async function loadActuatorChannelCounts(actuatorDevices: ActuatorDevice[]) {
  const map: Record<number, number> = {}
  for (const dev of actuatorDevices) {
    try {
      const chData = await deviceApi.getActuatorChannels({
        actuator_device_id: dev.id,
        page_size: 1,
        page: 1
      })
      map[dev.id] = chData.total
    } catch {
      map[dev.id] = 0
    }
  }
  channelCountMap.value = map
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { /* ignore */ }
}

async function loadGrowingZones() {
  try {
    const data = await greenhouseApi.getGrowingZones({ page_size: LARGE_PAGE_SIZE })
    growingZones.value = data.items
  } catch { /* ignore */ }
}

function onDeviceTypeChange() {
  pagination.page = 1
  fetchData()
}

function resetFilters() {
  filters.status = ''
  filters.greenhouse_id = null
  filters.zone_id = null
  filters.keyword = ''
  pagination.page = 1
  fetchData()
}

function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  formData.device_code = ''
  formData.name = ''
  formData.model = ''
  formData.firmware_version = ''
  formData.protocol = 'MQTT'
  formData.greenhouse_id = null
  formData.growing_zone_id = undefined
  dialogVisible.value = true
  formRef.value?.resetFields()
}

function openEditDialog(device: AnyDevice) {
  isEdit.value = true
  editingId.value = device.id
  formData.device_code = device.device_code
  formData.name = device.name
  formData.model = device.model || ''
  formData.firmware_version = device.firmware_version || ''
  formData.protocol = device.protocol
  formData.greenhouse_id = device.greenhouse_id
  formData.growing_zone_id = device.growing_zone_id
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try { await formRef.value.validate() } catch { return }

  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      const payload = {
        name: formData.name,
        model: formData.model || undefined,
        firmware_version: formData.firmware_version || undefined,
        greenhouse_id: formData.greenhouse_id ?? undefined,
        growing_zone_id: formData.growing_zone_id
      }
      if (activeDeviceType.value === 'sensor') {
        await deviceApi.updateSensorDevice(editingId.value, payload)
      } else {
        await deviceApi.updateActuatorDevice(editingId.value, payload)
      }
      ElMessage.success('设备更新成功')
    } else {
      const payload = {
        device_code: formData.device_code,
        name: formData.name,
        model: formData.model || undefined,
        firmware_version: formData.firmware_version || undefined,
        protocol: formData.protocol,
        greenhouse_id: formData.greenhouse_id!,
        growing_zone_id: formData.growing_zone_id
      }
      if (activeDeviceType.value === 'sensor') {
        await deviceApi.createSensorDevice(payload)
      } else {
        await deviceApi.createActuatorDevice(payload)
      }
      ElMessage.success('设备创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch { /* error handled */ }
  finally { submitLoading.value = false }
}

async function handleDelete(device: AnyDevice) {
  try {
    await ElMessageBox.confirm(`确认删除设备「${device.name}」？此操作不可撤销。`, '警告', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    if (activeDeviceType.value === 'sensor') {
      await deviceApi.deleteSensorDevice(device.id)
    } else {
      await deviceApi.deleteActuatorDevice(device.id)
    }
    ElMessage.success('设备已删除')
    fetchData()
  } catch (e: any) {
    if (e !== 'cancel') { /* error handled */ }
  }
}

onMounted(() => {
  fetchData()
  loadGreenhouses()
  loadGrowingZones()
})
</script>

<style scoped lang="scss">
.device-list-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .page-title {
    font-size: 22px;
    font-weight: 700;
    color: var(--color-text-primary);
    margin: 0;
    text-wrap: balance;
  }

  .header-actions {
    display: flex;
    align-items: center;
  }

  .filter-section {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
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

  .ml-sm {
    margin-left: 8px;
    color: var(--color-text-secondary);
  }
}
</style>
