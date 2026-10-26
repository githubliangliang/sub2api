const ACCOUNT_FILTER_STORAGE_KEY = 'account-filter-values'
const PLATFORMS = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
const STATUSES = ['active', 'inactive', 'error', 'rate_limited', 'temp_unschedulable', 'unschedulable']

interface AccountFilterValues {
  platform: string
  status: string
}

export function loadAccountFilters(): AccountFilterValues {
  const fallback = { platform: '', status: '' }
  try {
    const raw = localStorage.getItem(ACCOUNT_FILTER_STORAGE_KEY)
    if (!raw) return fallback
    const parsed: unknown = JSON.parse(raw)
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) return fallback
    const values = parsed as Record<string, unknown>
    return {
      platform: typeof values.platform === 'string' && PLATFORMS.includes(values.platform) ? values.platform : '',
      status: typeof values.status === 'string' && STATUSES.includes(values.status) ? values.status : ''
    }
  } catch {
    return fallback
  }
}

export function saveAccountFilters(filters: AccountFilterValues): void {
  try {
    localStorage.setItem(ACCOUNT_FILTER_STORAGE_KEY, JSON.stringify({ platform: filters.platform, status: filters.status }))
  } catch {
    // Filtering remains available when browser storage is unavailable.
  }
}
