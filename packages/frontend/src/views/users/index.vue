<template>
    <div class="users-page">
      <div class="page-header">
        <h1 class="page-title">用户管理</h1>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="启用" value="ENABLED" />
          <el-option label="禁用" value="DISABLED" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索用户名/昵称" clearable style="width: 200px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 用户列表 -->
      <div class="table-container">
        <el-table :data="users" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="120" />
          <el-table-column prop="nickname" label="昵称" width="120">
            <template #default="{ row }">
              {{ row.nickname || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="phone" label="手机" width="130">
            <template #default="{ row }">
              {{ row.phone || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="email" label="邮箱" min-width="180">
            <template #default="{ row }">
              {{ row.email || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="roles" label="角色" width="150">
            <template #default="{ row }">
              <el-tag v-for="role in row.roles" :key="role" size="small" class="role-tag">
                {{ getRoleName(role) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === UserStatus.ENABLED ? 'success' : 'danger'">
                {{ row.status === UserStatus.ENABLED ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
              <el-button
                v-if="row.status === UserStatus.ENABLED"
                type="danger"
                link
                @click="handleToggleStatus(row, UserStatus.DISABLED)"
              >
                禁用
              </el-button>
              <el-button v-else type="success" link @click="handleToggleStatus(row, UserStatus.ENABLED)">
                启用
              </el-button>
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

      <!-- 新增/编辑用户弹窗 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="500px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="formData.username" :disabled="isEdit" placeholder="请输入用户名" maxlength="32" autocomplete="username" name="username" />
          </el-form-item>
          <el-form-item v-if="!isEdit" label="密码" prop="password">
            <el-input v-model="formData.password" type="password" placeholder="请输入密码" maxlength="64" show-password autocomplete="new-password" name="password" />
          </el-form-item>
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="formData.nickname" placeholder="请输入昵称（可选）" maxlength="32" />
          </el-form-item>
          <el-form-item label="手机" prop="phone">
            <el-input v-model="formData.phone" placeholder="请输入手机号（可选）" maxlength="20" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="formData.email" placeholder="请输入邮箱（可选）" maxlength="64" />
          </el-form-item>
          <el-form-item label="角色" prop="roles">
            <el-select v-model="formData.roles" multiple placeholder="请选择角色" style="width: 100%">
              <el-option label="管理员" value="ADMIN" />
              <el-option label="操作员" value="OPERATOR" />
              <el-option label="只读用户" value="VIEWER" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { userApi } from '@/api'
import { formatDateTime, getRoleName } from '@/utils/format'
import { Role, UserStatus } from '@/types'
import type { User, UserFormData, CreateUserParams } from '@/types'

// 数据
const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)

// 筛选条件
const filters = reactive({
  status: '',
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

const formData = reactive<UserFormData>({
  username: '',
  password: '',
  nickname: '',
  phone: '',
  email: '',
  roles: [Role.VIEWER]
})

const formRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度为 3-32 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 64, message: '密码长度为 6-64 个字符', trigger: 'blur' }
  ],
  roles: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await userApi.getUsers({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: (filters.status as UserStatus) || undefined,
      keyword: filters.keyword || undefined
    })
    users.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 重置筛选
function resetFilters() {
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  fetchData()
}

// 打开新增弹窗
function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    username: '',
    password: '',
    nickname: '',
    phone: '',
    email: '',
    roles: [Role.VIEWER]
  })
  dialogVisible.value = true
}

// 打开编辑弹窗
function openEditDialog(user: User) {
  isEdit.value = true
  editingId.value = user.id
  Object.assign(formData, {
    username: user.username,
    password: '',
    nickname: user.nickname || '',
    phone: user.phone || '',
    email: user.email || '',
    roles: user.roles
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
      await userApi.updateUser(editingId.value, {
        nickname: formData.nickname,
        phone: formData.phone,
        email: formData.email,
        roles: formData.roles
      })
      ElMessage.success('用户更新成功')
    } else {
      const params: CreateUserParams = {
        username: formData.username!,
        password: formData.password!,
        roles: formData.roles,
        nickname: formData.nickname || undefined,
        phone: formData.phone || undefined,
        email: formData.email || undefined
      }
      await userApi.createUser(params)
      ElMessage.success('用户创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

// 切换用户状态
async function handleToggleStatus(user: User, status: UserStatus) {
  try {
    await userApi.updateUserStatus(user.id, status)
    ElMessage.success(status === UserStatus.ENABLED ? '用户已启用' : '用户已禁用')
    fetchData()
  } catch {
    // 错误已处理
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.users-page {
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

  .role-tag {
    margin-right: 4px;
  }
}
</style>
