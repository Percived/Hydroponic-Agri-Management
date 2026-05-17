<template>
  <div class="greenhouse-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/assets/greenhouses')" text>
          <el-icon><ArrowLeft /></el-icon>返回温室管理
        </el-button>
        <h1 class="page-title">{{ greenhouse?.name || '温室详情' }}</h1>
        <el-tag v-if="greenhouse" size="large" :type="greenhouse.status === 'ENABLED' ? 'success' : 'info'">
          {{ greenhouse.status === 'ENABLED' ? '已启用' : '已停用' }}
        </el-tag>
      </div>
      <div class="header-right" v-if="greenhouse">
        <el-button v-if="canManage" type="primary" @click="openEditDialog">
          <el-icon><Edit /></el-icon>编辑
        </el-button>
        <el-button v-if="canDelete" type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>删除
        </el-button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading && !greenhouse" class="loading-container">
      <el-skeleton :rows="6" animated />
    </div>

    <!-- Error -->
    <el-result
      v-else-if="errorMsg"
      icon="error"
      :title="errorMsg"
      sub-title="请检查网络连接或返回列表重试"
    >
      <template #extra>
        <el-button type="primary" @click="loadData">重新加载</el-button>
        <el-button @click="router.push('/assets/greenhouses')">返回列表</el-button>
      </template>
    </el-result>

    <!-- Content -->
    <template v-else-if="greenhouse">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <el-card shadow="never" class="info-card">
            <template #header><span class="card-title">温室信息</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="温室名称">{{ greenhouse.name }}</el-descriptions-item>
              <el-descriptions-item label="温室编号">{{ greenhouse.code }}</el-descriptions-item>
              <el-descriptions-item label="位置">{{ greenhouse.location || '-' }}</el-descriptions-item>
              <el-descriptions-item label="面积 (m²)">{{ greenhouse.area_sqm ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="启用状态">
                <el-tag :type="greenhouse.status === 'ENABLED' ? 'success' : 'info'">
                  {{ greenhouse.status === 'ENABLED' ? '已启用' : '已停用' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="种植区数量">{{ greenhouse.zone_count ?? zones.length }}</el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">{{ greenhouse.description || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDateTime(greenhouse.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatDateTime(greenhouse.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-row :gutter="16" class="stats-row">
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ greenhouse.zone_count ?? zones.length }}</div>
                <div class="stat-label">种植区</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ greenhouse.area_sqm ?? '-' }}</div>
                <div class="stat-label">总面积 (m²)</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ greenhouse.status === 'ENABLED' ? '运行中' : '已停用' }}</div>
                <div class="stat-label">运行状态</div>
              </div>
            </el-col>
          </el-row>
        </el-tab-pane>

        <!-- Tab 2: 种植区 -->
        <el-tab-pane label="种植区" name="zones">
          <el-table :data="zones" v-loading="zonesLoading" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="code" label="编号" width="140" />
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column label="系统类型" width="100">
              <template #default="{ row }">{{ systemTypeName(row.system_type) }}</template>
            </el-table-column>
            <el-table-column prop="tank_volume_liter" label="液槽容积 (L)" width="120">
              <template #default="{ row }">{{ row.tank_volume_liter ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="planting_density_per_sqm" label="种植密度 (/m²)" width="120">
              <template #default="{ row }">{{ row.planting_density_per_sqm ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 'ENABLED' ? 'success' : 'info'" size="small">
                  {{ row.status === 'ENABLED' ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="router.push(`/assets/growing-zones?greenhouse_id=${row.greenhouse_id}`)">
                  查看
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!zonesLoading && zones.length === 0" description="暂无种植区" />
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else description="温室不存在" />

    <!-- Edit Dialog -->
    <el-dialog v-model="editDialogVisible" title="编辑温室" width="500px">
      <el-form ref="editFormRef" :model="editForm" :rules="editFormRules" label-width="120px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="位置">
          <el-input v-model="editForm.location" />
        </el-form-item>
        <el-form-item label="面积 (m²)">
          <el-input-number v-model="editForm.area_sqm" :min="0" :precision="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editForm.status" active-value="ENABLED" inactive-value="DISABLED" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSubmitting" @click="handleEditSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft, Edit, Delete } from '@element-plus/icons-vue'
import { greenhouseApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime } from '@/utils/format'
import type { Greenhouse, GrowingZone } from '@/types'
import { Role } from '@/types'

const route = useRoute()
const router = useRouter()

const greenhouseId = computed(() => Number(route.params.id))
const { hasRole } = usePermission()
const canManage = computed(() => hasRole(Role.ADMIN))
const canDelete = computed(() => hasRole(Role.ADMIN))

const loading = ref(false)
const errorMsg = ref('')
const greenhouse = ref<Greenhouse | null>(null)
const activeTab = ref('info')

const zones = ref<GrowingZone[]>([])
const zonesLoading = ref(false)

function systemTypeName(type: string) {
  const map: Record<string, string> = { DWC: 'DWC', NFT: 'NFT', EBB_FLOW: '潮汐', DRIP: '滴灌' }
  return map[type] || type
}

async function loadData() {
  if (!greenhouseId.value) {
    errorMsg.value = '无效的温室 ID'
    return
  }
  errorMsg.value = ''
  loading.value = true
  try {
    greenhouse.value = await greenhouseApi.getGreenhouse(greenhouseId.value)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载温室信息失败'
    errorMsg.value = msg
  }
  loading.value = false
  // Load zones in background
  loadZones()
}

async function loadZones() {
  if (!greenhouseId.value) return
  zonesLoading.value = true
  try {
    const data = await greenhouseApi.getGreenhouseZones(greenhouseId.value)
    zones.value = data.items
  } catch {
    zones.value = []
  } finally {
    zonesLoading.value = false
  }
}

// Edit dialog
const editDialogVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editSubmitting = ref(false)

const editForm = reactive({
  name: '',
  location: '' as string,
  area_sqm: undefined as number | undefined,
  description: '' as string,
  status: 'ENABLED' as string
})

const editFormRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

function openEditDialog() {
  if (!greenhouse.value) return
  editForm.name = greenhouse.value.name
  editForm.location = greenhouse.value.location || ''
  editForm.area_sqm = greenhouse.value.area_sqm
  editForm.description = greenhouse.value.description || ''
  editForm.status = greenhouse.value.status
  editDialogVisible.value = true
}

async function handleEditSubmit() {
  if (!editFormRef.value) return
  try {
    await editFormRef.value.validate()
  } catch { return }
  if (!greenhouseId.value) return

  editSubmitting.value = true
  try {
    await greenhouseApi.updateGreenhouse(greenhouseId.value, {
      name: editForm.name,
      location: editForm.location || undefined,
      area_sqm: editForm.area_sqm,
      description: editForm.description || undefined,
      status: editForm.status
    })
    ElMessage.success('温室已更新')
    editDialogVisible.value = false
    await loadData()
  } catch { /* handled by interceptor */ }
  finally { editSubmitting.value = false }
}

async function handleDelete() {
  await ElMessageBox.confirm('确认删除该温室？', '提示', { type: 'warning' })
  try {
    await greenhouseApi.deleteGreenhouse(greenhouseId.value)
    ElMessage.success('已删除')
    router.push('/assets/greenhouses')
  } catch { /* handled by interceptor */ }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.greenhouse-detail-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
      .page-title {
        font-size: 22px;
        font-weight: 700;
        color: var(--color-text-primary);
        margin: 0;
      }
    }
    .header-right { display: flex; gap: 8px; }
  }
  .loading-container {
    padding: 40px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }
  .info-card { margin-bottom: 16px; }
  .card-title { font-size: 15px; font-weight: 600; }
  .stats-row {
    margin-bottom: 16px;
    .stat-card {
      text-align: center;
      padding: 20px 16px;
      background: var(--bg-card);
      border-radius: var(--radius-md);
      box-shadow: var(--shadow-card);
      .stat-value { font-size: 28px; font-weight: 700; color: var(--color-primary); }
      .stat-label { font-size: 13px; color: var(--color-text-secondary); margin-top: 4px; }
    }
  }
}
</style>
