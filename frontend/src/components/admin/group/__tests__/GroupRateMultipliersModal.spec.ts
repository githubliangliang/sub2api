import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRateMultipliersModal from '../GroupRateMultipliersModal.vue'

const mocks = vi.hoisted(() => ({ list: vi.fn(), getGroupRateMultipliers: vi.fn(), batchSetGroupRateMultipliers: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: { list: mocks.list }, groups: mocks } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
afterEach(() => vi.useRealTimers())
beforeEach(() => {
  vi.clearAllMocks()
  vi.useFakeTimers()
  mocks.getGroupRateMultipliers.mockResolvedValue([])
  mocks.list.mockResolvedValue({ items: [{ id: 7, email: 'user@example.com', status: 'active' }] })
  mocks.batchSetGroupRateMultipliers.mockResolvedValue(undefined)
})

async function openModal() {
  const wrapper = mount(GroupRateMultipliersModal, {
    props: { show: false, group: { id: 1, name: 'Group', platform: 'openai' } as AdminGroup },
    global: { stubs: {
      BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
      Icon: true, PlatformIcon: true, Pagination: true
    } }
  })
  await wrapper.setProps({ show: true })
  await flushPromises()
  return wrapper
}

async function selectUser() {
  const wrapper = await openModal()
  await wrapper.get('input[type="text"]').setValue('user')
  await vi.advanceTimersByTimeAsync(300)
  await flushPromises()
  await wrapper.findAll('button').find(b => b.text().includes('user@example.com'))!.trigger('click')
  return wrapper
}

describe('GroupRateMultipliersModal new override validation', () => {
  it.each(['', '0', '-1', '1e309'])('does not add invalid rate %j', async (value) => {
    const wrapper = await selectUser()
    const input = wrapper.get('input[placeholder="1.0"]')
    await input.setValue('100')
    await input.setValue(value)
    const add = wrapper.findAll('button').find(b => b.text() === 'common.add')!
    expect(add.attributes('disabled')).toBeDefined()
    await add.trigger('click')
    expect(wrapper.find('tbody tr').exists()).toBe(false)
    expect(mocks.batchSetGroupRateMultipliers).not.toHaveBeenCalled()
  })

  it.each([0.25, 0.5, 1])('saves positive rate %s', async (value) => {
    const wrapper = await selectUser()
    await wrapper.get('input[placeholder="1.0"]').setValue(String(value))
    await wrapper.findAll('button').find(b => b.text() === 'common.add')!.trigger('click')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [{ user_id: 7, rate_multiplier: value }])
  })
})

describe('GroupRateMultipliersModal existing override validation', () => {
  beforeEach(() => {
    mocks.getGroupRateMultipliers.mockResolvedValue([
      { user_id: 7, user_email: 'first@example.com', rate_multiplier: 1 },
      { user_id: 8, user_email: 'second@example.com', rate_multiplier: 2 }
    ])
  })

  it.each(['0', '-1', '1e309', '1e', '1abc'])('blocks the whole save for invalid edited rate %j and allows correction', async (value) => {
    const wrapper = await openModal()
    const inputs = wrapper.findAll<HTMLInputElement>('tbody input')
    // Browsers expose incomplete numeric input as an empty value with badInput=true.
    if (['1e', '1abc', '1e309'].includes(value)) {
      Object.defineProperty(inputs[0].element, 'validity', { configurable: true, value: { badInput: true } })
    }
    await inputs[0].setValue(value)
    await inputs[1].setValue('3')
    const save = wrapper.findAll('button').find(b => b.text() === 'common.save')!
    expect(save.attributes('disabled')).toBeDefined()
    // The submit handler must also reject invalid rows if invoked programmatically.
    ;(save.element as HTMLButtonElement).disabled = false
    await save.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRateMultipliers).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeUndefined()
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)

    Object.defineProperty(inputs[0].element, 'validity', { configurable: true, value: { badInput: false } })
    await inputs[0].setValue('0.5')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [
      { user_id: 7, rate_multiplier: 0.5 }, { user_id: 8, rate_multiplier: 3 }
    ])
  })

  it('clears an override without dropping untouched entries', async () => {
    const wrapper = await openModal()
    await wrapper.get('tbody input').setValue('')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [{ user_id: 8, rate_multiplier: 2 }])
  })

  it.each(['input[placeholder="1.0"]', 'input[placeholder="0.5"]', 'tbody input'])('does not replace incomplete numeric input with NaN in %s', async (selector) => {
    const wrapper = await openModal()
    const input = wrapper.get<HTMLInputElement>(selector)
    Object.defineProperty(input.element, 'validity', { configurable: true, value: { badInput: true } })
    const valueSetter = vi.spyOn(input.element, 'value', 'set')
    await input.setValue('1e')
    expect(valueSetter).not.toHaveBeenCalledWith('NaN')
    valueSetter.mockRestore()
  })

  it('blocks non-finite values produced by batch multiplication at final save', async () => {
    const wrapper = await openModal()
    await wrapper.get('input[placeholder="0.5"]').setValue('1e308')
    await wrapper.findAll('button').find(b => b.text() === 'admin.groups.applyMultiplier')!.trigger('click')
    const save = wrapper.findAll('button').find(b => b.text() === 'common.save')!
    expect(save.attributes('disabled')).toBeDefined()
    ;(save.element as HTMLButtonElement).disabled = false
    await save.trigger('click')
    expect(mocks.batchSetGroupRateMultipliers).not.toHaveBeenCalled()
  })

  it.each(['existing', 'new', 'batch'])('preserves a progressively typed exponent in the %s input until submission', async (path) => {
    const wrapper = path === 'new' ? await selectUser() : await openModal()
    const selector = path === 'existing' ? 'tbody input' : path === 'new' ? 'input[placeholder="1.0"]' : 'input[placeholder="0.5"]'
    const input = wrapper.get<HTMLInputElement>(selector)
    for (const value of ['2', '2e', '2e1']) {
      Object.defineProperty(input.element, 'validity', { configurable: true, value: { badInput: value === '2e' } })
      input.element.value = value
      await input.trigger('input')
    }
    expect.soft(input.element.value).toBe('2e1')
    // Append to the value actually left in the control, as the next keystroke would.
    input.element.value += '2'
    await input.trigger('input')
    expect.soft(input.element.value).toBe('2e12')
    if (path !== 'existing') {
      const action = path === 'new' ? 'common.add' : 'admin.groups.applyMultiplier'
      await wrapper.findAll('button').find(b => b.text() === action)!.trigger('click')
    }
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [
      { user_id: 7, rate_multiplier: 2e12 },
      { user_id: 8, rate_multiplier: path === 'batch' ? 4e12 : 2 }
    ])
  })
})
