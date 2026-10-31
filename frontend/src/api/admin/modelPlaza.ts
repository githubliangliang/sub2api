import { apiClient } from '../client'
import type { PlazaOfficialPricing } from '../modelPlaza'

export interface AccountModelPlazaItem {
  client_id: string
  platform: string
  upstream_ids: string[]
  account_count: number
  pricing: (PlazaOfficialPricing & { model: string }) | null
}

export interface AccountModelPlazaResponse {
  items: AccountModelPlazaItem[]
}

export async function listAccountModels(): Promise<AccountModelPlazaResponse> {
  const { data } = await apiClient.get<AccountModelPlazaResponse>('/admin/model-plaza')
  return data ?? { items: [] }
}

export const modelPlazaAPI = {
  listAccountModels
}

export default modelPlazaAPI
