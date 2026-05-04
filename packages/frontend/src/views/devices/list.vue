<template>
  <AppLayout>
    <div class="device-list-page">
      <div class="page-header">
        <h1 class="page-title">设备管理</h1>
        <el-button v-if="canEdit" type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增设备
        </el-button>
      </div>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-select v-model="filters.type" placeholder="设备类型" clearable style="width: 120px">
          <el-option label="传感器" value="SENSOR" />
          <el-option label="执行器" value="ACTUATOR" />
        </el-select>
        <el-select v-model="filters.category" placeholder="设备分类" clearable style="width: 120px">
          <el-option v-for="cat in categoryOptions" :key="cat.value" :label="cat.label" :value="cat.value" />
        </el-select>
        <el-select v-model="filters.greenhouse_id" placeholder="所属温室" clearable style="width: 150px">
          <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
        </el-select>
        <el-select v-model="filters.group_id" placeholder="所属分组" clearable style="width: 150px">
          <el-option v-for="group in deviceGroups" :key="group.id" :label="group.name" :value="group.id" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100px">
          <el-option label="启用" value="ENABLED" />
          <el-option label="禁用" value="DISABLED" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索设备编码/名称" clearable style="width: 200px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-container">
        <el-table :data="devices" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="device_code" label="设备编码" width="120" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              {{ row.type === 'SENSOR' ? '传感器' : '执行器' }}
            </template>
          </el-table-column>
          <el-table-column prop="category" label="分类" width="100">
            <template #default="{ row }">
              {{ getCategoryName(row.category) }}
            </template>
          </el-table-column>
          <el-table-column prop="greenhouse_id" label="温室" width="100">
            <template #default="{ row }">
              {{ getGreenhouseName(row.greenhouse_id) }}
            </template>
          </el-table-column>
          <el-table-column prop="group_id" label="分组" width="100">
            <template #default="{ row }">
              {{ getGroupName(row.group_id) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ENABLED' ? 'success' : 'danger'">
                {{ row.status === 'ENABLED' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="last_seen_at" label="在线状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.last_seen_at ? 'success' : 'danger'">
                {{ row.last_seen_at ? '在线' : '离线' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="goDetail(row.id)">详情</el-button>
              <el-button v-if="canEdit" type="primary" link @click="openEditDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
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
          <el-form-item label="设备类型" prop="type">
            <el-select v-model="formData.type" placeholder="请选择设备类型" style="width: 100%">
              <el-option label="传感器" value="SENSOR" />
              <el-option label="执行器" value="ACTUATOR" />
            </el-select>
          </el-form-item>
          <el-form-item label="设备分类" prop="category">
            <el-select v-model="formData.category" placeholder="请选择设备分类" style="width: 100%">
              <el-option v-for="cat in categoryOptions" :key="cat.value" :label="cat.label" :value="cat.value" />
            </el-select>
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
          <el-form-item label="所属分组" prop="group_id">
            <el-select v-model="formData.group_id" placeholder="请选择分组" clearable style="width: 100%">
              <el-option v-for="group in filteredDeviceGroups" :key="group.id" :label="group.name" :value="group.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="采样间隔" prop="sampling_interval_sec">
            <el-input-number v-model="formData.sampling_interval_sec" :min="5" :max="3600" />
            <span class="ml-sm">秒</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { AppLayout } from '@/components/layout'
import { useDeviceStore } from '@/stores/device'
import { useGreenhouseStore } from '@/stores/greenhouse'
import { usePermission } from '@/composables/usePermission'
import { getCategoryName } from '@/utils/format'
import { Device, DeviceFormData, DeviceGroup, DeviceType, DeviceProtocol, DeviceStatus } from '@/types'

const router = useRouter()
const deviceStore = useDeviceStore()
const greenhouseStore = useGreenhouseStore()
const { canEditDevice } = usePermission()

const canEdit = computed(() => canEditDevice())

// 数据
const devices = computed(() => deviceStore.devices)
const deviceGroups = computed(() => deviceStore.deviceGroups as DeviceGroup[])
const greenhouses = computed(() => greenhouseStore.greenhouses)
const total = computed(() => deviceStore.total)
const loading = computed(() => deviceStore.loading)

// 筛选条件
const filters = reactive({
  type: '',
  category: '',
  greenhouse_id: null as number | null,
  group_id: null as number | null,
  status: '',
  keyword: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 设备分类选项
const categoryOptions = [
  { label: '温度', value: 'TEMP' },
  { label: '湿度', value: 'HUMIDITY' },
  { label: 'pH值', value: 'PH' },
  { label: '电导率', value: 'EC' },
  { label: 'CO2', value: 'CO2' },
  { label: '光照', value: 'LIGHT' },
  { label: '风机', value: 'FAN' },
  { label: '水泵', value: 'PUMP' },
  { label: '阀门', value: 'VALVE' }
]

// 弹窗
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive<DeviceFormData>({
  device_code: '',
  name: '',
  type: DeviceType.SENSOR,
  category: 'TEMP',
  protocol: DeviceProtocol.MQTT,
  greenhouse_id: null,
  group_id: null,
  sampling_interval_sec: 60
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
  type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
  category: [{ required: true, message: '请选择设备分类', trigger: 'change' }],
  protocol: [{ required: true, message: '请选择通信协议', trigger: 'change' }]
}

// 根据选择的温室筛选分组
const filteredDeviceGroups = computed(() => {
  if (!formData.greenhouse_id) return deviceGroups.value
  return deviceGroups.value.filter(g => g.greenhouse_id === formData.greenhouse_id)
})

// 获取温室名称
function getGreenhouseName(greenhouseId: number | null): string {
  if (!greenhouseId) return '-'
  const gh = greenhouses.value.find(g => g.id === greenhouseId)
  return gh?.name || '-'
}

// 获取分组名称
function getGroupName(groupId: number | null): string {
  if (!groupId) return '-'
  const group = deviceGroups.value.find((g) => g.id === groupId)
  return group?.name || '-'
}

// 获取数据
async function fetchData() {
  try {
    await deviceStore.fetchDevices({
      page: pagination.page,
      page_size: pagination.pageSize,
      type: (filters.type as DeviceType) || undefined,
      category: filters.category || undefined,
      greenhouse_id: filters.greenhouse_id || undefined,
      group_id: filters.group_id || undefined,
      status: (filters.status as DeviceStatus) || undefined,
      keyword: filters.keyword || undefined
    })
  } catch (error) {
    console.error('[Devices] Failed to fetch device list:', error)
  }
}

// 重置筛选
function resetFilters() {
  filters.type = ''
  filters.category = ''
  filters.greenhouse_id = null
  filters.group_id = null
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  fetchData()
}

// 跳转详情
function goDetail(id: number) {
  router.push(`/devices/${id}`)
}

// 打开新增弹窗
function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    device_code: '',
    name: '',
    type: 'SENSOR',
    category: 'TEMP',
    protocol: 'MQTT',
    greenhouse_id: null,
    group_id: null,
    sampling_interval_sec: 60
  })
  dialogVisible.value = true
}

// 打开编辑弹窗
function openEditDialog(device: Device) {
  isEdit.value = true
  editingId.value = device.id
  Object.assign(formData, {
    device_code: device.device_code,
    name: device.name,
    type: device.type,
    category: device.category,
    protocol: device.protocol,
    greenhouse_id: device.greenhouse_id,
    group_id: device.group_id,
    sampling_interval_sec: device.sampling_interval_sec || 60
  })
  dialogVisible.value = true
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      await deviceStore.editDevice(editingId.value, {
        name: formData.name,
        category: formData.category,
        greenhouse_id: formData.greenhouse_id,
        group_id: formData.group_id,
        sampling_interval_sec: formData.sampling_interval_sec
      })
      ElMessage.success('设备更新成功')
    } else {
      await deviceStore.addDevice(formData)
      ElMessage.success('设备创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

// 监听温室变化，清空分组选择
watch(() => formData.greenhouse_id, () => {
  formData.group_id = null
})

onMounted(() => {
  fetchData()
  deviceStore.fetchDeviceGroups()
  greenhouseStore.fetchGreenhouses()
})
</script>

<style scoped lang="scss">
.device-list-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
  }

  .table-container {
    background: #fff;
    border-radius: 4px;
    padding: 16px;
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .ml-sm {
    margin-left: 8px;
    color: #909399;
  }
}
</style>
