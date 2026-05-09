<template>
  <div class="device-detail-page">
    <div class="page-header">
      <el-button @click="goBack" :icon="ArrowLeft">返回列表</el-button>
      <h1 class="page-title">
        设备详情 - {{ device?.device_code || "加载中..." }}
      </h1>
      <el-tag
        v-if="device"
        :type="deviceType === 'sensor' ? 'primary' : 'warning'"
        size="large"
      >
        {{ deviceType === "sensor" ? "传感器设备" : "执行器设备" }}
      </el-tag>
      <el-button
        v-if="device && canManage"
        type="primary"
        @click="openEditDialog"
      >
        <el-icon><Edit /></el-icon>编辑
      </el-button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="errorMsg" class="error-container">
      <el-result
        icon="error"
        :title="errorMsg"
        sub-title="请检查网络连接或返回列表重试"
      >
        <template #extra>
          <el-button type="primary" @click="loadData">重新加载</el-button>
          <el-button @click="goBack">返回列表</el-button>
        </template>
      </el-result>
    </div>

    <!-- 设备数据 -->
    <template v-else-if="device">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <el-card class="info-card">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="设备编码">{{
                device.device_code
              }}</el-descriptions-item>
              <el-descriptions-item label="设备名称">{{
                device.name
              }}</el-descriptions-item>
              <el-descriptions-item label="设备型号">{{
                device.model || "-"
              }}</el-descriptions-item>
              <el-descriptions-item label="固件版本">{{
                device.firmware_version || "-"
              }}</el-descriptions-item>
              <el-descriptions-item label="通信协议">{{
                device.protocol
              }}</el-descriptions-item>
              <el-descriptions-item label="所属温室">{{
                getGreenhouseName(device.greenhouse_id)
              }}</el-descriptions-item>
              <el-descriptions-item label="种植区">{{
                getZoneName(device.growing_zone_id)
              }}</el-descriptions-item>
              <el-descriptions-item label="在线状态">
                <el-tag
                  :type="
                    device.status === 'ONLINE'
                      ? 'success'
                      : device.status === 'FAULT'
                        ? 'warning'
                        : 'danger'
                  "
                >
                  {{
                    device.status === "ONLINE"
                      ? "在线"
                      : device.status === "FAULT"
                        ? "故障"
                        : "离线"
                  }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="最后上报">
                {{
                  device.last_seen_at ? formatDate(device.last_seen_at) : "-"
                }}
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">{{
                formatDate(device.created_at)
              }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{
                formatDate(device.updated_at)
              }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 通道列表 -->
          <el-card class="channels-card">
            <template #header>
              <div class="channels-card-header">
                <span>{{
                  deviceType === "sensor" ? "传感器通道" : "执行器通道"
                }}</span>
                <el-button
                  v-if="canManage"
                  size="small"
                  type="primary"
                  @click="openChannelCreateDialog"
                >
                  <el-icon><Plus /></el-icon>添加通道
                </el-button>
              </div>
            </template>
            <div v-if="channels.length === 0" class="channels-empty">
              暂无通道，点击上方按钮添加
            </div>
            <el-table v-else :data="channels" stripe size="small">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column
                prop="channel_code"
                label="通道编码"
                width="130"
              />
              <template v-if="deviceType === 'sensor'">
                <el-table-column
                  prop="metric_code"
                  label="指标代码"
                  width="100"
                >
                  <template #default="{ row }">{{
                    getMetricName(row.metric_code)
                  }}</template>
                </el-table-column>
                <el-table-column prop="unit" label="单位" width="80" />
                <el-table-column
                  prop="precision_digits"
                  label="精度"
                  width="70"
                />
                <el-table-column label="量程" width="150">
                  <template #default="{ row }"
                    >{{ row.range_min ?? "-" }} ~
                    {{ row.range_max ?? "-" }}</template
                  >
                </el-table-column>
                <el-table-column
                  prop="sampling_interval_sec"
                  label="采样间隔(s)"
                  width="110"
                />
              </template>
              <template v-else>
                <el-table-column
                  prop="actuator_type"
                  label="类型"
                  width="100"
                />
                <el-table-column
                  prop="current_state"
                  label="当前状态"
                  width="100"
                />
                <el-table-column
                  prop="rated_power_watt"
                  label="额定功率(W)"
                  width="120"
                >
                  <template #default="{ row }">{{
                    row.rated_power_watt ?? "-"
                  }}</template>
                </el-table-column>
              </template>
              <el-table-column prop="enabled" label="启用" width="70">
                <template #default="{ row }">
                  <el-switch
                    v-if="canManage"
                    v-model="row.enabled"
                    :loading="channelToggleLoading[row.id] === true"
                    @change="(val: boolean) => toggleChannelEnabled(row, val)"
                  />
                  <el-tag
                    v-else
                    :type="row.enabled ? 'success' : 'info'"
                    size="small"
                  >
                    {{ row.enabled ? "是" : "否" }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                prop="last_reported_at"
                label="最后上报"
                width="160"
              >
                <template #default="{ row }">{{
                  formatDate(row.last_reported_at)
                }}</template>
              </el-table-column>
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-if="canManage"
                    type="primary"
                    link
                    @click="openChannelEditDialog(row)"
                    >编辑</el-button
                  >
                  <el-button
                    v-if="canDelete"
                    type="danger"
                    link
                    @click="handleChannelDelete(row)"
                    >删除</el-button
                  >
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- Tab 2: 遥测数据（仅传感器设备） -->
        <el-tab-pane
          v-if="deviceType === 'sensor'"
          label="遥测数据"
          name="telemetry"
        >
          <div v-if="telemetryLoading" class="loading-placeholder">
            <el-skeleton :rows="3" animated />
          </div>
          <div
            v-else-if="telemetryByChannel.length === 0"
            class="empty-placeholder"
          >
            <el-empty description="暂无遥测数据" />
          </div>
          <div v-else class="telemetry-grid">
            <div
              v-for="item in telemetryByChannel"
              :key="`${item.channel_id}-${item.metric_code}`"
              class="telemetry-item"
            >
              <div class="metric-name">
                {{ getMetricName(item.metric_code) }}
              </div>
              <div class="metric-value">
                {{ formatNumber(item.value) }}
                <span class="metric-unit">{{
                  getChannelUnit(item.channel_id) || ""
                }}</span>
              </div>
              <div class="metric-meta">
                <el-tag
                  :type="item.quality_flag === 'normal' ? 'success' : 'danger'"
                  size="small"
                >
                  {{
                    item.quality_flag === "normal" ? "正常" : item.quality_flag
                  }}
                </el-tag>
                <span class="metric-time">{{
                  formatDate(item.collected_at, "HH:mm:ss")
                }}</span>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editDialogVisible" title="编辑设备" width="550px">
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="100px"
      >
        <el-form-item label="设备编码">
          <el-input :model-value="device?.device_code" disabled />
        </el-form-item>
        <el-form-item label="设备名称" prop="name">
          <el-input
            v-model="editForm.name"
            placeholder="请输入设备名称"
            maxlength="64"
            autocomplete="off"
            name="edit_name"
          />
        </el-form-item>
        <el-form-item label="设备型号" prop="model">
          <el-input
            v-model="editForm.model"
            placeholder="请输入型号（选填）"
            maxlength="64"
            autocomplete="off"
            name="edit_model"
          />
        </el-form-item>
        <el-form-item label="固件版本" prop="firmware_version">
          <el-input
            v-model="editForm.firmware_version"
            placeholder="请输入固件版本（选填）"
            maxlength="32"
            autocomplete="off"
            name="edit_firmware"
          />
        </el-form-item>
        <el-form-item label="所属温室" prop="greenhouse_id">
          <el-select
            v-model="editForm.greenhouse_id"
            placeholder="请选择温室"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="gh in greenhouses"
              :key="gh.id"
              :label="gh.name"
              :value="gh.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="种植区" prop="growing_zone_id">
          <el-select
            v-model="editForm.growing_zone_id"
            placeholder="请选择种植区"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="zone in filteredEditZones"
              :key="zone.id"
              :label="zone.name"
              :value="zone.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button
          v-if="canManage"
          type="primary"
          :loading="editSubmitLoading"
          @click="handleEditSubmit"
          >确定</el-button
        >
      </template>
    </el-dialog>

    <!-- 通道编辑弹窗 -->
    <el-dialog
      v-model="channelDialogVisible"
      :title="isChannelEdit ? '编辑通道' : '添加通道'"
      width="500px"
    >
      <el-form
        ref="channelFormRef"
        :model="channelForm"
        :rules="channelFormRules"
        label-width="110px"
      >
        <el-form-item label="通道编码" prop="channel_code">
          <el-input
            v-model="channelForm.channel_code"
            placeholder="请输入通道编码"
            maxlength="64"
            autocomplete="off"
            name="ch_code"
          />
        </el-form-item>
        <template v-if="deviceType === 'sensor'">
          <el-form-item label="指标" prop="metric_code">
            <el-select
              v-model="channelForm.metric_code"
              placeholder="请选择指标"
              style="width: 100%"
              @change="onChannelMetricChange"
            >
              <el-option
                v-for="m in metricDefs"
                :key="m.code"
                :label="`${m.name} (${m.unit})`"
                :value="m.code"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="单位">
            <el-input
              v-model="channelForm.unit"
              placeholder="选择指标后自动填充"
              disabled
            />
          </el-form-item>
          <el-form-item label="精度位数">
            <el-input-number
              v-model="channelForm.precision_digits"
              :min="0"
              :max="6"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="量程下限">
            <el-input-number
              v-model="channelForm.range_min"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="量程上限">
            <el-input-number
              v-model="channelForm.range_max"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="采样间隔(秒)">
            <el-input-number
              v-model="channelForm.sampling_interval_sec"
              :min="1"
              :max="86400"
              style="width: 100%"
            />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="类型" prop="actuator_type">
            <el-select
              v-model="channelForm.actuator_type"
              placeholder="请选择类型"
              style="width: 100%"
            >
              <el-option
                v-for="opt in actuatorTypeOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="额定功率(W)">
            <el-input-number
              v-model="channelForm.rated_power_watt"
              :min="0"
              style="width: 100%"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="channelDialogVisible = false">取消</el-button>
        <el-button
          v-if="canManage"
          type="primary"
          :loading="channelSubmitLoading"
          @click="handleChannelSubmit"
          >确定</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowLeft, Edit, Plus } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox, FormInstance, FormRules } from "element-plus";
import { deviceApi, telemetryApi, greenhouseApi, metricApi } from "@/api";
import { usePermission } from "@/composables/usePermission";
import { formatDate, formatNumber, getMetricName } from "@/utils/format";
import { actuatorTypeOptions } from "@/utils/device";
import { LARGE_PAGE_SIZE } from "@/utils/constants";
import { Role } from "@/types";
import type {
  SensorDevice,
  ActuatorDevice,
  SensorChannel,
  ActuatorChannel,
  TelemetryRecord,
  Greenhouse,
  GrowingZone,
  MetricDefinition,
} from "@/types";

const route = useRoute();
const router = useRouter();
const { canControlDevice, hasRole } = usePermission();
const canManage = computed(() => canControlDevice());
const canDelete = computed(() => hasRole(Role.ADMIN));

const deviceId = computed(() => Number(route.params.id));

const loading = ref(false);
const errorMsg = ref("");
const activeTab = ref("info");
const telemetryLoading = ref(false);

// Determine device type (sensor or actuator) from route query or by trial
const deviceType = ref<"sensor" | "actuator">(
  (route.query.type as "sensor" | "actuator") || "sensor",
);

type AnyDevice = SensorDevice | ActuatorDevice;
const device = ref<AnyDevice | null>(null);
const channels = ref<SensorChannel[] | ActuatorChannel[]>([]);
const telemetryByChannel = ref<Array<TelemetryRecord & { channel_id: number }>>(
  [],
);
const channelToggleLoading = reactive<Record<number, boolean>>({});
const greenhouses = ref<Greenhouse[]>([]);
const growingZones = ref<GrowingZone[]>([]);
const allZones = ref<GrowingZone[]>([]);

// Edit dialog state
const editDialogVisible = ref(false);
const editFormRef = ref<FormInstance>();
const editSubmitLoading = ref(false);
const editForm = reactive({
  name: "",
  model: "",
  firmware_version: "",
  greenhouse_id: null as number | null,
  growing_zone_id: undefined as number | undefined,
});
const editFormRules: FormRules = {
  name: [
    { required: true, message: "请输入设备名称", trigger: "blur" },
    { min: 1, max: 64, message: "设备名称长度为 1-64 个字符", trigger: "blur" },
  ],
  greenhouse_id: [{ required: true, message: "请选择温室", trigger: "change" }],
};

const filteredEditZones = computed(() => {
  if (!editForm.greenhouse_id) return allZones.value;
  return allZones.value.filter(
    (z) => z.greenhouse_id === editForm.greenhouse_id,
  );
});

// Channel dialog state
const channelDialogVisible = ref(false);
const isChannelEdit = ref(false);
const channelFormRef = ref<FormInstance>();
const channelSubmitLoading = ref(false);
const editingChannelId = ref<number | null>(null);
const metricDefs = ref<MetricDefinition[]>([]);

const channelForm = reactive({
  channel_code: "",
  // sensor fields
  metric_code: "",
  unit: "",
  precision_digits: undefined as number | undefined,
  range_min: undefined as number | undefined,
  range_max: undefined as number | undefined,
  sampling_interval_sec: undefined as number | undefined,
  // actuator fields
  actuator_type: "",
  rated_power_watt: undefined as number | undefined,
});

const channelFormRules: FormRules = {
  channel_code: [
    { required: true, message: "请输入通道编码", trigger: "blur" },
    { min: 1, max: 64, message: "通道编码长度为 1-64 个字符", trigger: "blur" },
  ],
};

// Channel unit lookup map
const channelUnitMap = computed(() => {
  const map = new Map<number, string>();
  for (const ch of channels.value) {
    if ("unit" in ch) {
      map.set((ch as SensorChannel).id, (ch as SensorChannel).unit);
    }
  }
  return map;
});

function getChannelUnit(channelId: number): string {
  return channelUnitMap.value.get(channelId) || "";
}

function getGreenhouseName(id: number): string {
  const gh = greenhouses.value.find((g) => g.id === id);
  return gh?.name || "-";
}

function getZoneName(id: number | undefined | null): string {
  if (!id) return "-";
  const zone = growingZones.value.find((z) => z.id === id);
  return zone?.name || "-";
}

function goBack() {
  router.push(
    deviceType.value === "actuator"
      ? "/assets/actuator-devices"
      : "/assets/sensor-devices",
  );
}

function openEditDialog() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  if (!device.value) return;
  editForm.name = device.value.name;
  editForm.model = device.value.model || "";
  editForm.firmware_version = device.value.firmware_version || "";
  editForm.greenhouse_id = device.value.greenhouse_id;
  editForm.growing_zone_id = device.value.growing_zone_id;
  editDialogVisible.value = true;
  editFormRef.value?.clearValidate();
}

async function handleEditSubmit() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  if (!editFormRef.value || !device.value) return;
  try {
    await editFormRef.value.validate();
  } catch {
    return;
  }

  editSubmitLoading.value = true;
  try {
    const payload = {
      name: editForm.name,
      model: editForm.model || undefined,
      firmware_version: editForm.firmware_version || undefined,
      greenhouse_id: editForm.greenhouse_id ?? undefined,
      growing_zone_id: editForm.growing_zone_id,
    };
    if (deviceType.value === "sensor") {
      await deviceApi.updateSensorDevice(deviceId.value, payload);
    } else {
      await deviceApi.updateActuatorDevice(deviceId.value, payload);
    }
    ElMessage.success("设备更新成功");
    editDialogVisible.value = false;
    await loadData();
  } catch {
    /* error handled */
  } finally {
    editSubmitLoading.value = false;
  }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE });
    metricDefs.value = data.items;
  } catch {
    /* ignore */
  }
}

function onChannelMetricChange() {
  const def = metricDefs.value.find((m) => m.code === channelForm.metric_code);
  if (def) channelForm.unit = def.unit;
}

function openChannelCreateDialog() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  isChannelEdit.value = false;
  editingChannelId.value = null;
  channelForm.channel_code = "";
  channelForm.metric_code = "";
  channelForm.unit = "";
  channelForm.precision_digits = undefined;
  channelForm.range_min = undefined;
  channelForm.range_max = undefined;
  channelForm.sampling_interval_sec = undefined;
  channelForm.actuator_type = "";
  channelForm.rated_power_watt = undefined;
  channelDialogVisible.value = true;
  channelFormRef.value?.clearValidate();
}

function openChannelEditDialog(ch: SensorChannel | ActuatorChannel) {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  isChannelEdit.value = true;
  editingChannelId.value = ch.id;
  channelForm.channel_code = ch.channel_code;
  if ("metric_code" in ch) {
    const sch = ch as SensorChannel;
    channelForm.metric_code = sch.metric_code;
    channelForm.unit = sch.unit;
    channelForm.precision_digits = sch.precision_digits;
    channelForm.range_min = sch.range_min;
    channelForm.range_max = sch.range_max;
    channelForm.sampling_interval_sec = sch.sampling_interval_sec;
    channelForm.actuator_type = "";
    channelForm.rated_power_watt = undefined;
  } else {
    const ach = ch as ActuatorChannel;
    channelForm.metric_code = "";
    channelForm.unit = "";
    channelForm.precision_digits = undefined;
    channelForm.range_min = undefined;
    channelForm.range_max = undefined;
    channelForm.sampling_interval_sec = undefined;
    channelForm.actuator_type = ach.actuator_type;
    channelForm.rated_power_watt = ach.rated_power_watt;
  }
  channelDialogVisible.value = true;
  channelFormRef.value?.clearValidate();
}

async function handleChannelSubmit() {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  if (!channelFormRef.value) return;
  try {
    await channelFormRef.value.validate();
  } catch {
    return;
  }

  channelSubmitLoading.value = true;
  try {
    if (deviceType.value === "sensor") {
      const payload = {
        sensor_device_id: deviceId.value,
        channel_code: channelForm.channel_code,
        metric_code: channelForm.metric_code,
        unit: channelForm.unit,
        precision_digits: channelForm.precision_digits,
        range_min: channelForm.range_min,
        range_max: channelForm.range_max,
        sampling_interval_sec: channelForm.sampling_interval_sec,
      };
      if (isChannelEdit.value && editingChannelId.value) {
        await deviceApi.updateSensorChannel(editingChannelId.value, payload);
        ElMessage.success("通道更新成功");
      } else {
        await deviceApi.createSensorChannel(payload);
        ElMessage.success("通道创建成功");
      }
    } else {
      const payload = {
        actuator_device_id: deviceId.value,
        channel_code: channelForm.channel_code,
        actuator_type: channelForm.actuator_type as any,
        rated_power_watt: channelForm.rated_power_watt,
      };
      if (isChannelEdit.value && editingChannelId.value) {
        await deviceApi.updateActuatorChannel(editingChannelId.value, payload);
        ElMessage.success("通道更新成功");
      } else {
        await deviceApi.createActuatorChannel(payload);
        ElMessage.success("通道创建成功");
      }
    }
    channelDialogVisible.value = false;
    // Reload channels
    await loadChannels();
    // Reload telemetry for sensor
    if (deviceType.value === "sensor") {
      fetchTelemetry();
    }
  } catch {
    /* error handled */
  } finally {
    channelSubmitLoading.value = false;
  }
}

async function handleChannelDelete(ch: SensorChannel | ActuatorChannel) {
  if (!canDelete.value) {
    ElMessage.error("没有权限执行此操作");
    return;
  }
  try {
    await ElMessageBox.confirm(
      `确认删除通道「${ch.channel_code}」？此操作不可撤销。`,
      "警告",
      {
        type: "warning",
        confirmButtonText: "确定",
        cancelButtonText: "取消",
      },
    );
    if (deviceType.value === "sensor") {
      await deviceApi.deleteSensorChannel(ch.id);
    } else {
      await deviceApi.deleteActuatorChannel(ch.id);
    }
    ElMessage.success("通道已删除");
    await loadChannels();
    if (deviceType.value === "sensor") {
      fetchTelemetry();
    }
  } catch (e: any) {
    if (e !== "cancel") {
      /* error handled */
    }
  }
}

async function toggleChannelEnabled(
  ch: SensorChannel | ActuatorChannel,
  enabled: boolean,
) {
  if (!canManage.value) {
    ElMessage.error("没有权限执行此操作");
    ch.enabled = !enabled;
    return;
  }
  channelToggleLoading[ch.id] = true;
  try {
    if (deviceType.value === "sensor") {
      await deviceApi.updateSensorChannel(ch.id, { enabled: enabled ? 1 : 0 });
    } else {
      await deviceApi.updateActuatorChannel(ch.id, {
        enabled: enabled ? 1 : 0,
      });
    }
    ElMessage.success(enabled ? "已启用" : "已停用");
  } catch {
    ch.enabled = !enabled;
  } finally {
    channelToggleLoading[ch.id] = false;
  }
}

async function loadChannels() {
  try {
    if (deviceType.value === "sensor") {
      const chData = await deviceApi.getSensorChannels({
        sensor_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE,
      });
      channels.value = chData.items;
    } else {
      const chData = await deviceApi.getActuatorChannels({
        actuator_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE,
      });
      channels.value = chData.items;
    }
  } catch {
    /* ignore */
  }
}

async function loadData() {
  if (!deviceId.value) {
    errorMsg.value = "无效的设备 ID";
    return;
  }
  loading.value = true;
  errorMsg.value = "";

  try {
    if (deviceType.value === "sensor") {
      device.value = await deviceApi.getSensorDevice(deviceId.value);
      const chData = await deviceApi.getSensorChannels({
        sensor_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE,
      });
      channels.value = chData.items;
    } else {
      device.value = await deviceApi.getActuatorDevice(deviceId.value);
      const chData = await deviceApi.getActuatorChannels({
        actuator_device_id: deviceId.value,
        page_size: LARGE_PAGE_SIZE,
      });
      channels.value = chData.items;
    }
  } catch (e: any) {
    // If first attempt fails, try the other device type
    if (deviceType.value === "sensor") {
      try {
        deviceType.value = "actuator";
        device.value = await deviceApi.getActuatorDevice(deviceId.value);
        const chData = await deviceApi.getActuatorChannels({
          actuator_device_id: deviceId.value,
          page_size: LARGE_PAGE_SIZE,
        });
        channels.value = chData.items;
      } catch {
        errorMsg.value = e?.message || "加载设备信息失败";
      }
    } else {
      errorMsg.value = e?.message || "加载设备信息失败";
    }
    return;
  } finally {
    loading.value = false;
  }

  // Load greenhouses and zones for display
  loadGreenhouses();
  loadAllZones();
  loadMetrics();
  if (device.value?.greenhouse_id) {
    loadGrowingZones(device.value.greenhouse_id);
  }

  // Load telemetry for sensor channels
  if (deviceType.value === "sensor") {
    fetchTelemetry();
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

async function loadGrowingZones(greenhouseId: number) {
  try {
    const data = await greenhouseApi.getGrowingZones({
      greenhouse_id: greenhouseId,
      page_size: LARGE_PAGE_SIZE,
    });
    growingZones.value = data.items;
  } catch {
    /* ignore */
  }
}

async function loadAllZones() {
  try {
    const data = await greenhouseApi.getGrowingZones({
      page_size: LARGE_PAGE_SIZE,
    });
    allZones.value = data.items;
  } catch {
    /* ignore */
  }
}

async function fetchTelemetry() {
  const sensorChs = channels.value as SensorChannel[];
  if (sensorChs.length === 0) return;

  telemetryLoading.value = true;
  try {
    const results: Array<TelemetryRecord & { channel_id: number }> = [];
    for (const ch of sensorChs) {
      try {
        const record = await telemetryApi.getChannelLatest(ch.id);
        results.push({ ...record, channel_id: ch.id });
      } catch {
        // Skip channels with no data
      }
    }
    telemetryByChannel.value = results;
  } finally {
    telemetryLoading.value = false;
  }
}

onMounted(() => {
  loadData();
});
</script>

<style scoped lang="scss">
.device-detail-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;
  }
  .page-title {
    font-size: 22px;
    font-weight: 700;
    color: var(--color-text-primary);
    margin: 0;
    flex: 1;
  }
  .loading-container,
  .error-container {
    padding: 40px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }
  .info-card,
  .channels-card {
    margin-bottom: 16px;
  }
  .channels-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .channels-empty {
    text-align: center;
    color: var(--color-text-secondary);
    padding: 24px 0;
    font-size: 14px;
  }
  .loading-placeholder,
  .empty-placeholder {
    padding: 20px;
    text-align: center;
    color: var(--color-text-secondary);
  }
  .telemetry-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }
  .telemetry-item {
    padding: 16px;
    background: var(--color-primary-bg-light);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color-light);
  }
  .metric-name {
    font-size: 14px;
    color: var(--color-text-regular);
    margin-bottom: 8px;
  }
  .metric-value {
    font-size: 24px;
    font-weight: 600;
    color: var(--color-text-primary);
    margin-bottom: 8px;
  }
  .metric-unit {
    font-size: 14px;
    font-weight: normal;
    color: var(--color-text-secondary);
  }
  .metric-meta {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .metric-time {
    font-size: 12px;
    color: var(--color-text-secondary);
  }
}
</style>
