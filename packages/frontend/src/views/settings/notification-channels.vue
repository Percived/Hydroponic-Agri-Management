<template>
    <div class="notification-page">
      <div class="page-header">
        <h1 class="page-title">通知渠道</h1>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>新增渠道
        </el-button>
      </div>

      <!-- 数据表格 -->
      <div class="table-container">
        <el-table :data="channels" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="channel_type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ ChannelTypeNames[row.channel_type] || row.channel_type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="min_alert_level" label="最低告警级别" width="120">
            <template #default="{ row }">
              <el-tag :type="row.min_alert_level === 'CRITICAL' ? 'danger' : row.min_alert_level === 'WARN' ? 'warning' : 'info'" size="small">
                {{ row.min_alert_level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="启用" width="80">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                {{ row.enabled ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="160">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
              <el-button
                v-if="row.channel_type === 'WEBHOOK'"
                type="success"
                link
                @click="testChannel(row.id)"
                :loading="testLoading[row.id]"
              >测试</el-button>
              <el-button type="danger" link @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 新增/编辑弹窗 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑通知渠道' : '新增通知渠道'" width="550px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
          <el-form-item label="渠道名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入渠道名称" maxlength="64" />
          </el-form-item>
          <el-form-item label="渠道类型" prop="channel_type">
            <el-select v-model="formData.channel_type" placeholder="请选择渠道类型" :disabled="isEdit" style="width: 100%">
              <el-option label="邮件" value="EMAIL" />
              <el-option label="短信" value="SMS" />
              <el-option label="Webhook" value="WEBHOOK" />
            </el-select>
          </el-form-item>
          <el-form-item label="配置信息" prop="config">
            <el-input v-model="formData.configStr" type="textarea" placeholder='JSON 格式，Webhook 示例: {"url":"https://...","secret":"..."}' rows="4" />
          </el-form-item>
          <el-form-item label="最低告警级别">
            <el-select v-model="formData.min_alert_level" placeholder="请选择级别" style="width: 100%">
              <el-option label="INFO" value="INFO" />
              <el-option label="WARN" value="WARN" />
              <el-option label="CRITICAL" value="CRITICAL" />
            </el-select>
          </el-form-item>
          <el-form-item label="启用">
            <el-switch v-model="formData.enabled" />
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
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { notificationApi } from '@/api'
import { formatDate } from '@/utils/format'
import { NotificationChannel, ChannelTypeNames } from '@/types'

const channels = ref<NotificationChannel[]>([])
const loading = ref(false)

// 弹窗
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)
const testLoading = ref<Record<number, boolean>>({})

const formData = reactive({
  name: '',
  channel_type: 'WEBHOOK',
  configStr: '',
  min_alert_level: 'WARN',
  enabled: true
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }, { min: 1, max: 64, message: '长度为1-64个字符', trigger: 'blur' }],
  channel_type: [{ required: true, message: '请选择渠道类型', trigger: 'change' }],
  config: [{ required: true, message: '请输入配置信息', trigger: 'blur' }]
}

async function fetchChannels() {
  loading.value = true
  try {
    const data = await notificationApi.getChannels()
    channels.value = data.items
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function openCreateDialog() {
  isEdit.value = false; editingId.value = null
  Object.assign(formData, { name: '', channel_type: 'WEBHOOK', configStr: '', min_alert_level: 'WARN', enabled: true })
  dialogVisible.value = true
}

function openEditDialog(row: NotificationChannel) {
  isEdit.value = true; editingId.value = row.id
  formData.name = row.name
  formData.channel_type = row.channel_type
  formData.configStr = JSON.stringify(row.config, null, 2)
  formData.min_alert_level = row.min_alert_level
  formData.enabled = row.enabled
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try { await formRef.value.validate() } catch { return }

  let config: any
  try { config = JSON.parse(formData.configStr) }
  catch { ElMessage.warning('配置信息 JSON 格式不正确'); return }

  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      await notificationApi.updateChannel(editingId.value, {
        name: formData.name, config, min_alert_level: formData.min_alert_level, enabled: formData.enabled
      })
      ElMessage.success('渠道更新成功')
    } else {
      await notificationApi.createChannel({
        channel_type: formData.channel_type, name: formData.name, config,
        min_alert_level: formData.min_alert_level, enabled: formData.enabled
      })
      ElMessage.success('渠道创建成功')
    }
    dialogVisible.value = false
    fetchChannels()
  } catch {} finally { submitLoading.value = false }
}

async function testChannel(id: number) {
  testLoading.value[id] = true
  try {
    const data = await notificationApi.testChannel(id)
    ElMessage.success(data.sent ? '测试发送成功' : '测试发送失败')
  } catch {} finally { testLoading.value[id] = false }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该通知渠道吗？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    await notificationApi.deleteChannel(id)
    ElMessage.success('删除成功')
    fetchChannels()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

onMounted(() => { fetchChannels() })
</script>

<style scoped lang="scss">
.notification-page {
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
}
</style>
