import { get, post, put, del } from './request'
import type {
  SensorDevice,
  SensorDeviceListResponse,
  CreateSensorDeviceRequest,
  UpdateSensorDeviceRequest,
  SensorChannel,
  SensorChannelListResponse,
  CreateSensorChannelRequest,
  UpdateSensorChannelRequest,
  ActuatorDevice,
  ActuatorDeviceListResponse,
  CreateActuatorDeviceRequest,
  UpdateActuatorDeviceRequest,
  ActuatorChannel,
  ActuatorChannelListResponse,
  CreateActuatorChannelRequest,
  UpdateActuatorChannelRequest
} from '@/types'

// ===== Sensor Devices =====

export const getSensorDevices = (params?: Record<string, unknown>) =>
  get<SensorDeviceListResponse>('/sensor-devices', params)

export const getSensorDevice = (id: number) =>
  get<SensorDevice>(`/sensor-devices/${id}`)

export const createSensorDevice = (data: CreateSensorDeviceRequest) =>
  post<{ id: number }>('/sensor-devices', data)

export const updateSensorDevice = (id: number, data: UpdateSensorDeviceRequest) =>
  put<SensorDevice>(`/sensor-devices/${id}`, data)

export const deleteSensorDevice = (id: number) =>
  del<void>(`/sensor-devices/${id}`)

// ===== Sensor Channels =====

export const getSensorChannels = (params?: Record<string, unknown>) =>
  get<SensorChannelListResponse>('/sensor-channels', params)

export const getSensorChannel = (id: number) =>
  get<SensorChannel>(`/sensor-channels/${id}`)

export const createSensorChannel = (data: CreateSensorChannelRequest) =>
  post<{ id: number }>('/sensor-channels', data)

export const updateSensorChannel = (id: number, data: UpdateSensorChannelRequest) =>
  put<SensorChannel>(`/sensor-channels/${id}`, data)

export const deleteSensorChannel = (id: number) =>
  del<void>(`/sensor-channels/${id}`)

// ===== Actuator Devices =====

export const getActuatorDevices = (params?: Record<string, unknown>) =>
  get<ActuatorDeviceListResponse>('/actuator-devices', params)

export const getActuatorDevice = (id: number) =>
  get<ActuatorDevice>(`/actuator-devices/${id}`)

export const createActuatorDevice = (data: CreateActuatorDeviceRequest) =>
  post<{ id: number }>('/actuator-devices', data)

export const updateActuatorDevice = (id: number, data: UpdateActuatorDeviceRequest) =>
  put<ActuatorDevice>(`/actuator-devices/${id}`, data)

export const deleteActuatorDevice = (id: number) =>
  del<void>(`/actuator-devices/${id}`)

// ===== Actuator Channels =====

export const getActuatorChannels = (params?: Record<string, unknown>) =>
  get<ActuatorChannelListResponse>('/actuator-channels', params)

export const getActuatorChannel = (id: number) =>
  get<ActuatorChannel>(`/actuator-channels/${id}`)

export const createActuatorChannel = (data: CreateActuatorChannelRequest) =>
  post<{ id: number }>('/actuator-channels', data)

export const updateActuatorChannel = (id: number, data: UpdateActuatorChannelRequest) =>
  put<ActuatorChannel>(`/actuator-channels/${id}`, data)

export const deleteActuatorChannel = (id: number) =>
  del<void>(`/actuator-channels/${id}`)
