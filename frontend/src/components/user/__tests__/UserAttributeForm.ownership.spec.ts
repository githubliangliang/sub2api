import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import UserAttributeForm from '../UserAttributeForm.vue'

const mocks = vi.hoisted(() => ({ getUserAttributeValues: vi.fn(), listEnabledDefinitions: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { userAttributes: mocks } }))
enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())
beforeEach(() => {
  vi.resetAllMocks()
  mocks.listEnabledDefinitions.mockResolvedValue([{ id: 1, name: 'Team', type: 'text' }])
  mocks.getUserAttributeValues.mockResolvedValue([])
  vi.spyOn(console, 'error').mockImplementation(() => {})
})
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: Error) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
function open() {
  return mount(UserAttributeForm, { props: { userId: 1, modelValue: {} }, global: { stubs: { Select: true } } })
}
const value = (text: string) => [{ attribute_id: 1, value: text }]

describe('user attribute request ownership', () => {
  it('ignores a previous user response after the current user loads', async () => {
    const old = deferred<ReturnType<typeof value>>()
    mocks.getUserAttributeValues.mockReturnValueOnce(old.promise).mockResolvedValueOnce(value('current'))
    const w = open(); await w.setProps({ userId: 2 }); await flushPromises()
    old.resolve(value('old')); await flushPromises()
    expect((w.get('input').element as HTMLInputElement).value).toBe('current')
    expect(w.emitted('update:modelValue')).toEqual([[{ 1: 'current' }]])
  })

  it('does not restore values after resetting to a new user', async () => {
    const old = deferred<ReturnType<typeof value>>()
    mocks.getUserAttributeValues.mockReturnValueOnce(old.promise)
    const w = open(); await flushPromises(); await w.setProps({ userId: undefined })
    old.resolve(value('old')); await flushPromises()
    expect((w.get('input').element as HTMLInputElement).value).toBe('')
    expect(w.emitted('update:modelValue')).toBeUndefined()
  })

  it('clears the previous values when the current user read fails', async () => {
    mocks.getUserAttributeValues.mockResolvedValueOnce(value('old')).mockRejectedValueOnce(new Error('unavailable'))
    const w = open(); await flushPromises(); await w.setProps({ userId: 2 }); await flushPromises()
    expect((w.get('input').element as HTMLInputElement).value).toBe('')
    expect(w.get('input').attributes('disabled')).toBeDefined()
  })

  it('does not report stale errors or enable editing while the current read is pending', async () => {
    const old = deferred<ReturnType<typeof value>>()
    const current = deferred<ReturnType<typeof value>>()
    mocks.getUserAttributeValues.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = open(); await w.setProps({ userId: 2 }); await flushPromises()
    old.reject(new Error('stale')); await flushPromises()
    expect(console.error).not.toHaveBeenCalled()
    expect(w.get('input').attributes('disabled')).toBeDefined()
    current.resolve(value('current')); await flushPromises()
    expect(w.get('input').attributes('disabled')).toBeUndefined()
  })

  it('ignores pending values and definition failures after unmount', async () => {
    const old = deferred<ReturnType<typeof value>>()
    const definitions = deferred<never>()
    mocks.getUserAttributeValues.mockReturnValueOnce(old.promise)
    mocks.listEnabledDefinitions.mockReturnValueOnce(definitions.promise)
    const w = open(); w.unmount()
    old.resolve(value('old')); definitions.reject(new Error('stale definitions')); await flushPromises()
    expect(w.emitted('update:modelValue')).toBeUndefined()
    expect(console.error).not.toHaveBeenCalled()
  })

  it('accepts an explicit empty model update for a new user', async () => {
    const w = mount(UserAttributeForm, { props: { modelValue: { 1: 'draft' } }, global: { stubs: { Select: true } } })
    await flushPromises(); await w.setProps({ modelValue: {} })
    expect((w.get('input').element as HTMLInputElement).value).toBe('')
  })
})
