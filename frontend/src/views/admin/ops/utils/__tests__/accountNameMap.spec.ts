import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { clearAccountNameCache, getAccountNameMap } from '../accountNameMap'

const { listAccounts } = vi.hoisted(() => ({ listAccounts: vi.fn() }))
vi.mock('@/api/admin/accounts', () => ({ accountsAPI: { list: listAccounts } }))

describe('ops account name lookup', () => {
  beforeEach(() => { clearAccountNameCache(); listAccounts.mockReset() })
  afterEach(() => vi.useRealTimers())

  it('paginates lite records and shares current in-flight work', async () => {
    let finish!: (value: unknown) => void
    listAccounts.mockReturnValueOnce(new Promise(resolve => { finish = resolve }))
      .mockResolvedValueOnce({ items: [{ id: 2, name: 'second' }], total: 2 })
    const first = getAccountNameMap('session-a')
    const second = getAccountNameMap('session-a')
    expect(listAccounts).toHaveBeenCalledTimes(1)
    finish({ items: [{ id: 1, name: 'first' }], total: 2 })
    const [a, b] = await Promise.all([first, second])
    expect(a).toBe(b)
    expect([...a]).toEqual([[1, 'first'], [2, 'second']])
    expect(listAccounts).toHaveBeenCalledWith(1, 1000, { lite: '1', include_scheduler_score: '0' }, expect.objectContaining({ signal: expect.any(AbortSignal) }))
    expect(listAccounts).toHaveBeenCalledTimes(2)
    expect(await getAccountNameMap('session-a')).toBe(a)
  })

  it('refreshes renamed accounts on force and expires cached values', async () => {
    vi.useFakeTimers()
    listAccounts.mockResolvedValueOnce({ items: [{ id: 1, name: 'old' }], total: 1 })
      .mockResolvedValueOnce({ items: [{ id: 1, name: 'renamed' }], total: 1 })
      .mockResolvedValueOnce({ items: [{ id: 1, name: 'latest' }], total: 1 })
    expect((await getAccountNameMap('a')).get(1)).toBe('old')
    expect((await getAccountNameMap('a', true)).get(1)).toBe('renamed')
    await vi.advanceTimersByTimeAsync(60001)
    expect((await getAccountNameMap('a')).get(1)).toBe('latest')
  })

  it('replaces an older pending lookup when a manual refresh requests current names', async () => {
    let finishOld!: (value: unknown) => void
    listAccounts.mockReturnValueOnce(new Promise(resolve => { finishOld = resolve }))
      .mockResolvedValueOnce({ items: [{ id: 1, name: 'renamed' }], total: 1 })
    const oldLookup = getAccountNameMap('a')
    const signal = listAccounts.mock.calls[0][3].signal as AbortSignal
    const manualRefresh = getAccountNameMap('a', true)
    expect(listAccounts).toHaveBeenCalledTimes(2)
    expect(signal.aborted).toBe(true)
    expect((await manualRefresh).get(1)).toBe('renamed')
    finishOld({ items: [{ id: 1, name: 'old-name' }], total: 1 })
    await oldLookup
    expect((await getAccountNameMap('a')).get(1)).toBe('renamed')
  })

  it('terminates empty and repeated result pages without infinite requests', async () => {
    listAccounts.mockResolvedValue({ items: [{ id: 1, name: 'same' }], total: 9000 })
    expect((await getAccountNameMap('a')).size).toBe(1)
    expect(listAccounts).toHaveBeenCalledTimes(2)
    listAccounts.mockReset().mockResolvedValue({ items: [], total: 2000 })
    expect((await getAccountNameMap('a', true)).size).toBe(0)
    expect(listAccounts).toHaveBeenCalledTimes(1)
  })

  it('does not retain failed requests as successful empty caches', async () => {
    listAccounts.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ items: [{ id: 1, name: 'ok' }], total: 1 })
    await expect(getAccountNameMap('a')).rejects.toThrow('offline')
    expect((await getAccountNameMap('a')).get(1)).toBe('ok')
  })

  it('isolates sessions and ignores a previous session response arriving later', async () => {
    let finish!: (value: unknown) => void
    listAccounts.mockReturnValueOnce(new Promise(resolve => { finish = resolve }))
      .mockResolvedValueOnce({ items: [{ id: 2, name: 'new-session' }], total: 1 })
    const previous = getAccountNameMap('old-session')
    const signal = listAccounts.mock.calls[0][3].signal as AbortSignal
    expect((await getAccountNameMap('new-session')).get(2)).toBe('new-session')
    expect(signal.aborted).toBe(true)
    finish({ items: [{ id: 1, name: 'old-session' }], total: 1 })
    await previous
    expect((await getAccountNameMap('new-session')).has(1)).toBe(false)
    expect(listAccounts).toHaveBeenCalledTimes(2)
    expect((await getAccountNameMap('')).size).toBe(0)
  })
})
