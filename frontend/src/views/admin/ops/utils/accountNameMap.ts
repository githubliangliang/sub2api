import { accountsAPI } from '@/api/admin/accounts'

const ACCOUNT_NAME_PAGE_SIZE = 1000
const ACCOUNT_NAME_CACHE_TTL_MS = 60_000
let session = ''
let generation = 0
let cached: { names: Map<number, string>; expiresAt: number } | null = null
let pending: { promise: Promise<Map<number, string>>; controller: AbortController; forced: boolean } | null = null

export function clearAccountNameCache(): void {
  generation++
  pending?.controller.abort()
  pending = null
  cached = null
  session = ''
}

export function getAccountNameMap(sessionKey: string, force = false): Promise<Map<number, string>> {
  if (session !== sessionKey) {
    clearAccountNameCache()
    session = sessionKey
  }
  if (!sessionKey) return Promise.resolve(new Map())
  if (pending && (!force || pending.forced)) return pending.promise
  if (pending) {
    // A manual refresh must not reuse a snapshot that was already loading before it.
    generation++
    pending.controller.abort()
    pending = null
  }
  if (!force && cached && cached.expiresAt > Date.now()) return Promise.resolve(cached.names)

  const requestGeneration = generation
  const controller = new AbortController()
  const promise = (async () => {
    const names = new Map<number, string>()
    // Bound even a broken upstream pagination response; missing names keep their ID fallback.
    for (let page = 1; page <= 1000; page++) {
      const res = await accountsAPI.list(page, ACCOUNT_NAME_PAGE_SIZE, {
        lite: '1', include_scheduler_score: '0'
      }, { signal: controller.signal })
      if (requestGeneration !== generation) return names
      const previousSize = names.size
      for (const item of res.items || []) {
        if (Number.isInteger(item.id) && item.id > 0 && typeof item.name === 'string' && item.name.trim()) {
          names.set(item.id, item.name)
        }
      }
      if (names.size === previousSize || !Number.isFinite(res.total) || names.size >= res.total) break
    }
    if (requestGeneration === generation) {
      cached = { names, expiresAt: Date.now() + ACCOUNT_NAME_CACHE_TTL_MS }
    }
    return names
  })()
  pending = { promise, controller, forced: force }
  void promise.finally(() => {
    if (pending?.promise === promise) pending = null
  }).catch(() => {})
  return promise
}
