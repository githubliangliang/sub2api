import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardStats from '../UserDashboardStats.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string, params?: { count?: number }) =>
    key === 'dashboard.platformCount' ? `platform-count:${params?.count}` : key })
}))

const stats = (overrides = {}) => ({ total_actual_cost: 0, today_actual_cost: 0, by_platform: [], ...overrides } as any)
const quota = (platform: string, daily: number | null = null) => ({
  platform, daily_limit_usd: daily, weekly_limit_usd: null, monthly_limit_usd: null,
  daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0
} as any)

describe('UserDashboardStats configured platform cards', () => {
  it('does not create platform cards from unlimited quota placeholders', () => {
    const wrapper = mount(UserDashboardStats, {
      props: { stats: stats(), balance: 0, isSimple: false, platformQuotas: [quota('openai'), quota('gemini')] }
    })
    expect(wrapper.text()).not.toContain('dashboard.platformBreakdown')
    wrapper.unmount()
  })

  it('counts the union of usage and configured limits without counting the other card', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: stats({ total_actual_cost: 2, today_actual_cost: 2, by_platform: [
          { platform: 'openai', total_actual_cost: 1, today_actual_cost: 1, total_requests: 1, total_tokens: 1 }
        ] }),
        balance: 0, isSimple: false,
        platformQuotas: [quota('anthropic', 0), quota('gemini')]
      }
    })
    expect(wrapper.text()).toContain('platform-count:2')
    expect(wrapper.text()).toContain('Claude')
    expect(wrapper.text()).toContain('OpenAI')
    expect(wrapper.text()).toContain('dashboard.platformOther')
    expect(wrapper.text()).not.toContain('Gemini')
    expect(wrapper.text()).toContain('dashboard.platformQuota.disabled')
    wrapper.unmount()
  })

  it('keeps the simple-mode platform section hidden', () => {
    const wrapper = mount(UserDashboardStats, {
      props: { stats: stats(), balance: 0, isSimple: true, platformQuotas: [quota('openai', 10)] }
    })
    expect(wrapper.text()).not.toContain('dashboard.platformBreakdown')
    wrapper.unmount()
  })
})
