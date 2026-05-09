import type { PaginatedResponse } from './api'

export type RecipeStatus = 'DRAFT' | 'ACTIVE' | 'ARCHIVED'
export type BatchRecipeBindingType = 'PRIMARY' | 'SECONDARY'
export type BatchRecipeBindingStatus = 'ACTIVE' | 'INACTIVE'

export interface NutrientRecipe {
  id: number
  recipe_code: string
  name: string
  crop_variety_id?: number | null
  description?: string
  version: string
  status: RecipeStatus
  effective_from?: string | null
  effective_to?: string | null
  created_by?: number | null
  published_by?: number | null
  published_at?: string | null
  created_at: string
  updated_at: string
}

export interface RecipeStageTarget {
  id: number
  recipe_id: number
  growth_stage_id?: number | null
  metric_code: string
  target_min?: number | null
  target_max?: number | null
  tolerance?: number | null
  unit?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface RecipeIonTarget {
  id: number
  recipe_id: number
  growth_stage_id?: number | null
  ion_code: string
  target_min_mg_l?: number | null
  target_max_mg_l?: number | null
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface RecipeTargetsResponse {
  recipe: NutrientRecipe
  stage_targets: RecipeStageTarget[]
  ion_targets: RecipeIonTarget[]
}

export interface BatchRecipeBinding {
  id: number
  batch_id: number
  recipe_id: number
  binding_type: BatchRecipeBindingType
  version: string
  effective_from: string
  effective_to?: string | null
  status: BatchRecipeBindingStatus
}

export interface RecipeQueryParams {
  page?: number
  page_size?: number
  status?: RecipeStatus
}

export interface CreateRecipeParams {
  recipe_code: string
  name: string
  crop_variety_id?: number
  description?: string
  version?: string
  status?: RecipeStatus
  effective_from?: string
  effective_to?: string
  stage_targets?: CreateStageTargetParams[]
  ion_targets?: CreateIonTargetParams[]
}

export interface CreateStageTargetParams {
  id?: number
  growth_stage_id?: number | null
  metric_code: string
  target_min?: number | null
  target_max?: number | null
  tolerance?: number | null
  unit?: string
  enabled?: boolean
}

export interface CreateIonTargetParams {
  id?: number
  growth_stage_id?: number | null
  ion_code: string
  target_min_mg_l?: number | null
  target_max_mg_l?: number | null
  enabled?: boolean
}

export interface UpdateRecipeTargetsRequest {
  stage_targets: CreateStageTargetParams[]
  ion_targets: CreateIonTargetParams[]
}

export type UpdateRecipeParams = Partial<
  Pick<CreateRecipeParams, 'name' | 'crop_variety_id' | 'description' | 'version' | 'status' | 'effective_from' | 'effective_to'>
>

export type RecipeListResponse = PaginatedResponse<NutrientRecipe>
