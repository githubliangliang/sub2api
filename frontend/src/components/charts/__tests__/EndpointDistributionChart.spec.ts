import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import EndpointDistributionChart from '../EndpointDistributionChart.vue'

vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))
vi.mock('vue-chartjs', () => ({
  Doughnut: { props: ['data'], template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>' }
}))

const endpointStats = [{ endpoint: '/v1/messages', requests: 4, total_tokens: 1000, cost: 17, actual_cost: 4.1, account_cost: 2.9 }]

describe('EndpointDistributionChart cost dimensions', () => {
  it('keeps actual deduction, account cost and standard price in separate columns', () => {
    const wrapper = mount(EndpointDistributionChart, { props: { endpointStats, metric: 'actual_cost', enableBreakdown: false } })
    const cells = wrapper.find('tbody tr').findAll('td')
    expect(cells.map(cell => cell.text())).toEqual(['/v1/messages', '4', '1.00K', '$4.10', '$2.90', '$17.00'])
    const chart = JSON.parse(wrapper.find('.chart-data').text())
    expect(chart.datasets[0].data).toEqual([4.1])
    wrapper.unmount()
  })

  it('hides account cost on a user surface even if a stale response contains it', () => {
    const wrapper = mount(EndpointDistributionChart, { props: { endpointStats, showAccountCost: false, enableBreakdown: false } })
    expect(wrapper.findAll('thead th')).toHaveLength(5)
    expect(wrapper.find('tbody tr').findAll('td')).toHaveLength(5)
    expect(wrapper.text()).not.toContain('admin.dashboard.accountCost')
    expect(wrapper.text()).not.toContain('$2.90')
    wrapper.unmount()
  })
})
