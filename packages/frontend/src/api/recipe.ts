import { del, get, post, put } from './request'
import type {
  CreateRecipeParams,
  NutrientRecipe,
  RecipeListResponse,
  RecipeQueryParams,
  RecipeTargetsResponse,
  UpdateRecipeParams,
  UpdateRecipeTargetsRequest
} from '@/types'

export function getRecipes(params?: RecipeQueryParams): Promise<RecipeListResponse> {
  return get<RecipeListResponse>('/recipes', params)
}

export function createRecipe(data: CreateRecipeParams): Promise<{ id: number }> {
  return post<{ id: number }>('/recipes', data)
}

export function getRecipeDetail(recipeId: number): Promise<NutrientRecipe> {
  return get<NutrientRecipe>(`/recipes/${recipeId}`)
}

export function updateRecipe(recipeId: number, data: UpdateRecipeParams): Promise<void> {
  return put<void>(`/recipes/${recipeId}`, data)
}

export function publishRecipe(recipeId: number, data: { version: string }): Promise<NutrientRecipe> {
  return post<NutrientRecipe>(`/recipes/${recipeId}/publish`, data)
}

export function deleteRecipe(recipeId: number): Promise<void> {
  return del<void>(`/recipes/${recipeId}`)
}

// ===== Targets (bulk replace via PUT) =====

export function getRecipeTargets(recipeId: number): Promise<RecipeTargetsResponse> {
  return get<RecipeTargetsResponse>(`/recipes/${recipeId}/targets`)
}

export function updateRecipeTargets(recipeId: number, data: UpdateRecipeTargetsRequest): Promise<void> {
  return put<void>(`/recipes/${recipeId}/targets`, data)
}
