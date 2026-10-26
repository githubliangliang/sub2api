import type { TempUnschedulableScope } from '@/types'

// Match the backend's legacy defaults so opening and saving old rules is neutral.
export function resolveTempUnschedScope(value: unknown, errorCode: number | null): TempUnschedulableScope {
  const scope = typeof value === 'string' ? value.trim().toLowerCase() : ''
  if (scope === 'account' || scope === 'model') return scope
  return errorCode === 401 ? 'account' : 'model'
}
