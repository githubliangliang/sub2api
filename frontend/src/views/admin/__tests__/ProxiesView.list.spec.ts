import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, it, expect, vi } from 'vitest'

const { get, showError } = vi.hoisted(() => ({ get: vi.fn(), showError: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: { get } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn(), showWarning: vi.fn() }) }))
vi.mock('vue-i18n', async () => ({ ...await vi.importActual('vue-i18n'), useI18n: () => ({ t: (key: string) => key }) }))

import ProxiesView from '../ProxiesView.vue'

describe('proxy list refresh', () => {
  beforeEach(() => {
    get.mockReset()
    showError.mockReset()
  })

  it('handles invalid backup-selector data without rejecting the mounted page', async () => {
    get.mockImplementation((url: string) => Promise.resolve({ data: url.endsWith('/all') ? null : { items: [], total: 0, pages: 0 } }))
    const wrapper = mount(ProxiesView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
      BaseDialog: true, ConfirmDialog: true, ImportDataModal: true, ProxyAdBanner: true, Icon: true,
    } } })
    try {
      await flushPromises()
      expect(showError).toHaveBeenCalledWith('admin.proxies.failedToLoad')
      expect(wrapper.find('button[title="common.refresh"]').exists()).toBe(true)
    } finally {
      wrapper.unmount()
    }
  })

  it('retains visible rows when a successful refresh contains an invalid list', async () => {
    const errorLog = vi.spyOn(console, 'error').mockImplementation(() => {})
    let malformed = false
    get.mockImplementation((url: string) => Promise.resolve({ data: url.endsWith('/all') ? [] : malformed ? { items: null } : {
      items: [{ id: 9, name: 'Keep my proxy', protocol: 'http', host: 'localhost', port: 8080, status: 'active', fallback_mode: 'none', account_count: 0 }], total: 1, pages: 1
    } }))
    const wrapper = mount(ProxiesView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
      BaseDialog: true, ConfirmDialog: true, ImportDataModal: true, ProxyAdBanner: true, Icon: true,
    } } })
    try {
      await flushPromises()
      expect(wrapper.text()).toContain('Keep my proxy')
      malformed = true
      await wrapper.get('button[title="common.refresh"]').trigger('click')
      await flushPromises()
      expect(wrapper.text()).toContain('Keep my proxy')
      expect(showError).toHaveBeenCalledWith('admin.proxies.failedToLoad')
    } finally {
      wrapper.unmount()
      errorLog.mockRestore()
    }
  })
})
