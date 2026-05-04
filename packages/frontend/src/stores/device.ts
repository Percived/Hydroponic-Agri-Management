import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getDevices, getDevice, createDevice, updateDevice, updateDeviceStatus, getDeviceHealth } from '@/api/device'
import { getDeviceGroups } from '@/api/device-group'
import { Device, DeviceGroup, DeviceQueryParams, DeviceFormData, DeviceHealth, PaginatedData } from '@/types'

export const useDeviceStore = defineStore('device', () => {
  // 状态
  const devices = ref<Device[]>([])
  const currentDevice = ref<Device | null>(null)
  const deviceGroups = ref<DeviceGroup[]>([])
  const total = ref(0)
  const loading = ref(false)

  // 获取设备列表
  async function fetchDevices(params: DeviceQueryParams): Promise<PaginatedData<Device>> {
    loading.value = true
    try {
      const result = await getDevices(params)
      devices.value = result.items
      total.value = result.total
      return result
    } finally {
      loading.value = false
    }
  }

  // 获取设备详情
  async function fetchDevice(id: number): Promise<Device> {
    loading.value = true
    try {
      const result = await getDevice(id)
      currentDevice.value = result
      return result
    } finally {
      loading.value = false
    }
  }

  // 获取设备健康状态
  async function fetchDeviceHealth(id: number): Promise<DeviceHealth> {
    return getDeviceHealth(id)
  }

  // 新增设备
  async function addDevice(data: DeviceFormData): Promise<number> {
    const result = await createDevice(data)
    return result.id
  }

  // 更新设备
  async function editDevice(id: number, data: Partial<DeviceFormData>): Promise<void> {
    await updateDevice(id, data)
    if (currentDevice.value?.id === id) {
      currentDevice.value = { ...currentDevice.value, ...data } as Device
    }
  }

  // 更新设备状态
  async function setDeviceStatus(id: number, status: string): Promise<void> {
    await updateDeviceStatus(id, status)
    const device = devices.value.find((d) => d.id === id)
    if (device) {
      device.status = status as Device['status']
    }
    if (currentDevice.value?.id === id) {
      currentDevice.value.status = status as Device['status']
    }
  }

  // 获取设备分组列表
  async function fetchDeviceGroups(greenhouseId?: number): Promise<DeviceGroup[]> {
    const result = await getDeviceGroups(greenhouseId)
    deviceGroups.value = result.items
    return result.items
  }

  // 清除当前设备
  function clearCurrentDevice() {
    currentDevice.value = null
  }

  return {
    devices,
    currentDevice,
    deviceGroups,
    total,
    loading,
    fetchDevices,
    fetchDevice,
    fetchDeviceHealth,
    addDevice,
    editDevice,
    setDeviceStatus,
    fetchDeviceGroups,
    clearCurrentDevice
  }
})