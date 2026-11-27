import { apiClient } from '../client'
import type { PlazaOfficialPricing } from '../modelPlaza'

export interface AccountModelPlazaItem {
  client_id: string
  platform: string
  upstream_ids: string[]
  account_count: number
  pricing: (PlazaOfficialPricing & { model: string; manual?: boolean }) | null
}

export interface AccountModelPlazaResponse {
  items: AccountModelPlazaItem[]
}

export async function listAccountModels(): Promise<AccountModelPlazaResponse> {
  const { data } = await apiClient.get<AccountModelPlazaResponse>('/admin/model-plaza')
  return data ?? { items: [] }
}

export interface UpdateReferencePricingRequest {
  client_id: string
  platform: string
  pricing: PlazaOfficialPricing | null
}

export async function updateReferencePricing(request: UpdateReferencePricingRequest): Promise<void> {
  await apiClient.put('/admin/model-plaza/pricing', request)
}

export const modelPlazaAPI = {
  listAccountModels,
  updateReferencePricing
}

export default modelPlazaAPI
