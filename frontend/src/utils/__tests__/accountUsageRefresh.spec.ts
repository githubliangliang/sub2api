import { describe, expect, it } from 'vitest'
import { buildAccountUsageRefreshKey, buildOpenAIUsageRefreshKey } from '../accountUsageRefresh'

describe('buildOpenAIUsageRefreshKey', () => {
  it('会在 codex 快照变化时生成不同 key', () => {
    const base = {
      id: 1,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T09:59:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 0,
        codex_7d_used_percent: 0
      }
    } as any

    const next = {
      ...base,
      extra: {
        ...base.extra,
        codex_usage_updated_at: '2026-03-07T10:01:00Z',
        codex_5h_used_percent: 100
      }
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('会在 last_used_at 变化时生成不同 key', () => {
    const base = {
      id: 3,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 12,
        codex_7d_used_percent: 24
      }
    } as any

    const next = {
      ...base,
      last_used_at: '2026-03-07T10:02:00Z'
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('非 OpenAI OAuth 账号返回空 key', () => {
    expect(buildOpenAIUsageRefreshKey({
      id: 2,
      platform: 'anthropic',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {}
    } as any)).toBe('')
  })
})


describe('buildAccountUsageRefreshKey', () => {
  it.each(['oauth', 'setup-token'])('tracks every Claude %s passive snapshot input', (type) => {
    const base = { id: 1, platform: 'anthropic', type, extra: {} } as any
    for (const field of ['last_used_at', 'session_window_start', 'session_window_end', 'session_window_status']) {
      expect(buildAccountUsageRefreshKey({ ...base, [field]: 'new' })).not.toBe(buildAccountUsageRefreshKey(base))
    }
    for (const field of ['session_window_utilization', 'passive_usage_sampled_at', 'passive_usage_7d_utilization', 'passive_usage_7d_reset', 'passive_usage_7d_oi_utilization', 'passive_usage_7d_oi_reset']) {
      expect(buildAccountUsageRefreshKey({ ...base, extra: { [field]: 1 } })).not.toBe(buildAccountUsageRefreshKey(base))
    }
  })

  it.each(['anthropic', 'gemini'])('ignores unrelated %s changes and extra key ordering', (platform) => {
    const base = { id: 1, platform, type: 'oauth', extra: { session_window_utilization: 0, other: 1 } } as any
    expect(buildAccountUsageRefreshKey({ ...base, updated_at: 'new', current_concurrency: 5, extra: { other: 2, session_window_utilization: 0 } } as any)).toBe(buildAccountUsageRefreshKey(base))
    expect(buildAccountUsageRefreshKey({ ...base, last_used_at: 'new' })).not.toBe(buildAccountUsageRefreshKey(base))
  })

  it('retains OpenAI keys and excludes unsupported platforms/types', () => {
    const base = { id: 1, platform: 'openai', type: 'oauth', extra: {} } as any
    expect(buildAccountUsageRefreshKey(base)).toBe(buildOpenAIUsageRefreshKey(base))
    expect(buildAccountUsageRefreshKey({ ...base, platform: 'anthropic', type: 'apikey' })).toBe('')
    expect(buildAccountUsageRefreshKey({ ...base, platform: 'antigravity' })).toBe('')
  })
})
