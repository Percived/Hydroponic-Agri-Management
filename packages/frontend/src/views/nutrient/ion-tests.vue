<template>
  <div class="ion-tests-page">
    <div class="page-header">
      <h1 class="page-title">离子检测</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增检测
      </el-button>
    </div>

    <div class="filter-section">
      <el-input-number v-model="filters.tank_id" :min="1" placeholder="液槽ID" style="width: 140px" />
      <el-input-number v-model="filters.batch_id" :min="1" placeholder="批次ID" style="width: 140px" />
      <el-select v-model="filters.test_method" placeholder="检测方法" clearable style="width: 140px">
        <el-option label="实验室" value="LAB" />
        <el-option label="试纸" value="STRIP" />
        <el-option label="仪表" value="METER" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="tests" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tank_id" label="液槽ID" width="100" />
        <el-table-column prop="sample_code" label="样品编号" width="140" />
        <el-table-column prop="test_method" label="检测方法" width="100">
          <template #default="{ row }">{{ testMethodName(row.test_method) }}</template>
        </el-table-column>
        <el-table-column label="NO3-N" width="90">
          <template #default="{ row }">{{ formatNum(row.no3_n) }}</template>
        </el-table-column>
        <el-table-column label="NH4-N" width="90">
          <template #default="{ row }">{{ formatNum(row.nh4_n) }}</template>
        </el-table-column>
        <el-table-column label="P" width="90">
          <template #default="{ row }">{{ formatNum(row.p) }}</template>
        </el-table-column>
        <el-table-column label="K" width="90">
          <template #default="{ row }">{{ formatNum(row.k) }}</template>
        </el-table-column>
        <el-table-column label="Ca" width="90">
          <template #default="{ row }">{{ formatNum(row.ca) }}</template>
        </el-table-column>
        <el-table-column label="Mg" width="90">
          <template #default="{ row }">{{ formatNum(row.mg) }}</template>
        </el-table-column>
        <el-table-column label="EC" width="90">
          <template #default="{ row }">{{ formatNum(row.ec_at_sample) }}</template>
        </el-table-column>
        <el-table-column label="pH" width="90">
          <template #default="{ row }">{{ formatNum(row.ph_at_sample) }}</template>
        </el-table-column>
        <el-table-column prop="sampled_at" label="采样时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.sampled_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="removeTest(row.id)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑离子检测' : '新增离子检测'" width="700px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="110px">
        <el-row :gutter="16">
          <!-- 基础信息 -->
          <el-col :span="12">
            <el-form-item label="液槽ID" prop="tank_id">
              <el-input-number v-model="formData.tank_id" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="批次ID">
              <el-input-number v-model="formData.batch_id" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="样品编号" prop="sample_code">
              <el-input v-model="formData.sample_code" placeholder="请输入编号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测方法">
              <el-select v-model="formData.test_method" style="width: 100%">
                <el-option label="实验室" value="LAB" />
                <el-option label="试纸" value="STRIP" />
                <el-option label="仪表" value="METER" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="采样时间" prop="sampled_at">
              <el-date-picker
                v-model="formData.sampled_at"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测时间">
              <el-date-picker
                v-model="formData.tested_at"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 大量营养元素 -->
        <el-divider content-position="left">大量营养元素</el-divider>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="NO3-N">
              <el-input-number v-model="formData.no3_n" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="NH4-N">
              <el-input-number v-model="formData.nh4_n" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="P">
              <el-input-number v-model="formData.p" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="K">
              <el-input-number v-model="formData.k" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Ca">
              <el-input-number v-model="formData.ca" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Mg">
              <el-input-number v-model="formData.mg" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="S">
              <el-input-number v-model="formData.s" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 微量营养元素 -->
        <el-divider content-position="left">微量营养元素</el-divider>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="Fe">
              <el-input-number v-model="formData.fe" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Mn">
              <el-input-number v-model="formData.mn" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Zn">
              <el-input-number v-model="formData.zn" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="B">
              <el-input-number v-model="formData.b" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Cu">
              <el-input-number v-model="formData.cu" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Mo">
              <el-input-number v-model="formData.mo" :min="0" :precision="3" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 辅助信息 -->
        <el-divider content-position="left">辅助信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="采样EC">
              <el-input-number v-model="formData.ec_at_sample" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="采样pH">
              <el-input-number v-model="formData.ph_at_sample" :min="0" :max="14" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实验室">
              <el-input v-model="formData.lab_name" placeholder="实验室名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="报告URL">
              <el-input v-model="formData.report_url" placeholder="报告链接" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="formData.note" type="textarea" :rows="2" placeholder="备注信息" />
            </el-form-item>
          </el-col>
        </el-row>
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
import { nutrientApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { IonTestRecord } from '@/types'

const loading = ref(false)
const tests = ref<IonTestRecord[]>([])
const total = ref(0)

const filters = reactive({
  tank_id: undefined as number | undefined,
  batch_id: undefined as number | undefined,
  test_method: '' as string
})

const pagination = reactive({ page: 1, pageSize: 20 })

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const emptyForm = () => ({
  tank_id: undefined as number | undefined,
  batch_id: undefined as number | undefined,
  sample_code: '',
  sampled_at: '',
  tested_at: '' as string,
  test_method: 'LAB' as string,
  no3_n: undefined as number | undefined,
  nh4_n: undefined as number | undefined,
  p: undefined as number | undefined,
  k: undefined as number | undefined,
  ca: undefined as number | undefined,
  mg: undefined as number | undefined,
  s: undefined as number | undefined,
  fe: undefined as number | undefined,
  mn: undefined as number | undefined,
  zn: undefined as number | undefined,
  b: undefined as number | undefined,
  cu: undefined as number | undefined,
  mo: undefined as number | undefined,
  ec_at_sample: undefined as number | undefined,
  ph_at_sample: undefined as number | undefined,
  lab_name: '' as string,
  report_url: '' as string,
  note: '' as string
})

const formData = reactive(emptyForm())

const formRules: FormRules = {
  tank_id: [{ required: true, message: '请输入液槽ID', trigger: 'blur' }],
  sample_code: [
    { required: true, message: '请输入样品编号', trigger: 'blur' },
    { min: 1, max: 64, message: '编号长度为 1-64 个字符', trigger: 'blur' }
  ],
  sampled_at: [{ required: true, message: '请选择采样时间', trigger: 'change' }]
}

function testMethodName(method: string) {
  const map: Record<string, string> = { LAB: '实验室', STRIP: '试纸', METER: '仪表' }
  return map[method] || method
}

function formatNum(val: number | null | undefined) {
  if (val === null || val === undefined) return '-'
  return val.toFixed(2)
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.tank_id) params.tank_id = filters.tank_id
    if (filters.batch_id) params.batch_id = filters.batch_id
    if (filters.test_method) params.test_method = filters.test_method
    const data = await nutrientApi.getIonTests(params)
    tests.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.tank_id = undefined
  filters.batch_id = undefined
  filters.test_method = ''
  pagination.page = 1
  fetchData()
}

function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, emptyForm())
  dialogVisible.value = true
}

function openEditDialog(test: IonTestRecord) {
  isEdit.value = true
  editingId.value = test.id
  Object.assign(formData, {
    tank_id: test.tank_id,
    batch_id: test.batch_id,
    sample_code: test.sample_code,
    sampled_at: test.sampled_at,
    tested_at: test.tested_at || '',
    test_method: test.test_method,
    no3_n: test.no3_n,
    nh4_n: test.nh4_n,
    p: test.p,
    k: test.k,
    ca: test.ca,
    mg: test.mg,
    s: test.s,
    fe: test.fe,
    mn: test.mn,
    zn: test.zn,
    b: test.b,
    cu: test.cu,
    mo: test.mo,
    ec_at_sample: test.ec_at_sample,
    ph_at_sample: test.ph_at_sample,
    lab_name: test.lab_name || '',
    report_url: test.report_url || '',
    note: test.note || ''
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    const payload = {
      tank_id: formData.tank_id!,
      batch_id: formData.batch_id || undefined,
      sample_code: formData.sample_code,
      sampled_at: formData.sampled_at,
      test_method: formData.test_method || undefined,
      tested_at: formData.tested_at || undefined,
      no3_n: formData.no3_n,
      nh4_n: formData.nh4_n,
      p: formData.p,
      k: formData.k,
      ca: formData.ca,
      mg: formData.mg,
      s: formData.s,
      fe: formData.fe,
      mn: formData.mn,
      zn: formData.zn,
      b: formData.b,
      cu: formData.cu,
      mo: formData.mo,
      ec_at_sample: formData.ec_at_sample,
      ph_at_sample: formData.ph_at_sample,
      lab_name: formData.lab_name || undefined,
      report_url: formData.report_url || undefined,
      note: formData.note || undefined
    }
    if (isEdit.value && editingId.value) {
      await nutrientApi.updateIonTest(editingId.value, payload)
      ElMessage.success('离子检测已更新')
    } else {
      await nutrientApi.createIonTest(payload)
      ElMessage.success('离子检测已创建')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    submitLoading.value = false
  }
}

async function removeTest(id: number) {
  await ElMessageBox.confirm('确认删除该离子检测记录？', '提示', { type: 'warning' })
  await nutrientApi.deleteIonTest(id)
  ElMessage.success('已删除')
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.ion-tests-page {
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
  }
  .filter-section {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
    margin-bottom: 16px;
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
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
}
</style>
