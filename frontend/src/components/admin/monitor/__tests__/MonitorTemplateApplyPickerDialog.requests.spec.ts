import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AssociatedMonitorsResponse } from '@/api/admin/channelMonitorTemplate'
import MonitorTemplateApplyPickerDialog from '../MonitorTemplateApplyPickerDialog.vue'

const mocks = vi.hoisted(() => ({
  listAssociatedMonitors: vi.fn(), apply: vi.fn(), showError: vi.fn(), showSuccess: vi.fn(),
}))
vi.mock('@/api/admin', () => ({ adminAPI: { channelMonitorTemplate: mocks } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => mocks }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
beforeEach(() => {
  vi.resetAllMocks()
  mocks.apply.mockResolvedValue({ affected: 1 })
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

const items = (...ids: number[]): AssociatedMonitorsResponse => ({
  items: ids.map(id => ({ id, name: `Monitor ${id}`, provider: 'openai', api_mode: 'responses', enabled: true })),
})

function mountDialog() {
  return mount(MonitorTemplateApplyPickerDialog, {
    props: { show: true, templateId: 1, templateName: 'Template' },
    global: {
      stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' } },
    },
  })
}

describe('template association request ownership', () => {
  it('applies the current template only to its own monitors after an older response', async () => {
    const old = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockReturnValueOnce(old.promise).mockResolvedValueOnce(items(20))
    const wrapper = mountDialog()
    await wrapper.setProps({ templateId: 2 })
    await flushPromises()
    old.resolve(items(10))
    await flushPromises()
    expect(wrapper.text()).toContain('Monitor 20')
    expect(wrapper.text()).not.toContain('Monitor 10')
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(mocks.apply).toHaveBeenCalledWith(2, [20])
    expect(wrapper.emitted('applied')).toEqual([[1]])
    expect(wrapper.emitted('close')).toEqual([[]])
  })

  it.each(['resolve', 'reject'] as const)('keeps a newer request loading when the older one %ss', async outcome => {
    const old = deferred<AssociatedMonitorsResponse>()
    const current = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = mountDialog()
    await wrapper.setProps({ templateId: 2 })
    if (outcome === 'resolve') old.resolve(items(10))
    else old.reject({ message: 'obsolete failure' })
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('common.loading')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve(items(20))
    await flushPromises()
    expect(wrapper.text()).toContain('Monitor 20')
  })

  it('does not reuse rows or selection while reopening the same template', async () => {
    const current = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockResolvedValueOnce(items(10)).mockReturnValueOnce(current.promise)
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    expect(wrapper.text()).not.toContain('Monitor 10')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve(items(11))
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenCalledWith(1, [11])
  })

  it.each(['close', 'unmount'] as const)('ignores a pending list error after %s', async action => {
    const pending = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockReturnValueOnce(pending.promise)
    const wrapper = mountDialog()
    if (action === 'close') await wrapper.setProps({ show: false })
    else wrapper.unmount()
    pending.reject({ message: 'obsolete failure' })
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it('invalidates pending reads immediately when cancel is clicked', async () => {
    const pending = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockReturnValueOnce(pending.promise)
    const wrapper = mountDialog()
    await wrapper.get('button.btn-secondary').trigger('click')
    pending.reject({ message: 'closed failure' })
    await flushPromises()
    expect(wrapper.emitted('close')).toEqual([[]])
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it('reports a current failure and permits a successful retry on reopening', async () => {
    mocks.listAssociatedMonitors.mockRejectedValueOnce({ message: 'current failure' }).mockResolvedValueOnce(items(20))
    const wrapper = mountDialog()
    await flushPromises()
    expect(mocks.showError).toHaveBeenCalledWith('current failure')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenCalledWith(1, [20])
  })

  it('clears the previous template when no template is selected', async () => {
    mocks.listAssociatedMonitors.mockResolvedValue(items(10))
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.setProps({ templateId: null })
    expect(wrapper.text()).not.toContain('Monitor 10')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    expect(mocks.listAssociatedMonitors).toHaveBeenCalledTimes(1)
  })

  it('keeps only the latest retry for the same template and rejects apply while it loads', async () => {
    const old = deferred<AssociatedMonitorsResponse>()
    const current = deferred<AssociatedMonitorsResponse>()
    mocks.listAssociatedMonitors.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = mountDialog()
    const methods = wrapper.vm as unknown as { fetchMonitors: (id: number) => Promise<void>; handleApply: () => Promise<void> }
    const retry = methods.fetchMonitors(1)
    await methods.handleApply()
    expect(mocks.apply).not.toHaveBeenCalled()
    old.resolve(items(10))
    await flushPromises()
    expect(wrapper.text()).toContain('common.loading')
    expect(wrapper.text()).not.toContain('Monitor 10')
    current.resolve(items(11))
    await retry
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenCalledWith(1, [11])
  })
})

describe('template apply session ownership', () => {
  it.each([
    { outcome: 'resolve', transition: 'reopen' },
    { outcome: 'reject', transition: 'reopen' },
    { outcome: 'resolve', transition: 'template switch' },
    { outcome: 'reject', transition: 'template switch' },
  ])('ignores an old apply $outcome after $transition', async ({ outcome, transition }) => {
    const old = deferred<{ affected: number }>()
    const current = deferred<{ affected: number }>()
    mocks.listAssociatedMonitors.mockResolvedValueOnce(items(10)).mockResolvedValueOnce(items(20))
    mocks.apply.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenNthCalledWith(1, 1, [10])
    if (transition === 'reopen') await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, templateId: 2 })
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeUndefined()
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenNthCalledWith(2, 2, [20])
    if (outcome === 'resolve') old.resolve({ affected: 1 })
    else old.reject({ message: 'old apply failure' })
    await flushPromises()
    expect(wrapper.emitted('applied')).toBeUndefined()
    expect(wrapper.emitted('close')).toBeUndefined()
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(mocks.showSuccess).not.toHaveBeenCalled()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve({ affected: 2 })
    await flushPromises()
    expect(wrapper.emitted('applied')).toEqual([[2]])
    expect(wrapper.emitted('close')).toEqual([[]])
  })

  it.each(['close', 'unmount'] as const)('ignores a pending apply result after %s', async action => {
    const pending = deferred<{ affected: number }>()
    mocks.listAssociatedMonitors.mockResolvedValue(items(10))
    mocks.apply.mockReturnValueOnce(pending.promise)
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    if (action === 'close') await wrapper.setProps({ show: false })
    else wrapper.unmount()
    pending.resolve({ affected: 1 })
    await flushPromises()
    expect(wrapper.emitted('applied')).toBeUndefined()
    expect(wrapper.emitted('close')).toBeUndefined()
    expect(mocks.showSuccess).not.toHaveBeenCalled()
  })

  it('binds the request to the selected snapshot and prevents duplicate submission', async () => {
    const pending = deferred<{ affected: number }>()
    mocks.listAssociatedMonitors.mockResolvedValue(items(10, 11))
    mocks.apply.mockReturnValueOnce(pending.promise)
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.findAll('li')[1].trigger('click')
    await wrapper.get('button.btn-primary').trigger('click')
    await wrapper.get('button.btn-primary').trigger('click')
    await (wrapper.vm as unknown as { handleApply: () => Promise<void> }).handleApply()
    await wrapper.findAll('li')[1].trigger('click')
    expect(mocks.apply).toHaveBeenCalledTimes(1)
    expect(mocks.apply).toHaveBeenCalledWith(1, [10])
    pending.resolve({ affected: 1 })
    await flushPromises()
    expect(wrapper.emitted('applied')).toEqual([[1]])
  })

  it('rejects direct apply calls after cancellation while the parent is closing', async () => {
    mocks.listAssociatedMonitors.mockResolvedValue(items(10))
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('button.btn-secondary').trigger('click')
    await (wrapper.vm as unknown as { handleApply: () => Promise<void> }).handleApply()
    expect(mocks.apply).not.toHaveBeenCalled()
  })

  it('reports a current apply failure and permits retry', async () => {
    mocks.listAssociatedMonitors.mockResolvedValue(items(10))
    mocks.apply.mockRejectedValueOnce({ message: 'current apply failure' }).mockResolvedValueOnce({ affected: 1 })
    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(mocks.showError).toHaveBeenCalledWith('current apply failure')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeUndefined()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(mocks.apply).toHaveBeenCalledTimes(2)
    expect(wrapper.emitted('applied')).toEqual([[1]])
  })

  it('keeps select all and select none behavior', async () => {
    mocks.listAssociatedMonitors.mockResolvedValue(items(10, 11))
    const wrapper = mountDialog()
    await flushPromises()
    const button = (label: string) => wrapper.findAll('button').find(item => item.text() === label)!
    await button('admin.channelMonitor.template.selectNone').trigger('click')
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await button('common.selectAll').trigger('click')
    await wrapper.get('button.btn-primary').trigger('click')
    expect(mocks.apply).toHaveBeenCalledWith(1, [10, 11])
  })
})
