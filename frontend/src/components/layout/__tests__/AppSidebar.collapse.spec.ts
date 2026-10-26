import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const { route, app, auth, settings } = vi.hoisted(() => ({
  route: { path: '/admin/channels/pricing' },
  app: { sidebarCollapsed: false, mobileOpen: false, siteName: 'Test', siteLogo: '', siteVersion: '', publicSettingsLoaded: true, sidebarScrollTop: 0, cachedPublicSettings: { hidden_menu_keys: [] as string[], custom_menu_items: [] }, toggleSidebar: vi.fn(), setMobileOpen: vi.fn() },
  auth: { isAdmin: true, isSimpleMode: false },
  settings: { customMenuItems: [], hiddenMenuKeys: [] as string[], opsMonitoringEnabled: true, fetch: vi.fn().mockResolvedValue(undefined) },
}))
vi.mock('@/stores', () => ({ useAppStore: () => app, useAuthStore: () => auth, useAdminSettingsStore: () => settings, useOnboardingStore: () => ({ isCurrentStep: () => false }) }))
vi.mock('@/stores/app', () => ({ useAppStore: () => app }))
vi.mock('vue-router', () => ({ useRoute: () => route, useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async () => ({ ...await vi.importActual('vue-i18n'), useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useBatchImageAccess', () => ({ useBatchImageAccess: () => ({ canUseBatchImage: { value: false }, refreshBatchImageAccess: vi.fn().mockResolvedValue(undefined) }) }))
vi.mock('@/utils/userHomePathStore', () => ({ currentSignedInHomePath: () => '/dashboard' }))

import AppSidebar from '../AppSidebar.vue'

const renderSidebar = () => mount(AppSidebar, { global: { stubs: { VersionBadge: true, RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } } } })

describe('sidebar group choices', () => {
  beforeEach(() => {
    auth.isAdmin = true
    auth.isSimpleMode = false
    settings.opsMonitoringEnabled = true
    settings.hiddenMenuKeys = []
    app.cachedPublicSettings.hidden_menu_keys = []
  })

  it.each([false, true])('shows enabled ops to administrators in simple mode=%s', (simple) => {
    auth.isSimpleMode = simple
    const wrapper = renderSidebar()
    expect(wrapper.find('a[href="/admin/ops"]').exists()).toBe(true)
    if (simple) expect(wrapper.find('a[href="/admin/groups"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it.each(['disabled', 'public-hidden', 'admin-hidden', 'regular-user'])('keeps the ops menu hidden for %s', (reason) => {
    if (reason === 'disabled') settings.opsMonitoringEnabled = false
    if (reason === 'public-hidden') app.cachedPublicSettings.hidden_menu_keys = ['/admin/ops']
    if (reason === 'admin-hidden') settings.hiddenMenuKeys = ['/admin/ops']
    if (reason === 'regular-user') auth.isAdmin = false
    const wrapper = renderSidebar()
    expect(wrapper.find('a[href="/admin/ops"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('collapses and reopens a group while its child route stays active', async () => {
    const wrapper = renderSidebar()
    await flushPromises()
    expect(wrapper.find('a[href="/admin/channels/pricing"]').exists()).toBe(true)
    const group = wrapper.findAll('button').find(button => button.text().includes('nav.channelManagement'))!
    await group.trigger('click')
    expect(wrapper.find('a[href="/admin/channels/pricing"]').exists()).toBe(false)
    await group.trigger('click')
    expect(wrapper.find('a[href="/admin/channels/pricing"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('retains hidden-menu and simple-mode filtering', () => {
    app.cachedPublicSettings.hidden_menu_keys = ['/admin/channels']
    let wrapper = renderSidebar()
    expect(wrapper.text()).not.toContain('nav.channelManagement')
    wrapper.unmount()
    app.cachedPublicSettings.hidden_menu_keys = []
    auth.isSimpleMode = true
    wrapper = renderSidebar()
    expect(wrapper.find('a[href="/admin/groups"]').exists()).toBe(false)
    expect(wrapper.find('a[href="/admin/accounts"]').exists()).toBe(true)
    wrapper.unmount()
  })
})
