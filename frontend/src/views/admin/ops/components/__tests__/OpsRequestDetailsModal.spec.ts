import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { OpsDashboardOverview } from '@/api/admin/ops'
import { flushPromises, mount, shallowMount } from '@vue/test-utils'
import OpsDashboardHeader from '../OpsDashboardHeader.vue'
import OpsRequestDetailsModal from '../OpsRequestDetailsModal.vue'
import { clearAccountNameCache } from '../../utils/accountNameMap'
import { useAuthStore } from '@/stores/auth'

const { listRequestDetails, listAccounts, viewport } = vi.hoisted(() => ({
  listRequestDetails: vi.fn(),
  listAccounts: vi.fn(),
  viewport: { desktop: true },
}))

vi.mock('@vueuse/core', () => ({ useMediaQuery: () => ref(viewport.desktop) }))
vi.mock('@/api/admin/ops', () => ({ opsAPI: { listRequestDetails } }))
vi.mock('@/api/admin/accounts', () => ({ accountsAPI: { list: listAccounts } }))
vi.mock('@/stores/auth', async () => {
  const { reactive } = await import('vue')
  const state = reactive({ token: 'session-a' })
  return { useAuthStore: () => state }
})
vi.mock('@/api', () => ({ adminAPI: { groups: { getAll: vi.fn().mockResolvedValue([]) } } }))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError: vi.fn() }),
  useAdminSettingsStore: () => ({ opsRealtimeMonitoringEnabled: false }),
}))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key }),
}))

async function openDetails(sort: 'duration_desc' | 'ttft_desc') {
  const wrapper = mount(OpsRequestDetailsModal, {
    props: { modelValue: false, timeRange: '1h', preset: { title: 'Details', sort } },
    global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, Pagination: true } },
  })
  await wrapper.setProps({ modelValue: true })
  await flushPromises()
  return wrapper
}

describe('Ops request latency details', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    clearAccountNameCache()
    useAuthStore().token = 'session-a'
    listAccounts.mockReset().mockResolvedValue({ items: [], total: 0 })
    viewport.desktop = true
    listRequestDetails.mockResolvedValue({
      items: [
        { kind: 'success', created_at: '2026-09-10T00:00:00Z', duration_ms: 12000, first_token_ms: 800 },
        { kind: 'success', created_at: '2026-09-10T00:00:01Z', duration_ms: 9000, first_token_ms: 0 },
        { kind: 'success', created_at: '2026-09-10T00:00:02Z', duration_ms: 5000, first_token_ms: null },
      ],
      total: 3,
    })
  })

  it('opens the TTFT card with first-token sorting and successful requests', async () => {
    const wrapper = shallowMount(OpsDashboardHeader, {
      props: { overview: {} as OpsDashboardOverview, platform: '', groupId: null, timeRange: '1h', queryMode: 'auto', loading: false, lastUpdated: null },
    })
    await flushPromises()
    const button = wrapper.findAll('button').find((item) =>
      item.text() === 'admin.ops.requestDetails.details' &&
      item.element.parentElement?.textContent?.includes('TTFT'),
    )
    expect(button).toBeDefined()
    await button!.trigger('click')
    expect(wrapper.emitted('openRequestDetails')).toEqual([[
      { title: 'admin.ops.ttftLabel', kind: 'success', sort: 'ttft_desc' },
    ]])
    wrapper.unmount()
  })

  it.each([true, false])('shows TTFT rather than total duration (desktop: %s)', async (desktop) => {
    viewport.desktop = desktop
    const wrapper = await openDetails('ttft_desc')
    expect(listRequestDetails).toHaveBeenCalledWith(expect.objectContaining({ sort: 'ttft_desc' }))
    expect(wrapper.text()).toContain('admin.ops.ttftLabel')
    expect(wrapper.text()).toContain('800 ms')
    expect(wrapper.text()).toContain('0 ms')
    expect(wrapper.text()).not.toContain('12000 ms')
    expect(wrapper.text()).not.toContain('9000 ms')
    expect(wrapper.text()).not.toContain('5000 ms')
    if (desktop) expect(wrapper.findAll('tbody tr')[2].findAll('td')[5].text()).toBe('-')
    else expect(wrapper.text()).toContain('admin.ops.ttftLabel: -')
    wrapper.unmount()
  })

  it('keeps total duration for duration details', async () => {
    const wrapper = await openDetails('duration_desc')
    expect(listRequestDetails).toHaveBeenCalledWith(expect.objectContaining({ sort: 'duration_desc' }))
    expect(wrapper.text()).toContain('admin.ops.requestDetails.table.duration')
    expect(wrapper.text()).toContain('12000 ms')
    expect(wrapper.text()).not.toContain('800 ms')
    wrapper.unmount()
  })

  it.each([true, false])('shows account names with id fallback without blocking details (desktop: %s)', async (desktop) => {
    viewport.desktop = desktop
    listRequestDetails.mockResolvedValue({ items: [
      { kind: 'success', account_id: 101, duration_ms: 100, model: 'text-model' },
      { kind: 'error', account_id: 999, duration_ms: 200, model: 'text-model' },
      { kind: 'success', account_id: null, duration_ms: 300, model: 'text-model' }
    ], total: 3 })
    let finish!: (value: unknown) => void
    listAccounts.mockReturnValueOnce(new Promise(resolve => { finish = resolve }))
      .mockResolvedValue({ items: [{ id: 101, name: 'account-alpha' }], total: 1 })
    const wrapper = await openDetails('duration_desc')
    expect(wrapper.text()).toContain('text-model')
    expect(wrapper.text()).toContain('#101')
    finish({ items: [{ id: 101, name: 'account-alpha' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).toContain('account-alpha')
    expect(wrapper.text()).toContain('#999')
    expect(listAccounts).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('refreshes account names on manual refresh and retains details when lookup fails', async () => {
    listRequestDetails.mockResolvedValue({ items: [{ account_id: 101, model: 'text-model' }], total: 1 })
    listAccounts.mockResolvedValueOnce({ items: [{ id: 101, name: 'old-name' }], total: 1 })
      .mockResolvedValueOnce({ items: [{ id: 101, name: 'new-name' }], total: 1 })
      .mockRejectedValueOnce(new Error('offline'))
    const wrapper = await openDetails('duration_desc')
    expect(wrapper.text()).toContain('old-name')
    const refresh = wrapper.findAll('button').find(button => button.text() === 'common.refresh')!
    await refresh.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('new-name')
    await refresh.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('text-model')
    wrapper.unmount()
  })

  it('replaces an older forced name snapshot on every explicit manual refresh', async () => {
    let finishOldRefresh!: (value: unknown) => void
    listRequestDetails.mockResolvedValue({ items: [{ account_id: 101, model: 'text-model' }], total: 1 })
    listAccounts.mockResolvedValueOnce({ items: [{ id: 101, name: 'initial-name' }], total: 1 })
      .mockReturnValueOnce(new Promise(resolve => { finishOldRefresh = resolve }))
      .mockResolvedValueOnce({ items: [{ id: 101, name: 'renamed-again' }], total: 1 })
    const wrapper = await openDetails('duration_desc')
    const refresh = wrapper.findAll('button').find(button => button.text() === 'common.refresh')!
    await refresh.trigger('click')
    await flushPromises()
    expect(listAccounts).toHaveBeenCalledTimes(2)
    const oldSignal = listAccounts.mock.calls[1][3].signal as AbortSignal
    await refresh.trigger('click')
    await flushPromises()
    expect(listAccounts).toHaveBeenCalledTimes(3)
    expect(oldSignal.aborted).toBe(true)
    expect(wrapper.text()).toContain('renamed-again')
    finishOldRefresh({ items: [{ id: 101, name: 'old-refresh-name' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).not.toContain('old-refresh-name')
    expect(wrapper.text()).toContain('renamed-again')
    wrapper.unmount()
  })

  it('does not apply stale details or names after paging, closing or session changes', async () => {
    let finishNames!: (value: unknown) => void
    listRequestDetails.mockResolvedValue({ items: [{ account_id: 101, model: 'old-session-model' }], total: 20 })
    listAccounts.mockReturnValueOnce(new Promise(resolve => { finishNames = resolve }))
    const wrapper = await openDetails('duration_desc')
    listRequestDetails.mockResolvedValue({ items: [{ account_id: 202, model: 'new-session-model' }], total: 1 })
    listAccounts.mockResolvedValue({ items: [{ id: 202, name: 'new-session-name' }], total: 1 })
    useAuthStore().token = 'session-b'
    await flushPromises()
    finishNames({ items: [{ id: 101, name: 'old-session-name' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).toContain('new-session-name')
    expect(wrapper.text()).not.toContain('old-session-name')
    await wrapper.setProps({ modelValue: false })
    wrapper.unmount()
  })

  it('ignores a slower earlier page response after the filter context changes', async () => {
    let finishEarlierPage!: (value: unknown) => void
    listRequestDetails.mockResolvedValueOnce({ items: [{ model: 'page-one' }], total: 30 })
      .mockReturnValueOnce(new Promise(resolve => { finishEarlierPage = resolve }))
      .mockResolvedValueOnce({ items: [{ model: 'latest-page' }], total: 30 })
    const wrapper = await openDetails('duration_desc')
    const pagination = wrapper.findComponent({ name: 'Pagination' })
    pagination.vm.$emit('update:page', 2)
    await flushPromises()
    await wrapper.setProps({ timeRange: '24h' })
    await flushPromises()
    finishEarlierPage({ items: [{ model: 'stale-page' }], total: 30 })
    await flushPromises()
    expect(wrapper.text()).toContain('latest-page')
    expect(wrapper.text()).not.toContain('stale-page')
    wrapper.unmount()
  })

  it('does not write late names into a closed dialog and shares their pending lookup on reopening', async () => {
    let finishNames!: (value: unknown) => void
    listRequestDetails.mockResolvedValue({ items: [{ account_id: 101, model: 'text-model' }], total: 1 })
    listAccounts.mockReturnValue(new Promise(resolve => { finishNames = resolve }))
    const wrapper = await openDetails('duration_desc')
    await wrapper.setProps({ modelValue: false })
    await wrapper.setProps({ modelValue: true })
    await flushPromises()
    expect(listAccounts).toHaveBeenCalledTimes(1)
    await wrapper.setProps({ modelValue: false })
    finishNames({ items: [{ id: 101, name: 'late-name' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).not.toContain('late-name')
    wrapper.unmount()
  })

  it('resumes pending account names when returning from error details without reloading the saved list', async () => {
    let finishNames!: (value: unknown) => void
    listRequestDetails.mockResolvedValue({ items: [{ kind: 'error', account_id: 101, model: 'saved-context' }], total: 1 })
    listAccounts.mockReturnValue(new Promise(resolve => { finishNames = resolve }))
    const wrapper = await openDetails('duration_desc')
    await wrapper.setProps({ modelValue: false })
    await wrapper.setProps({ modelValue: true, resumeState: true })
    expect(listRequestDetails).toHaveBeenCalledTimes(1)
    finishNames({ items: [{ id: 101, name: 'resumed-account' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).toContain('saved-context')
    expect(wrapper.text()).toContain('resumed-account')
    wrapper.unmount()
  })
})
