import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRPMOverridesModal from '../GroupRPMOverridesModal.vue'

const mocks = vi.hoisted(() => ({ list: vi.fn(), getGroupRPMOverrides: vi.fn(), batchSetGroupRPMOverrides: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: { list: mocks.list }, groups: mocks } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
afterEach(() => vi.useRealTimers())
beforeEach(() => {
  vi.clearAllMocks()
  vi.useFakeTimers()
  mocks.getGroupRPMOverrides.mockResolvedValue([])
  mocks.list.mockResolvedValue({ items: [{ id: 7, email: 'user@example.com', status: 'active' }] })
  mocks.batchSetGroupRPMOverrides.mockResolvedValue(undefined)
})

async function openModal() {
  const wrapper = mount(GroupRPMOverridesModal, {
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

describe('GroupRPMOverridesModal new override validation', () => {
  it.each(['', '1.5', '-1', '-0.5', '1e309'])('does not add invalid RPM %j', async (value) => {
    const wrapper = await selectUser()
    const input = wrapper.get('input[placeholder="100"]')
    await input.setValue('100')
    await input.setValue(value)
    const add = wrapper.findAll('button').find(b => b.text() === 'common.add')!
    expect(add.attributes('disabled')).toBeDefined()
    await add.trigger('click')
    expect(wrapper.find('tbody tr').exists()).toBe(false)
    expect(mocks.batchSetGroupRPMOverrides).not.toHaveBeenCalled()
  })

  it.each([0, 2, 100])('saves integer RPM %i including unlimited zero', async (value) => {
    const wrapper = await selectUser()
    await wrapper.get('input[placeholder="100"]').setValue(String(value))
    await wrapper.findAll('button').find(b => b.text() === 'common.add')!.trigger('click')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(1, [{ user_id: 7, rpm_override: value }])
  })
})

describe('GroupRPMOverridesModal existing override validation', () => {
  beforeEach(() => {
    mocks.getGroupRPMOverrides.mockResolvedValue([
      { user_id: 7, user_email: 'first@example.com', rpm_override: 100 },
      { user_id: 8, user_email: 'second@example.com', rpm_override: 200 }
    ])
  })

  it.each(['', '1.5', '-0.5', '-1', '1e309', '1e', '1abc'])('blocks the whole save for invalid edited RPM %j and allows zero after correction', async (value) => {
    const wrapper = await openModal()
    const inputs = wrapper.findAll<HTMLInputElement>('tbody input')
    await inputs[0].setValue(value)
    await inputs[1].setValue('2')
    const save = wrapper.findAll('button').find(b => b.text() === 'common.save')!
    expect(save.attributes('disabled')).toBeDefined()
    ;(save.element as HTMLButtonElement).disabled = false
    await save.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRPMOverrides).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeUndefined()
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
    await inputs[0].setValue('0')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(1, [
      { user_id: 7, rpm_override: 0 }, { user_id: 8, rpm_override: 2 }
    ])
  })

  it('parses a complete exponent without truncating it', async () => {
    const wrapper = await openModal()
    await wrapper.get('tbody input').setValue('1e3')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(1, [
      { user_id: 7, rpm_override: 1000 }, { user_id: 8, rpm_override: 200 }
    ])
  })

  it.each(['input[placeholder="100"]', 'tbody input'])('does not replace incomplete numeric input with NaN in %s', async (selector) => {
    const wrapper = await openModal()
    const input = wrapper.get<HTMLInputElement>(selector)
    const valueSetter = vi.spyOn(input.element, 'value', 'set')
    await input.setValue('1e')
    expect(valueSetter).not.toHaveBeenCalledWith('NaN')
    valueSetter.mockRestore()
  })

  it.each(['existing', 'new'])('preserves a progressively typed exponent in the %s input until submission', async (path) => {
    const wrapper = path === 'new' ? await selectUser() : await openModal()
    const input = wrapper.get<HTMLInputElement>(path === 'existing' ? 'tbody input' : 'input[placeholder="100"]')
    for (const value of ['2', '2e', '2e1']) {
      input.element.value = value
      await input.trigger('input')
    }
    expect.soft(input.element.value).toBe('2e1')
    input.element.value += '2'
    await input.trigger('input')
    expect.soft(input.element.value).toBe('2e12')
    if (path === 'new') await wrapper.findAll('button').find(b => b.text() === 'common.add')!.trigger('click')
    await wrapper.findAll('button').find(b => b.text() === 'common.save')!.trigger('click')
    await flushPromises()
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(1, [
      { user_id: 7, rpm_override: 2e12 }, { user_id: 8, rpm_override: 200 }
    ])
  })
})
