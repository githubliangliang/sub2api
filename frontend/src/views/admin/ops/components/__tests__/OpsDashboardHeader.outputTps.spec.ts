import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import OpsDashboardHeader from '../OpsDashboardHeader.vue'
import type { OpsDashboardOverview } from '@/api/admin/ops'

vi.mock('@/api', () => ({ adminAPI: { groups: { getAll: vi.fn().mockResolvedValue([]) } } }))
vi.mock('@/api/admin/ops', () => ({ opsAPI: { getRealtimeTrafficSummary: vi.fn().mockResolvedValue({ enabled: true, summary: null }) } }))
vi.mock('@/stores', () => ({ useAdminSettingsStore: () => ({ opsRealtimeMonitoringEnabled: true, setOpsRealtimeMonitoringEnabledLocal: vi.fn() }) }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key })
}))
enableAutoUnmount(afterEach)

const overview = { tps: { avg: 242.7 } } as OpsDashboardOverview
function mountHeader(value = overview) {
  return mount(OpsDashboardHeader, {
    props: { overview: value, platform: '', groupId: null, timeRange: '1h', queryMode: 'auto', loading: false, lastUpdated: null },
    global: { stubs: { Select: true, HelpTooltip: true, BaseDialog: true, Icon: true } }
  })
}

describe('window output TPS', () => {
  it('shows the distribution and sample count separately from aggregate throughput', async () => {
    const wrapper = mountHeader({ ...overview, output_tps: { p5: 2.35, p10: 3.7, p50: 55, avg: 277.75, sample_count: 4 } })
    await flushPromises()
    const card = wrapper.get('[data-testid="output-tps-card"]')
    for (const [metric, value] of Object.entries({ p5: '2.4', p10: '3.7', p50: '55.0', avg: '277.8' })) {
      expect(card.get(`[data-testid="output-tps-${metric}"]`).text()).toBe(value)
    }
    expect(card.get('[data-testid="output-tps-samples"]').text()).toContain('4')
    expect(card.text()).toContain('tok/s (P50)')
    expect(card.text()).not.toContain('242.7')
  })

  it('uses refreshed overview values in fullscreen and preserves small valid rates', async () => {
    const wrapper = mountHeader({ ...overview, output_tps: { p5: 1, p10: 2, p50: 3, avg: 4, sample_count: 5 } })
    await flushPromises()
    await wrapper.setProps({ fullscreen: true, platform: 'grok', groupId: 123, timeRange: '24h', queryMode: 'preagg',
      overview: { ...overview, output_tps: { p5: 0.01, p10: 0.02, p50: 0.04, avg: 0.05, sample_count: 1 } } })
    expect(wrapper.get('[data-testid="output-tps-p50"]').text()).toBe('0.0')
    expect(wrapper.get('[data-testid="output-tps-samples"]').text()).toContain('1')
    await wrapper.setProps({ overview: { ...overview, output_tps: null } })
    expect(wrapper.get('[data-testid="output-tps-p50"]').text()).toBe('—')
  })

  it.each([undefined, null, { p5: null, p10: null, p50: null, avg: null, sample_count: 0 }])('shows missing rates without inventing zeroes: %s', async (output_tps) => {
    const wrapper = mountHeader({ ...overview, output_tps })
    await flushPromises()
    for (const metric of ['p5', 'p10', 'p50', 'avg']) {
      expect(wrapper.get(`[data-testid="output-tps-${metric}"]`).text()).toBe('—')
    }
    expect(wrapper.get('[data-testid="output-tps-samples"]').text()).toContain(output_tps == null ? '—' : '0')
  })
})
