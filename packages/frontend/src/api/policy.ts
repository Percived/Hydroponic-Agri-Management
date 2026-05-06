import { get, post, put, del } from './request'
import type {
  ControlPolicy,
  ControlPolicyListResponse,
  CreatePolicyRequest,
  PolicyCondition,
  CreateConditionRequest,
  PolicyTarget,
  CreateTargetRequest,
  PolicyExecution,
  PolicyExecutionListResponse
} from '@/types'

// ===== Policies =====

export const getPolicies = (params?: Record<string, unknown>) =>
  get<ControlPolicyListResponse>('/policies', params)

export const getPolicy = (id: number) =>
  get<ControlPolicy>(`/policies/${id}`)

export const createPolicy = (data: CreatePolicyRequest) =>
  post<{ id: number }>('/policies', data)

export const updatePolicy = (id: number, data: Partial<CreatePolicyRequest>) =>
  put<ControlPolicy>(`/policies/${id}`, data)

export const deletePolicy = (id: number) =>
  del<void>(`/policies/${id}`)

export const publishPolicy = (id: number) =>
  post<ControlPolicy>(`/policies/${id}/publish`)

export const archivePolicy = (id: number) =>
  post<ControlPolicy>(`/policies/${id}/archive`)

// ===== Policy Conditions =====

export const getPolicyConditions = (policyId: number) =>
  get<{ items: PolicyCondition[] }>(`/policies/${policyId}/conditions`)

export const createPolicyCondition = (policyId: number, data: CreateConditionRequest) =>
  post<{ id: number }>(`/policies/${policyId}/conditions`, data)

export const updatePolicyCondition = (policyId: number, conditionId: number, data: Partial<CreateConditionRequest>) =>
  put<PolicyCondition>(`/policies/${policyId}/conditions/${conditionId}`, data)

export const deletePolicyCondition = (policyId: number, conditionId: number) =>
  del<void>(`/policies/${policyId}/conditions/${conditionId}`)

// ===== Policy Targets =====

export const getPolicyTargets = (policyId: number) =>
  get<{ items: PolicyTarget[] }>(`/policies/${policyId}/targets`)

export const createPolicyTarget = (policyId: number, data: CreateTargetRequest) =>
  post<{ id: number }>(`/policies/${policyId}/targets`, data)

export const updatePolicyTarget = (policyId: number, targetId: number, data: Partial<CreateTargetRequest>) =>
  put<PolicyTarget>(`/policies/${policyId}/targets/${targetId}`, data)

export const deletePolicyTarget = (policyId: number, targetId: number) =>
  del<void>(`/policies/${policyId}/targets/${targetId}`)

// ===== Policy Executions =====

export const executePolicy = (policyId: number) =>
  post<PolicyExecution>(`/policies/${policyId}/execute`)

export const getPolicyExecutions = (params?: Record<string, unknown>) =>
  get<PolicyExecutionListResponse>('/policy-executions', params)
