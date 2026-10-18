import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRPMOverridesModal from '../GroupRPMOverridesModal.vue'

const mocks = vi.hoisted(() => ({ getGroupRPMOverrides: vi.fn(), batchSetGroupRPMOverrides: vi.fn(), listUsers: vi.fn(), clearGroupRPMOverrides: vi.fn(), showError: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: mocks, users: { list: mocks.listUsers } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: mocks.showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
beforeEach(() => vi.resetAllMocks())
afterEach(() => vi.restoreAllMocks())

function deferred() {
  let resolve!: (value: unknown) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

const group = (id: number) => ({ id, name: `Group ${id}`, platform: 'openai', rpm_limit: 100 } as AdminGroup)
const entries = (id: number) => [{ user_id: id, user_name: '', user_email: `user${id}@example.com`, user_status: 'active', rpm_override: 100 }]

async function openGroup() {
  const wrapper = mount(GroupRPMOverridesModal, {
    props: { show: false, group: group(1) },
    global: { stubs: {
      BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
      Icon: true, PlatformIcon: true, Pagination: true
    } }
  })
  await wrapper.setProps({ show: true })
  return wrapper
}

describe('group RPM override request ownership', () => {
  it('keeps the reopened group data when an older response finishes last', async () => {
    const old = deferred()
    mocks.getGroupRPMOverrides.mockReturnValueOnce(old.promise).mockResolvedValueOnce(entries(2))
    const wrapper = await openGroup()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, group: group(2) })
    await flushPromises()
    old.resolve(entries(1))
    await flushPromises()
    expect(wrapper.get('tbody').text()).toContain('user2@example.com')
    expect(wrapper.get('tbody').text()).not.toContain('user1@example.com')
    await wrapper.get('tbody input[type="number"]').setValue('200')
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(2, [{ user_id: 2, rpm_override: 200 }])
  })

  it('does not end the new load or report an error when an old request fails', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const old = deferred()
    const current = deferred()
    mocks.getGroupRPMOverrides.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await openGroup()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, group: group(2) })
    old.reject(new Error('old request failed'))
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(wrapper.find('svg.animate-spin').exists()).toBe(true)
    current.resolve(entries(2))
    await flushPromises()
    expect(wrapper.get('tbody').text()).toContain('user2@example.com')
  })

  it('loads a different group while the dialog remains open', async () => {
    mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockResolvedValueOnce(entries(2))
    const wrapper = await openGroup()
    await flushPromises()
    await wrapper.setProps({ group: group(2) })
    await flushPromises()
    expect(mocks.getGroupRPMOverrides).toHaveBeenLastCalledWith(2)
    expect(wrapper.get('tbody').text()).toContain('user2@example.com')
  })

  it('discards the previous group draft before loading the next group', async () => {
    mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockReturnValueOnce(deferred().promise)
    const wrapper = await openGroup()
    await flushPromises()
    await wrapper.get('tbody input[type="number"]').setValue('200')
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, group: group(2) })
    expect(wrapper.findAll('button').some(button => button.text() === 'common.save')).toBe(false)
  })

  it('ignores a request failure after unmounting', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const request = deferred()
    mocks.getGroupRPMOverrides.mockReturnValue(request.promise)
    const wrapper = await openGroup()
    wrapper.unmount()
    request.reject(new Error('offline'))
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
  })
})


describe('pending edits belong to their original group', () => {
  it.each(['resolve', 'reject'] as const)('ignores a late save %s while a new save is pending', async (outcome) => {
    const old = deferred()
    const current = deferred()
    mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockResolvedValueOnce(entries(2))
    mocks.batchSetGroupRPMOverrides.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await openGroup()
    await flushPromises()
    await wrapper.get('tbody input[type="number"]').setValue('200')
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    await wrapper.setProps({ group: group(2) })
    await flushPromises()
    await wrapper.get('tbody input[type="number"]').setValue('200')
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    if (outcome === 'resolve') old.resolve(undefined)
    else old.reject(new Error('old save failed'))
    await flushPromises()
    expect(wrapper.emitted('close')).toBeUndefined()
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(wrapper.findAll('button').find(button => button.text() === 'common.save')!.attributes('disabled')).toBeDefined()
    current.resolve(undefined)
    await flushPromises()
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('ignores a search result after switching groups', async () => {
    vi.useFakeTimers()
    try {
      const old = deferred()
      mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockResolvedValueOnce(entries(2))
      mocks.listUsers.mockReturnValueOnce(old.promise)
      const wrapper = await openGroup()
      await flushPromises()
      await wrapper.get('input[type="text"]').setValue('old query')
      await vi.advanceTimersByTimeAsync(300)
      expect(mocks.listUsers).toHaveBeenCalledTimes(1)
      await wrapper.setProps({ group: group(2) })
      old.resolve({ items: [{ id: 99, email: 'stale@example.com' }] })
      await flushPromises()
      expect(wrapper.text()).not.toContain('stale@example.com')
      expect(wrapper.get('tbody').text()).toContain('user2@example.com')
    } finally { vi.useRealTimers() }
  })
})


it('ignores a late clear response for a previous group', async () => {
  const old = deferred()
  mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockResolvedValueOnce(entries(2))
  mocks.clearGroupRPMOverrides.mockReturnValueOnce(old.promise)
  const wrapper = await openGroup()
  await flushPromises()
  await wrapper.findAll('button').find(button => button.text() === 'admin.groups.clearAll')!.trigger('click')
  await wrapper.setProps({ group: group(2) })
  await flushPromises()
  old.resolve(undefined)
  await flushPromises()
  expect(wrapper.get('tbody').text()).toContain('user2@example.com')
})
