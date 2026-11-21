import type { Account } from '@/types'

const normalizeUsageRefreshValue = (value: unknown): string => {
  if (value == null) return ''
  return String(value)
}

export const buildOpenAIUsageRefreshKey = (account: Pick<Account, 'id' | 'platform' | 'type' | 'updated_at' | 'last_used_at' | 'rate_limit_reset_at' | 'extra'>): string => {
  if (account.platform !== 'openai' || account.type !== 'oauth') {
    return ''
  }

  const extra = account.extra ?? {}
  return [
    account.id,
    account.updated_at,
    account.last_used_at,
    account.rate_limit_reset_at,
    extra.codex_usage_updated_at,
    extra.codex_5h_used_percent,
    extra.codex_5h_reset_at,
    extra.codex_5h_reset_after_seconds,
    extra.codex_5h_window_minutes,
    extra.codex_7d_used_percent,
    extra.codex_7d_reset_at,
    extra.codex_7d_reset_after_seconds,
    extra.codex_7d_window_minutes
  ].map(normalizeUsageRefreshValue).join('|')
}

export const buildAccountUsageRefreshKey = (account: Pick<Account, 'id' | 'platform' | 'type' | 'updated_at' | 'last_used_at' | 'rate_limit_reset_at' | 'session_window_start' | 'session_window_end' | 'session_window_status' | 'extra'>): string => {
  if (account.platform === 'openai') return buildOpenAIUsageRefreshKey(account)
  const isClaude = account.platform === 'anthropic' && (account.type === 'oauth' || account.type === 'setup-token')
  if (!isClaude && account.platform !== 'gemini') return ''
  const extra = account.extra ?? {}
  return JSON.stringify([
    account.id, account.platform, account.type, account.last_used_at ?? null,
    account.session_window_start ?? null, account.session_window_end ?? null,
    ...(isClaude ? [account.session_window_status ?? null, extra.session_window_utilization ?? null,
      extra.passive_usage_sampled_at ?? null, extra.passive_usage_7d_utilization ?? null,
      extra.passive_usage_7d_reset ?? null, extra.passive_usage_7d_oi_utilization ?? null,
      extra.passive_usage_7d_oi_reset ?? null] : [])
  ])
}
