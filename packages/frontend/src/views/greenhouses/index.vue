<template>
  <AppLayout>
    <div class="greenhouses-page">
      <div class="page-header">
        <h1 class="page-title">温室管理</h1>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增温室
        </el-button>
      </div>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-input v-model="filters.keyword" placeholder="搜索温室名称" clearable style="width: 200px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 温室列表 -->
      <div class="table-container">
        <el-table :data="greenhouses" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="温室名称" width="150" />
          <el-table-column prop="location" label="位置" width="150">
            <template #default="{ row }">
              {{ row.location || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="200">
            <template #default="{ row }">
              {{ row.description || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
              <el-button type="primary" link @click="goGroups(row)">分组管理</el-button>
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
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑温室' : '新增温室'" width="500px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
          <el-form-item label="名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入温室名称" maxlength="64" />
          </el-form-item>
          <el-form-item label="位置" prop="location">
            <el-input v-model="formData.location" placeholder="请输入位置（可选）" maxlength="128" />
          </el-form-item>
          <el-form-item label="描述" prop="description">
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="3"
              placeholder="请输入描述（可选）"
              maxlength="255"
              show-word-limit
            />
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
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { AppLayout } from '@/components/layout'
import { greenhouseApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { Greenhouse, GreenhouseFormData } from '@/types'

const router = useRouter()

// 数据
const loading = ref(false)
const greenhouses = ref<Greenhouse[]>([])
const total = ref(0)

// 筛选条件
const filters = reactive({
  keyword: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 弹窗
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive<GreenhouseFormData>({
  name: '',
  location: '',
  description: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入温室名称', trigger: 'blur' },
    { min: 1, max: 64, message: '温室名称长度为 1-64 个字符', trigger: 'blur' }
  ]
}

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await greenhouseApi.getGreenhouses({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: filters.keyword || undefined
    })
    greenhouses.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 重置筛选
function resetFilters() {
  filters.keyword = ''
  pagination.page = 1
  fetchData()
}

// 打开新增弹窗
function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    name: '',
    location: '',
    description: ''
  })
  dialogVisible.value = true
}

// 打开编辑弹窗
function openEditDialog(greenhouse: Greenhouse) {
  isEdit.value = true
  editingId.value = greenhouse.id
  Object.assign(formData, {
    name: greenhouse.name,
    location: greenhouse.location || '',
    description: greenhouse.description || ''
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
      await greenhouseApi.updateGreenhouse(editingId.value, {
        name: formData.name,
        location: formData.location || undefined,
        description: formData.description || undefined
      })
      ElMessage.success('温室更新成功')
    } else {
      await greenhouseApi.createGreenhouse(formData)
      ElMessage.success('温室创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

// 跳转分组管理
function goGroups(greenhouse: Greenhouse) {
  router.push({ path: '/device-groups', query: { greenhouse_id: greenhouse.id } })
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.greenhouses-page {
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
}
</style>
