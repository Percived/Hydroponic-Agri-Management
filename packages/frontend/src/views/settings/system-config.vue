<template>
    <div class="system-config-page">
      <div class="page-header">
        <h1 class="page-title">系统配置</h1>
      </div>

      <div class="table-container">
        <el-table :data="configs" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="config_key" label="配置键" min-width="180" />
          <el-table-column prop="config_value" label="配置值" min-width="200">
            <template #default="{ row }">
              <template v-if="editingKey === row.config_key">
                <el-input v-model="editingValue" size="small" style="width: 200px" @keyup.enter="saveEdit(row)" @keyup.escape="cancelEdit" />
                <el-button type="primary" size="small" @click="saveEdit(row)" :loading="saveLoading" style="margin-left: 8px">保存</el-button>
                <el-button size="small" @click="cancelEdit">取消</el-button>
              </template>
              <template v-else>
                <span
                  :class="{ 'editable-cell': !isSensitiveKey(row.config_key) }"
                  @click="startEdit(row)"
                >{{ row.config_value }}</span>
              </template>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="说明" min-width="200" show-overflow-tooltip />
          <el-table-column prop="updated_at" label="更新时间" width="160">
            <template #default="{ row }">{{ formatDate(row.updated_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right" v-if="false">
            <template #default="{ row }">
              <el-button v-if="!isSensitiveKey(row.config_key)" type="primary" link @click="startEdit(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSystemConfigs, updateSystemConfig, SystemConfigItem } from '@/api/system-config'
import { formatDate } from '@/utils/format'

const configs = ref<SystemConfigItem[]>([])
const loading = ref(false)
const editingKey = ref<string | null>(null)
const editingValue = ref('')
const saveLoading = ref(false)

const sensitiveKeys = ['jwt_secret', 'db_password', 'mqtt_password']

function isSensitiveKey(key: string): boolean {
  return sensitiveKeys.includes(key)
}

async function fetchConfigs() {
  loading.value = true
  try {
    const data = await getSystemConfigs()
    configs.value = data.items
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function startEdit(row: SystemConfigItem) {
  if (isSensitiveKey(row.config_key)) return
  editingKey.value = row.config_key
  editingValue.value = row.config_value
}

function cancelEdit() {
  editingKey.value = null
  editingValue.value = ''
}

async function saveEdit(row: SystemConfigItem) {
  if (!editingKey.value) return
  saveLoading.value = true
  try {
    await updateSystemConfig({
      config_key: row.config_key,
      config_value: editingValue.value,
      description: row.description
    })
    ElMessage.success('配置更新成功')
    cancelEdit()
    fetchConfigs()
  } catch {} finally { saveLoading.value = false }
}

onMounted(() => { fetchConfigs() })
</script>

<style scoped lang="scss">
.system-config-page {
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

  .editable-cell {
    cursor: pointer;
    color: var(--color-primary);

    &:hover {
      text-decoration: underline;
    }
  }
}
</style>
