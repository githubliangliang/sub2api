import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AccountPriorityCell from '../AccountPriorityCell.vue'
import type { Account } from '@/types'
import { update } from '@/api/admin/accounts'

vi.mock('@/api/admin/accounts', () => ({ update: vi.fn() }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

const account = (overrides: Partial<Account> = {}) => ({
  id: 7, name: 'Claude 1', platform: 'anthropic', type: 'oauth', priority: 3,
  ...overrides,
}) as Account

const mountCell = (value = account()) => mount(AccountPriorityCell, { props: { account: value } })

beforeEach(() => {
  vi.useFakeTimers()
  vi.mocked(update).mockReset().mockImplementation(async (id, req) => account({ id, priority: req.priority }))
})
afterEach(() => {
  vi.useRealTimers()
})

describe('AccountPriorityCell', () => {
  it('batches rapid +/- clicks into a single priority-only update', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await wrapper.get('[data-testid="account-priority-decrement"]').trigger('click')
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('4')
    expect(update).not.toHaveBeenCalled()

    await vi.runAllTimersAsync()
    await flushPromises()

    expect(update).toHaveBeenCalledTimes(1)
    expect(update).toHaveBeenCalledWith(7, { priority: 4 })
    expect(wrapper.emitted('updated')?.[0]?.[0]).toMatchObject({ id: 7, priority: 4 })
  })

  it('does not go below 1', async () => {
    const wrapper = mountCell(account({ priority: 1 }))
    const dec = wrapper.get('[data-testid="account-priority-decrement"]')
    expect(dec.attributes('disabled')).toBeDefined()
    // 到达下限时按钮仍应随悬停显隐，而不是常驻半透明
    await dec.trigger('click')
    await vi.runAllTimersAsync()
    expect(update).not.toHaveBeenCalled()
  })

  it('saves a typed value on Enter and ignores unchanged input', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    const input = wrapper.get('[data-testid="account-priority-input"]')
    await input.setValue('12')
    await input.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(update).toHaveBeenCalledWith(7, { priority: 12 })

    vi.mocked(update).mockClear()
    await wrapper.setProps({ account: account({ priority: 12 }) })
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    await wrapper.get('[data-testid="account-priority-input"]').trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(update).not.toHaveBeenCalled()
  })

  it('Escape cancels typing without saving', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    const input = wrapper.get('[data-testid="account-priority-input"]')
    await input.setValue('50')
    await input.trigger('keydown', { key: 'Escape' })
    await flushPromises()
    expect(update).not.toHaveBeenCalled()
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('3')
  })

  it('reverts and emits an error when the update fails', async () => {
    vi.mocked(update).mockRejectedValueOnce(new Error('boom'))
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await vi.runAllTimersAsync()
    await flushPromises()
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('3')
    expect(wrapper.emitted('error')).toHaveLength(1)
    expect(wrapper.emitted('updated')).toBeUndefined()
  })
})


describe('AccountPriorityCell lifecycle and input', () => {
  it('waits 450ms and disables controls until the request settles', async () => {
    let finish!: (value: Account) => void
    vi.mocked(update).mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await vi.advanceTimersByTimeAsync(449)
    expect(update).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(update).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="account-priority-increment"]').attributes('disabled')).toBeDefined()
    finish(account({ priority: 4 }))
    await flushPromises()
    expect(wrapper.get('[data-testid="account-priority-increment"]').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it.each(['', 'abc', '12oops', '1e3', '2.5'])('does not submit invalid input %j', async (value) => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    const input = wrapper.get('[data-testid="account-priority-input"]')
    await input.setValue(value)
    await input.trigger('blur')
    await flushPromises()
    expect(update).not.toHaveBeenCalled()
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('3')
    wrapper.unmount()
  })

  it('bounds typed values and submits Enter then blur only once', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    const input = wrapper.get('[data-testid="account-priority-input"]')
    await input.setValue('10000')
    await input.trigger('keydown', { key: 'Enter' })
    await input.trigger('blur')
    await flushPromises()
    expect(update).toHaveBeenCalledTimes(1)
    expect(update).toHaveBeenCalledWith(7, { priority: 9999 })
    expect(wrapper.get('[data-testid="account-priority-increment"]').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })

  it('supports keyboard adjustment and cancellation of a pending button edit', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await wrapper.get('[data-testid="account-priority-value"]').trigger('click')
    const input = wrapper.get('[data-testid="account-priority-input"]')
    await input.trigger('keydown', { key: 'ArrowUp' })
    expect((input.element as HTMLInputElement).value).toBe('5')
    await input.trigger('keydown', { key: 'ArrowDown' })
    expect((input.element as HTMLInputElement).value).toBe('4')
    await input.trigger('keydown', { key: 'Escape' })
    await vi.runAllTimersAsync()
    expect(update).not.toHaveBeenCalled()
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('3')
    wrapper.unmount()
  })

  it('flushes a pending edit once on unmount', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    wrapper.unmount()
    await vi.runAllTimersAsync()
    await flushPromises()
    expect(update).toHaveBeenCalledTimes(1)
    expect(update).toHaveBeenCalledWith(7, { priority: 4 })
  })

  it('saves pending changes to their original account when a row is reused', async () => {
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await wrapper.setProps({ account: account({ id: 8, priority: 20 }) })
    await vi.runAllTimersAsync()
    await flushPromises()
    expect(update).toHaveBeenCalledWith(7, { priority: 4 })
    expect(update).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('20')
    expect(wrapper.emitted('updated')).toBeUndefined()
    wrapper.unmount()
  })

  it('ignores an old account response after reuse', async () => {
    let finish!: (value: Account) => void
    vi.mocked(update).mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mountCell()
    await wrapper.get('[data-testid="account-priority-increment"]').trigger('click')
    await vi.runAllTimersAsync()
    await wrapper.setProps({ account: account({ id: 8, priority: 20 }) })
    finish(account({ priority: 4 }))
    await flushPromises()
    expect(wrapper.get('[data-testid="account-priority-value"]').text()).toBe('20')
    expect(wrapper.emitted('updated')).toBeUndefined()
    wrapper.unmount()
  })
})
