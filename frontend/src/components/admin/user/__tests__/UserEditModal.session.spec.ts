import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminUser } from '@/types'
import UserEditModal from '../UserEditModal.vue'
import UserAttributeForm from '@/components/user/UserAttributeForm.vue'

const mocks = vi.hoisted(() => ({ update: vi.fn(), updateValues: vi.fn(), values: vi.fn(), definitions: vi.fn(), success: vi.fn(), error: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: { update: mocks.update }, userAttributes: { updateUserAttributeValues: mocks.updateValues, getUserAttributeValues: mocks.values, listEnabledDefinitions: mocks.definitions } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: mocks.success, showError: mocks.error }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({ ...(await importOriginal<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
function deferred<T>() { let resolve!: (value: T) => void; let reject!: (error: Error) => void; const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej }); return { promise, resolve, reject } }
const user = (id = 1) => ({ id, email: `user${id}@example.test`, username: `user${id}`, notes: '', role: 'user', concurrency: 1, rpm_limit: 0 } as AdminUser)
function open() {
  return mount(UserEditModal, { props: { show: true, user: user() }, global: { stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' }, TotpStepUpDialog: true, Icon: true, Select: true } } })
}
beforeEach(() => {
  vi.resetAllMocks()
  mocks.update.mockResolvedValue({}); mocks.updateValues.mockResolvedValue({})
  mocks.definitions.mockResolvedValue([{ id: 1, name: 'Team', type: 'text' }])
  mocks.values.mockImplementation((id: number) => Promise.resolve([{ attribute_id: 1, value: `loaded-${id}` }]))
})

describe('user edit save ownership', () => {
  it('saves a snapshot for the originating user while a new user is edited', async () => {
    const old = deferred<unknown>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises()
    await w.getComponent(UserAttributeForm).get('input').setValue('submitted-1')
    await w.get('form').trigger('submit'); await w.setProps({ user: user(2) }); await flushPromises()
    await w.getComponent(UserAttributeForm).get('input').setValue('draft-2')
    old.resolve({}); await flushPromises()
    expect(mocks.updateValues).toHaveBeenCalledWith(1, { 1: 'submitted-1' })
    expect(w.emitted('success')).toBeUndefined(); expect(w.emitted('close')).toBeUndefined()
    expect(mocks.success).not.toHaveBeenCalled()
    expect((w.getComponent(UserAttributeForm).get('input').element as HTMLInputElement).value).toBe('draft-2')
  })

  it('resets the form on close and reopen of the same user and ignores old completion', async () => {
    const old = deferred<unknown>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises(); await w.get('input[type="email"]').setValue('draft@example.test')
    await w.get('form').trigger('submit'); await w.setProps({ show: false }); await w.setProps({ show: true }); await flushPromises()
    old.resolve({}); await flushPromises()
    expect((w.get('input[type="email"]').element as HTMLInputElement).value).toBe('user1@example.test')
    expect(w.emitted('close')).toBeUndefined()
  })

  it('reloads attributes for a reopened user and drops a previous session read', async () => {
    const old = deferred<{ attribute_id: number; value: string }[]>()
    mocks.values.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises(); await w.setProps({ show: false }); await w.setProps({ show: true }); await flushPromises()
    old.resolve([{ attribute_id: 1, value: 'previous-session' }]); await flushPromises()
    expect(mocks.values).toHaveBeenCalledTimes(2)
    expect((w.getComponent(UserAttributeForm).get('input').element as HTMLInputElement).value).toBe('loaded-1')
    await w.get('form').trigger('submit'); await flushPromises()
    expect(mocks.updateValues).toHaveBeenCalledWith(1, { 1: 'loaded-1' })
  })

  it('keeps a stale error and finally from changing a new pending submission', async () => {
    const old = deferred<unknown>(); const current = deferred<unknown>()
    mocks.update.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = open(); await flushPromises(); await w.get('form').trigger('submit')
    await w.setProps({ user: user(2) }); await flushPromises(); await w.get('form').trigger('submit')
    old.reject(new Error('stale')); await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(w.get('button[form="edit-user-form"]').attributes('disabled')).toBeDefined()
    current.resolve({}); await flushPromises(); expect(w.emitted('success')).toHaveLength(1)
  })

  it('does not write attributes that failed to load while retaining a normal user update', async () => {
    const error = vi.spyOn(console, 'error').mockImplementation(() => {})
    mocks.values.mockRejectedValueOnce(new Error('unavailable'))
    const w = open(); await flushPromises(); await w.get('form').trigger('submit'); await flushPromises()
    expect(mocks.update).toHaveBeenCalledTimes(1); expect(mocks.updateValues).not.toHaveBeenCalled()
    expect(w.emitted('success')).toHaveLength(1); error.mockRestore()
  })

  it('invalidates a save on cancel immediately before the parent updates show', async () => {
    const old = deferred<unknown>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises(); await w.get('form').trigger('submit')
    await w.findAll('button').find(button => button.text() === 'common.cancel')!.trigger('click')
    old.resolve({}); await flushPromises()
    expect(w.props('show')).toBe(true); expect(w.emitted('close')).toHaveLength(1)
    expect(w.emitted('success')).toBeUndefined(); expect(mocks.success).not.toHaveBeenCalled()
    await (w.vm as any).handleUpdateUser(); expect(mocks.update).toHaveBeenCalledTimes(1)
  })

  it('prevents duplicate submissions and drops completion after unmount', async () => {
    const old = deferred<unknown>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises(); await w.get('form').trigger('submit'); await w.get('form').trigger('submit')
    expect(mocks.update).toHaveBeenCalledTimes(1)
    w.unmount(); old.resolve({}); await flushPromises()
    expect(w.emitted('success')).toBeUndefined(); expect(mocks.success).not.toHaveBeenCalled()
  })
})
