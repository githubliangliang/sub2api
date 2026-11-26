import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminUser } from '@/types'
import UserEditModal from '../UserEditModal.vue'

const mocks = vi.hoisted(() => ({ update: vi.fn(), updateValues: vi.fn(), success: vi.fn(), error: vi.fn(), verify: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: { update: mocks.update }, userAttributes: { updateUserAttributeValues: mocks.updateValues } } }))
vi.mock('@/api', () => ({ totpAPI: { stepUp: mocks.verify } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: mocks.success, showError: mocks.error }) }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showSuccess: mocks.success, showError: mocks.error }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({ ...(await importOriginal<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
const required = () => ({ status: 403, code: 'STEP_UP_REQUIRED' })
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
const user = (id = 1) => ({ id, email: `user${id}@example.test`, username: `user${id}`, role: 'user', concurrency: 1, rpm_limit: 0 } as AdminUser)
function open() {
  return mount(UserEditModal, { props: { show: true, user: user() }, global: { stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' }, UserAttributeForm: true, Icon: true } } })
}
async function promote(w: ReturnType<typeof open>) {
  await w.get('select').setValue('admin'); await w.get('form').trigger('submit'); await flushPromises()
}
async function cancelEditor(w: ReturnType<typeof open>) {
  await w.findAll('button').find(button => button.text() === 'common.cancel')!.trigger('click')
}
async function verify(w: ReturnType<typeof open>) {
  await w.get('input[autocomplete="one-time-code"]').setValue('123456'); await flushPromises()
}
const promptVisible = (w: ReturnType<typeof open>) => w.text().includes('stepUp.title')
beforeEach(() => {
  vi.resetAllMocks(); mocks.update.mockResolvedValue({}); mocks.updateValues.mockResolvedValue({}); mocks.verify.mockResolvedValue({})
})

describe('user edit session with the real step-up controller and dialog', () => {
  it('does not open an old STEP_UP_REQUIRED prompt after another user opens', async () => {
    const old = deferred<unknown>(); mocks.update.mockReturnValueOnce(old.promise)
    const w = open(); await promote(w); await cancelEditor(w); await w.setProps({ user: user(2) })
    old.reject(required()); await flushPromises()
    expect(promptVisible(w)).toBe(false)
    expect(mocks.update).toHaveBeenCalledTimes(1); expect(mocks.error).not.toHaveBeenCalled()
  })

  it('cancels the current prompt immediately when the editor closes', async () => {
    mocks.update.mockRejectedValueOnce(required())
    const w = open(); await promote(w); expect(promptVisible(w)).toBe(true)
    await cancelEditor(w); await flushPromises()
    expect(w.props('show')).toBe(true); expect(promptVisible(w)).toBe(false)
    expect(w.emitted('close')).toHaveLength(1); expect(w.emitted('success')).toBeUndefined()
    expect(mocks.verify).not.toHaveBeenCalled(); expect(mocks.update).toHaveBeenCalledTimes(1)
  })

  it('does not let an old required response join a new user prompt or retry its action', async () => {
    const old = deferred<unknown>()
    mocks.update.mockReturnValueOnce(old.promise).mockRejectedValueOnce(required())
    const w = open(); await promote(w); await cancelEditor(w); await w.setProps({ user: user(2) }); await promote(w)
    expect(promptVisible(w)).toBe(true)
    old.reject(required()); await flushPromises(); await verify(w)
    expect(mocks.update.mock.calls.map(([id]) => id)).toEqual([1, 2, 2])
    expect(w.emitted('success')).toHaveLength(1); expect(mocks.error).not.toHaveBeenCalled()
  })

  it('keeps a closed session verification result away from a newer prompt', async () => {
    const oldVerification = deferred<unknown>()
    mocks.verify.mockReturnValueOnce(oldVerification.promise)
    mocks.update.mockRejectedValueOnce(required()).mockRejectedValueOnce(required())
    const w = open(); await promote(w); await verify(w)
    expect(mocks.verify).toHaveBeenCalledTimes(1)
    await cancelEditor(w); await w.setProps({ user: user(2) }); await promote(w)
    oldVerification.resolve({}); await flushPromises()
    expect(promptVisible(w)).toBe(true)
    expect(mocks.update.mock.calls.map(([id]) => id)).toEqual([1, 2])
    await verify(w)
    expect(mocks.update.mock.calls.map(([id]) => id)).toEqual([1, 2, 2])
    expect(w.emitted('success')).toHaveLength(1)
  })

  it('does not toast an old verification failure into a new user prompt', async () => {
    const oldVerification = deferred<unknown>()
    mocks.verify.mockReturnValueOnce(oldVerification.promise)
    mocks.update.mockRejectedValueOnce(required()).mockRejectedValueOnce(required())
    const w = open(); await promote(w); await verify(w)
    await cancelEditor(w); await w.setProps({ user: user(2) }); await promote(w)
    oldVerification.reject(new Error('old verification failed')); await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(promptVisible(w)).toBe(true)
    expect(mocks.update.mock.calls.map(([id]) => id)).toEqual([1, 2])
  })

  it('still prompts, verifies and retries a current role promotion exactly once', async () => {
    mocks.update.mockRejectedValueOnce(required())
    const w = open(); await promote(w)
    expect(promptVisible(w)).toBe(true); expect(mocks.update).toHaveBeenCalledTimes(1)
    await verify(w)
    expect(mocks.verify).toHaveBeenCalledWith('123456')
    expect(mocks.update).toHaveBeenCalledTimes(2)
    expect(mocks.update.mock.calls[0]).toEqual(mocks.update.mock.calls[1])
    expect(w.emitted('success')).toHaveLength(1); expect(w.emitted('close')).toHaveLength(1)
    expect(mocks.error).not.toHaveBeenCalled()
  })
})
