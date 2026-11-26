import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminUser } from '@/types'
import UserAllowedGroupsModal from '../UserAllowedGroupsModal.vue'

const mocks = vi.hoisted(() => ({ list: vi.fn(), update: vi.fn(), showSuccess: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: { list: mocks.list }, users: { update: mocks.update } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: mocks.showSuccess }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())
const group = (id: number, exclusive = true) => ({ id, name: `Group ${id}`, platform: 'openai', is_exclusive: exclusive, subscription_type: 'standard', status: 'active', rate_multiplier: 1 })
const response = { items: [group(7), group(8, false)] }
const user = (id = 1) => ({ id, email: `user${id}@example.test`, allowed_groups: [7], group_rates: { 7: 0.5, 8: 0.8 } } as unknown as AdminUser)
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (error: Error) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
beforeEach(() => {
  vi.resetAllMocks()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  mocks.list.mockResolvedValue(response)
  mocks.update.mockResolvedValue(undefined)
})
async function openDialog() {
  const wrapper = mount(UserAllowedGroupsModal, {
    props: { show: false, user: user() },
    global: { stubs: { BaseDialog: { name: 'BaseDialog', props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' }, PlatformIcon: true } }
  })
  await wrapper.setProps({ show: true })
  return wrapper
}

describe('UserAllowedGroupsModal load and save ownership', () => {
  it('blocks button and method saves while groups are loading', async () => {
    const pending = deferred<typeof response>(); mocks.list.mockReturnValueOnce(pending.promise)
    const w = await openDialog(); await flushPromises()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (w.vm as any).handleSave()
    expect(mocks.update).not.toHaveBeenCalled()
    pending.resolve(response); await flushPromises()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeUndefined()
  })

  it('blocks saves after a failed load and permits a successful retry', async () => {
    mocks.list.mockRejectedValueOnce(new Error('Offline'))
    const w = await openDialog(); await flushPromises()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (w.vm as any).handleSave(); expect(mocks.update).not.toHaveBeenCalled()
    await w.get('button[data-test="retry-groups"]').trigger('click'); await flushPromises()
    await w.get('button.btn-primary').trigger('click'); await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(1, { allowed_groups: [7], group_rates: { 7: 0.5, 8: 0.8 } })
  })

  it('preserves public default access, exclusive selection and explicit rate removal', async () => {
    const w = await openDialog(); await flushPromises()
    await w.get('input[type="checkbox"]').setValue(false)
    await w.findAll('input[type="number"]')[0]!.setValue('')
    await w.get('button.btn-primary').trigger('click'); await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(1, { allowed_groups: [], group_rates: { 7: null, 8: 0.8 } })
    expect(w.emitted('success')).toHaveLength(1)
    expect(w.emitted('close')).toHaveLength(1)
  })

  it('does not reuse a previous successful load when reopening fails', async () => {
    const w = await openDialog(); await flushPromises()
    await w.setProps({ show: false }); mocks.list.mockRejectedValueOnce(new Error('Offline'))
    await w.setProps({ show: true }); await flushPromises()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (w.vm as any).handleSave(); expect(mocks.update).not.toHaveBeenCalled()
  })

  it('ignores earlier rows and initializes only the newly selected user', async () => {
    const old = deferred<typeof response>(); mocks.list.mockReturnValueOnce(old.promise).mockResolvedValueOnce({ items: [group(9)] })
    const w = await openDialog(); await w.setProps({ user: { ...user(2), allowed_groups: [9], group_rates: {} } })
    await flushPromises(); old.resolve({ items: [group(123)] }); await flushPromises()
    expect(w.text()).not.toContain('Group 123')
    await w.get('button.btn-primary').trigger('click'); await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(2, expect.objectContaining({ allowed_groups: [9] }))
  })

  it('ignores stale failures and completion while the new user is loading', async () => {
    const old = deferred<typeof response>(); const current = deferred<typeof response>()
    mocks.list.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = await openDialog(); await w.setProps({ user: user(2) })
    old.reject(new Error('Old failure')); await flushPromises()
    expect(console.error).not.toHaveBeenCalled()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve(response); await flushPromises()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeUndefined()
  })

  it('binds an in-flight save to its user and values without closing the next session', async () => {
    const saved = deferred<void>(); mocks.update.mockReturnValueOnce(saved.promise)
    const w = await openDialog(); await flushPromises()
    await w.get('button.btn-primary').trigger('click')
    await w.setProps({ user: { ...user(2), allowed_groups: [], group_rates: {} } }); await flushPromises()
    saved.resolve(); await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(1, { allowed_groups: [7], group_rates: { 7: 0.5, 8: 0.8 } })
    expect(w.emitted('success')).toBeUndefined(); expect(w.emitted('close')).toBeUndefined()
    expect(mocks.showSuccess).not.toHaveBeenCalled()
    expect(w.get('button.btn-primary').attributes('disabled')).toBeUndefined()
  })

  it('isolates a late save error and finally from a newer pending save', async () => {
    const old = deferred<void>(); const current = deferred<void>()
    mocks.update.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = await openDialog(); await flushPromises(); await w.get('button.btn-primary').trigger('click')
    await w.setProps({ show: false }); await w.setProps({ show: true }); await flushPromises()
    await w.get('button.btn-primary').trigger('click'); old.reject(new Error('Old save')); await flushPromises()
    expect(console.error).not.toHaveBeenCalled(); expect(w.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve(); await flushPromises(); expect(w.emitted('close')).toHaveLength(1)
  })

  it('keeps a later retry in charge when older reads finish out of order', async () => {
    const old = deferred<typeof response>(); const current = deferred<typeof response>()
    mocks.list.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = await openDialog()
    const retry = (w.vm as any).load(); current.resolve({ items: [group(9)] }); await retry
    old.resolve(response); await flushPromises()
    expect(w.text()).toContain('Group 9'); expect(w.text()).not.toContain('Group 7')
    await w.get('button.btn-primary').trigger('click'); await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(1, { allowed_groups: [], group_rates: undefined })
  })

  it('ignores reads after closing or unmounting', async () => {
    const old = deferred<typeof response>(); mocks.list.mockReturnValueOnce(old.promise)
    const w = await openDialog(); await w.setProps({ show: false }); w.unmount()
    old.reject(new Error('stale')); await flushPromises()
    expect(console.error).not.toHaveBeenCalled(); expect(mocks.update).not.toHaveBeenCalled()
  })

  it('invalidates a pending read immediately on cancel before the parent hides it', async () => {
    const old = deferred<typeof response>(); mocks.list.mockReturnValueOnce(old.promise)
    const w = await openDialog()
    await w.get('button.btn-secondary').trigger('click')
    old.resolve(response); await flushPromises()
    expect(w.props('show')).toBe(true)
    expect(w.text()).not.toContain('Group 7')
    await (w.vm as any).handleSave(); expect(mocks.update).not.toHaveBeenCalled()
  })

  it('invalidates a pending save immediately on dialog close before the parent hides it', async () => {
    const old = deferred<void>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = await openDialog(); await flushPromises(); await w.get('button.btn-primary').trigger('click')
    w.getComponent({ name: 'BaseDialog' }).vm.$emit('close')
    old.resolve(); await flushPromises()
    expect(w.props('show')).toBe(true); expect(w.emitted('close')).toHaveLength(1)
    expect(w.emitted('success')).toBeUndefined(); expect(mocks.showSuccess).not.toHaveBeenCalled()
  })

  it('prevents duplicate submissions and ignores completion after unmount', async () => {
    const saved = deferred<void>(); mocks.update.mockReturnValueOnce(saved.promise)
    const w = await openDialog(); await flushPromises()
    const save = (w.vm as any).handleSave(); await (w.vm as any).handleSave()
    expect(mocks.update).toHaveBeenCalledTimes(1)
    w.unmount(); saved.resolve(); await save; await flushPromises()
    expect(w.emitted('close')).toBeUndefined(); expect(mocks.showSuccess).not.toHaveBeenCalled()
  })
})
