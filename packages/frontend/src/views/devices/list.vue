<template>
    <div class="device-list-page">
      <div class="page-header">
        <h1 class="page-title">设备管理</h1>
        <div class="header-actions">
          <el-dropdown v-if="canEdit && selectedIds.length > 0" trigger="click" style="margin-right: 8px">
            <el-button type="warning">
              批量操作 ({{ selectedIds.length }})
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="openBatchStatusDialog('ENABLED')">批量启用</el-dropdown-item>
                <el-dropdown-item @click="openBatchStatusDialog('DISABLED')">批量禁用</el-dropdown-item>
                <el-dropdown-item @click="openBatchGroupDialog">批量修改分组</el-dropdown-item>
                <el-dropdown-item @click="openBatchIntervalDialog">批量修改采样间隔</el-dropdown-item>
                <el-dropdown-item divided @click="openBatchCommandDialog">批量下发命令</el-dropdown-item>
                <el-dropdown-item divided @click="openBatchDeleteDialog" style="color: #f56c6c">批量删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button v-if="canEdit" type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>新增设备
          </el-button>
        </div>
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
        <el-table :data="devices" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="50" v-if="canEdit" />
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="device_code" label="设备编码" width="120" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              {{ row.type === 'SENSOR' ? '传感器' : '执行器' }}
            </template>
          </el-table-column>
          <el-table-column prop="category" label="分类" width="100">
            <template #default="{ row }">{{ getCategoryName(row.category) }}</template>
          </el-table-column>
          <el-table-column prop="greenhouse_id" label="温室" width="100">
            <template #default="{ row }">{{ getGreenhouseName(row.greenhouse_id) }}</template>
          </el-table-column>
          <el-table-column prop="group_id" label="分组" width="100">
            <template #default="{ row }">{{ getGroupName(row.group_id) }}</template>
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

      <!-- 批量修改分组弹窗 -->
      <el-dialog v-model="batchGroupDialogVisible" title="批量修改分组" width="400px">
        <el-form label-width="80px">
          <el-form-item label="目标分组">
            <el-select v-model="batchGroupId" placeholder="请选择分组" clearable style="width: 100%">
              <el-option v-for="group in deviceGroups" :key="group.id" :label="group.name" :value="group.id" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="batchGroupDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="batchLoading" @click="handleBatchGroup">确定</el-button>
        </template>
      </el-dialog>

      <!-- 批量修改采样间隔弹窗 -->
      <el-dialog v-model="batchIntervalDialogVisible" title="批量修改采样间隔" width="400px">
        <el-form label-width="80px">
          <el-form-item label="采样间隔">
            <el-input-number v-model="batchInterval" :min="5" :max="3600" />
            <span class="ml-sm">秒</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="batchIntervalDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="batchLoading" @click="handleBatchInterval">确定</el-button>
        </template>
      </el-dialog>

      <!-- 批量下发命令弹窗 -->
      <el-dialog v-model="batchCommandDialogVisible" title="批量下发命令" width="500px">
        <el-form label-width="80px">
          <el-form-item label="命令类型">
            <el-input v-model="batchCommandType" placeholder="如: SWITCH, SET_VALUE" />
          </el-form-item>
          <el-form-item label="命令参数">
            <el-input v-model="batchCommandPayload" type="textarea" placeholder='JSON 格式，如: {"value": 25}' rows="4" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="batchCommandRemark" placeholder="选填" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="batchCommandDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="batchLoading" @click="handleBatchCommand">下发</el-button>
        </template>
      </el-dialog>

      <!-- 批量删除弹窗 -->
      <el-dialog v-model="batchDeleteDialogVisible" title="批量删除设备" width="450px">
        <p>确定要删除 {{ selectedIds.length }} 个设备吗？此操作不可撤销。</p>
        <el-form label-width="60px" style="margin-top: 12px">
          <el-form-item label="原因">
            <el-input v-model="batchDeleteReason" placeholder="请输入删除原因（选填）" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="batchDeleteDialogVisible = false">取消</el-button>
          <el-button type="danger" :loading="batchLoading" @click="handleBatchDelete">确认删除</el-button>
        </template>
      </el-dialog>

      <!-- 批量操作结果弹窗 -->
      <el-dialog v-model="batchResultDialogVisible" title="批量操作结果" width="500px">
        <el-table :data="batchResults" size="small" max-height="300">
          <el-table-column prop="device_id" label="设备 ID" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'FAILED' ? 'danger' : 'success'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="信息" show-overflow-tooltip />
        </el-table>
        <template #footer>
          <el-button @click="batchResultDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import { useDeviceStore } from '@/stores/device'
import { useGreenhouseStore } from '@/stores/greenhouse'
import { usePermission } from '@/composables/usePermission'
import { getCategoryName } from '@/utils/format'
import { batchUpdateDevices, batchDeleteDevices, batchCommands } from '@/api/device'
import { Device, DeviceFormData, DeviceGroup, DeviceType, DeviceProtocol, DeviceStatus } from '@/types'

const router = useRouter()
const deviceStore = useDeviceStore()
const greenhouseStore = useGreenhouseStore()
const { canEditDevice } = usePermission()

const canEdit = computed(() => canEditDevice())

const devices = computed(() => deviceStore.devices)
const deviceGroups = computed(() => deviceStore.deviceGroups as DeviceGroup[])
const greenhouses = computed(() => greenhouseStore.greenhouses)
const total = computed(() => deviceStore.total)
const loading = computed(() => deviceStore.loading)

// 多选
const selectedIds = ref<number[]>([])
function handleSelectionChange(rows: Device[]) {
  selectedIds.value = rows.map(r => r.id)
}

// 筛选
const filters = reactive({
  type: '', category: '', greenhouse_id: null as number | null,
  group_id: null as number | null, status: '', keyword: ''
})
const pagination = reactive({ page: 1, pageSize: 20 })

const categoryOptions = [
  { label: '温度', value: 'TEMP' }, { label: '湿度', value: 'HUMIDITY' },
  { label: 'pH值', value: 'PH' }, { label: '电导率', value: 'EC' },
  { label: 'CO2', value: 'CO2' }, { label: '光照', value: 'LIGHT' },
  { label: '风机', value: 'FAN' }, { label: '水泵', value: 'PUMP' }, { label: '阀门', value: 'VALVE' }
]

// 创建设备弹窗
const dialogVisible = ref(false), isEdit = ref(false), formRef = ref<FormInstance>()
const submitLoading = ref(false), editingId = ref<number | null>(null)
const formData = reactive<DeviceFormData>({
  device_code: '', name: '', type: DeviceType.SENSOR, category: 'TEMP',
  protocol: DeviceProtocol.MQTT, greenhouse_id: null, group_id: null, sampling_interval_sec: 60
})
const formRules: FormRules = {
  device_code: [{ required: true, message: '请输入设备编码', trigger: 'blur' }, { min: 1, max: 64, message: '设备编码长度为 1-64 个字符', trigger: 'blur' }],
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }, { min: 1, max: 64, message: '设备名称长度为 1-64 个字符', trigger: 'blur' }],
  type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
  category: [{ required: true, message: '请选择设备分类', trigger: 'change' }],
  protocol: [{ required: true, message: '请选择通信协议', trigger: 'change' }]
}

const filteredDeviceGroups = computed(() => {
  if (!formData.greenhouse_id) return deviceGroups.value
  return deviceGroups.value.filter(g => g.greenhouse_id === formData.greenhouse_id)
})

// 批量操作
const batchLoading = ref(false)
const batchResultDialogVisible = ref(false)
const batchResults = ref<any[]>([])

// 批量状态
function openBatchStatusDialog(status: string) {
  const action = status === 'ENABLED' ? '启用' : '禁用'
  ElMessageBox.confirm(`确定要批量${action} ${selectedIds.value.length} 个设备吗？`, '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(() => {
    batchLoading.value = true
    batchUpdateDevices(selectedIds.value, { status }).then((res) => {
      ElMessage.success(`成功更新 ${res.affected} 个设备`)
      selectedIds.value = []
      fetchData()
    }).catch(() => {}).finally(() => { batchLoading.value = false })
  }).catch(() => {})
}

// 批量分组
const batchGroupDialogVisible = ref(false), batchGroupId = ref<number | null>(null)
function openBatchGroupDialog() { batchGroupId.value = null; batchGroupDialogVisible.value = true }
async function handleBatchGroup() {
  if (batchGroupId.value === null) { ElMessage.warning('请选择分组'); return }
  batchLoading.value = true
  try {
    const res = await batchUpdateDevices(selectedIds.value, { group_id: batchGroupId.value })
    ElMessage.success(`成功更新 ${res.affected} 个设备`)
    batchGroupDialogVisible.value = false; selectedIds.value = []; fetchData()
  } catch {} finally { batchLoading.value = false }
}

// 批量间隔
const batchIntervalDialogVisible = ref(false), batchInterval = ref(60)
function openBatchIntervalDialog() { batchInterval.value = 60; batchIntervalDialogVisible.value = true }
async function handleBatchInterval() {
  if (batchInterval.value < 5 || batchInterval.value > 3600) { ElMessage.warning('采样间隔范围为 5-3600 秒'); return }
  batchLoading.value = true
  try {
    const res = await batchUpdateDevices(selectedIds.value, { sampling_interval_sec: batchInterval.value })
    ElMessage.success(`成功更新 ${res.affected} 个设备`)
    batchIntervalDialogVisible.value = false; selectedIds.value = []; fetchData()
  } catch {} finally { batchLoading.value = false }
}

// 批量命令
const batchCommandDialogVisible = ref(false), batchCommandType = ref(''), batchCommandPayload = ref(''), batchCommandRemark = ref('')
function openBatchCommandDialog() {
  batchCommandType.value = ''; batchCommandPayload.value = ''; batchCommandRemark.value = ''
  batchCommandDialogVisible.value = true
}
async function handleBatchCommand() {
  if (!batchCommandType.value.trim()) { ElMessage.warning('请输入命令类型'); return }
  let payload: any
  try { payload = JSON.parse(batchCommandPayload.value || '{}') }
  catch { ElMessage.warning('命令参数 JSON 格式不正确'); return }
  batchLoading.value = true
  try {
    const data = await batchCommands({
      target_type: 'devices', target_ids: selectedIds.value,
      command_type: batchCommandType.value, payload,
      remark: batchCommandRemark.value || undefined
    })
    batchResults.value = data.results.map(r => ({ device_id: r.device_id, status: r.status, message: r.message || r.command_id?.toString() || '' }))
    batchResultDialogVisible.value = true
    batchCommandDialogVisible.value = false
  } catch {} finally { batchLoading.value = false }
}

// 批量删除
const batchDeleteDialogVisible = ref(false), batchDeleteReason = ref('')
function openBatchDeleteDialog() {
  batchDeleteReason.value = ''; batchDeleteDialogVisible.value = true
}
async function handleBatchDelete() {
  batchLoading.value = true
  try {
    const res = await batchDeleteDevices(selectedIds.value, batchDeleteReason.value || undefined)
    ElMessage.success(`成功删除 ${res.deleted} 个设备`)
    batchDeleteDialogVisible.value = false; selectedIds.value = []; fetchData()
  } catch {} finally { batchLoading.value = false }
}

function getGreenhouseName(id: number | null): string {
  if (!id) return '-'; const gh = greenhouses.value.find(g => g.id === id); return gh?.name || '-'
}
function getGroupName(id: number | null): string {
  if (!id) return '-'; const g = deviceGroups.value.find(g => g.id === id); return g?.name || '-'
}

async function fetchData() {
  try {
    await deviceStore.fetchDevices({
      page: pagination.page, page_size: pagination.pageSize,
      type: (filters.type as DeviceType) || undefined, category: filters.category || undefined,
      greenhouse_id: filters.greenhouse_id || undefined, group_id: filters.group_id || undefined,
      status: (filters.status as DeviceStatus) || undefined, keyword: filters.keyword || undefined
    })
  } catch (e) { console.error(e) }
}

function resetFilters() {
  filters.type = ''; filters.category = ''; filters.greenhouse_id = null
  filters.group_id = null; filters.status = ''; filters.keyword = ''
  pagination.page = 1; fetchData()
}

function goDetail(id: number) { router.push(`/devices/${id}`) }

function openCreateDialog() {
  isEdit.value = false; editingId.value = null
  Object.assign(formData, { device_code: '', name: '', type: 'SENSOR', category: 'TEMP', protocol: 'MQTT', greenhouse_id: null, group_id: null, sampling_interval_sec: 60 })
  dialogVisible.value = true
}
function openEditDialog(device: Device) {
  isEdit.value = true; editingId.value = device.id
  Object.assign(formData, { device_code: device.device_code, name: device.name, type: device.type, category: device.category, protocol: device.protocol, greenhouse_id: device.greenhouse_id, group_id: device.group_id, sampling_interval_sec: device.sampling_interval_sec || 60 })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try { await formRef.value.validate() } catch { return }
  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      await deviceStore.editDevice(editingId.value, { name: formData.name, category: formData.category, greenhouse_id: formData.greenhouse_id, group_id: formData.group_id, sampling_interval_sec: formData.sampling_interval_sec })
      ElMessage.success('设备更新成功')
    } else {
      await deviceStore.addDevice(formData)
      ElMessage.success('设备创建成功')
    }
    dialogVisible.value = false; fetchData()
  } catch {} finally { submitLoading.value = false }
}

watch(() => formData.greenhouse_id, () => { formData.group_id = null })
onMounted(() => { fetchData(); deviceStore.fetchDeviceGroups(); greenhouseStore.fetchGreenhouses() })
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
