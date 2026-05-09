<template>
  <div class="growing-zones-page">
    <div class="page-header">
      <h1 class="page-title">种植区管理</h1>
      <el-button v-if="canManage" type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增种植区
      </el-button>
    </div>

    <div class="filter-section">
      <el-select
        v-model="filters.greenhouse_id"
        placeholder="选择温室"
        clearable
        filterable
        style="width: 200px"
      >
        <el-option
          v-for="g in greenhouses"
          :key="g.id"
          :label="`${g.name} (ID:${g.id})`"
          :value="g.id"
        />
      </el-select>
      <el-select
        v-model="filters.system_type"
        placeholder="系统类型"
        clearable
        style="width: 160px"
      >
        <el-option label="DWC" value="DWC" />
        <el-option label="NFT" value="NFT" />
        <el-option label="EBB_FLOW" value="EBB_FLOW" />
        <el-option label="DRIP" value="DRIP" />
      </el-select>
      <el-select
        v-model="filters.status"
        placeholder="状态"
        clearable
        style="width: 140px"
      >
        <el-option label="启用" value="ENABLED" />
        <el-option label="停用" value="DISABLED" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="zones" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="温室" width="160">
          <template #default="{ row }">{{ greenhouseName(row.greenhouse_id) }}</template>
        </el-table-column>
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="system_type" label="系统类型" width="120">
          <template #default="{ row }">{{
            systemTypeName(row.system_type)
          }}</template>
        </el-table-column>
        <el-table-column label="槽容积(L)" width="120">
          <template #default="{ row }">{{
            row.tank_volume_liter ?? "-"
          }}</template>
        </el-table-column>
        <el-table-column label="种植密度" width="120">
          <template #default="{ row }">{{
            row.planting_density_per_sqm ?? "-"
          }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-if="canManage"
              v-model="row.status"
              active-value="ENABLED"
              inactive-value="DISABLED"
              :loading="statusLoading[row.id] === true"
              @change="(val: 'ENABLED' | 'DISABLED') => updateStatus(row, val)"
            />
            <el-tag
              v-else
              :type="row.status === 'ENABLED' ? 'success' : 'info'"
            >
              {{ row.status === "ENABLED" ? "启用" : "停用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{
            formatDateTime(row.created_at)
          }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="canManage"
              type="primary"
              link
              @click="openEditDialog(row)"
              >编辑</el-button
            >
            <el-button
              v-if="canDelete"
              type="danger"
              link
              @click="removeZone(row.id)"
              >删除</el-button
            >
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

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑种植区' : '新增种植区'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="130px"
      >
        <el-form-item label="温室" prop="greenhouse_id">
          <el-select
            v-model="formData.greenhouse_id"
            placeholder="选择温室"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="g in greenhouses"
              :key="g.id"
              :label="`${g.name} (ID:${g.id})`"
              :value="g.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input
            v-model="formData.code"
            placeholder="请输入编号"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入名称"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item label="系统类型">
          <el-select v-model="formData.system_type" style="width: 100%">
            <el-option label="DWC" value="DWC" />
            <el-option label="NFT" value="NFT" />
            <el-option label="EBB_FLOW" value="EBB_FLOW" />
            <el-option label="DRIP" value="DRIP" />
          </el-select>
        </el-form-item>
        <el-form-item label="槽容积(L)">
          <el-input-number
            v-model="formData.tank_volume_liter"
            :min="0"
            :precision="1"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="种植密度(株/㎡)">
          <el-input-number
            v-model="formData.planting_density_per_sqm"
            :min="0"
            :precision="1"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          v-if="canManage"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
          >确定</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox, FormInstance, FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { greenhouseApi } from "@/api";
import { usePermission } from "@/composables/usePermission";
import { formatDateTime } from "@/utils/format";
import { buildIdLabelMap, fallbackIdLabel, greenhouseLabel } from "@/utils/labels";
import { LARGE_PAGE_SIZE } from "@/utils/constants";
import { Role } from "@/types";
import type { GrowingZone, Greenhouse } from "@/types";

const route = useRoute();
const { canControlDevice, hasRole } = usePermission();
const canManage = computed(() => canControlDevice());
const canDelete = computed(() => hasRole(Role.ADMIN));

const loading = ref(false);
const zones = ref<GrowingZone[]>([]);
const total = ref(0);
const greenhouses = ref<Greenhouse[]>([]);
const statusLoading = reactive<Record<number, boolean>>({});

const greenhouseLabelById = computed(() =>
  buildIdLabelMap(greenhouses.value, (g) => g.id, greenhouseLabel, "温室")
);

function greenhouseName(greenhouseId?: number) {
  if (!greenhouseId) return fallbackIdLabel("温室", greenhouseId);
  return greenhouseLabelById.value[greenhouseId] || fallbackIdLabel("温室", greenhouseId);
}

const filters = reactive({
  greenhouse_id: undefined as number | undefined,
  system_type: "" as string,
  status: "" as string,
});

const pagination = reactive({ page: 1, pageSize: 20 });

const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();
const submitLoading = ref(false);
const editingId = ref<number | null>(null);

const formData = reactive({
  greenhouse_id: undefined as number | undefined,
  code: "",
  name: "",
  system_type: "DWC" as string,
  tank_volume_liter: undefined as number | undefined,
  planting_density_per_sqm: undefined as number | undefined,
});

const formRules: FormRules = {
  greenhouse_id: [{ required: true, message: "请选择温室", trigger: "change" }],
  code: [
    { required: true, message: "请输入编号", trigger: "blur" },
    { min: 1, max: 64, message: "编号长度为 1-64 个字符", trigger: "blur" },
  ],
  name: [
    { required: true, message: "请输入名称", trigger: "blur" },
    { min: 1, max: 64, message: "名称长度为 1-64 个字符", trigger: "blur" },
  ],
};

function systemTypeName(type: string) {
  const map: Record<string, string> = {
    DWC: "DWC",
    NFT: "NFT",
    EBB_FLOW: "潮汐",
    DRIP: "滴灌",
  };
  return map[type] || type;
}

async function fetchData() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize,
    };
    if (filters.greenhouse_id) params.greenhouse_id = filters.greenhouse_id;
    if (filters.system_type) params.system_type = filters.system_type;
    if (filters.status) params.status = filters.status;
    const data = await greenhouseApi.getGrowingZones(params);
    zones.value = data.items;
    total.value = data.total;
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false;
  }
}

function resetFilters() {
  filters.greenhouse_id = undefined;
  filters.system_type = "";
  filters.status = "";
  pagination.page = 1;
  fetchData();
}

function openCreateDialog() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  isEdit.value = false;
  editingId.value = null;
  Object.assign(formData, {
    greenhouse_id: filters.greenhouse_id,
    code: "",
    name: "",
    system_type: "DWC",
    tank_volume_liter: undefined,
    planting_density_per_sqm: undefined,
  });
  dialogVisible.value = true;
}

function openEditDialog(zone: GrowingZone) {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  isEdit.value = true;
  editingId.value = zone.id;
  Object.assign(formData, {
    greenhouse_id: zone.greenhouse_id,
    code: zone.code,
    name: zone.name,
    system_type: zone.system_type || "DWC",
    tank_volume_liter: zone.tank_volume_liter,
    planting_density_per_sqm: zone.planting_density_per_sqm,
  });
  dialogVisible.value = true;
}

async function handleSubmit() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  if (!formRef.value) return;
  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  submitLoading.value = true;
  try {
    const payload = {
      greenhouse_id: formData.greenhouse_id!,
      code: formData.code,
      name: formData.name,
      system_type: formData.system_type || undefined,
      tank_volume_liter: formData.tank_volume_liter,
      planting_density_per_sqm: formData.planting_density_per_sqm,
    };
    if (isEdit.value && editingId.value) {
      await greenhouseApi.updateGrowingZone(editingId.value, payload);
      ElMessage.success("种植区已更新");
    } else {
      await greenhouseApi.createGrowingZone(payload);
      ElMessage.success("种植区已创建");
    }
    dialogVisible.value = false;
    fetchData();
  } catch {
    // handled by interceptor
  } finally {
    submitLoading.value = false;
  }
}

async function removeZone(id: number) {
  if (!canDelete.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  try {
    await ElMessageBox.confirm("确认删除该种植区？", "提示", {
      type: "warning",
    });
    await greenhouseApi.deleteGrowingZone(id);
    ElMessage.success("已删除");
    fetchData();
  } catch (e: any) {
    if (e !== "cancel") {
    }
  }
}

async function updateStatus(zone: GrowingZone, status: "ENABLED" | "DISABLED") {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    zone.status = status === "ENABLED" ? "DISABLED" : "ENABLED";
    return;
  }
  const prev = status === "ENABLED" ? "DISABLED" : "ENABLED";
  statusLoading[zone.id] = true;
  try {
    await greenhouseApi.updateGrowingZone(zone.id, { status });
    ElMessage.success(status === "ENABLED" ? "已启用" : "已停用");
  } catch {
    zone.status = prev;
  } finally {
    statusLoading[zone.id] = false;
  }
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({
      page_size: LARGE_PAGE_SIZE,
    });
    greenhouses.value = data.items;
  } catch {
    /* ignore */
  }
}

onMounted(() => {
  loadGreenhouses();
  // Support ?greenhouse_id= query param
  const qGreenhouseId = route.query.greenhouse_id;
  if (qGreenhouseId) {
    filters.greenhouse_id = Number(qGreenhouseId);
  }
  fetchData();
});
</script>

<style scoped lang="scss">
.growing-zones-page {
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
