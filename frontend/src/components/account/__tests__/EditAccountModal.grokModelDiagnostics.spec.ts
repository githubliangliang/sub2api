import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import type { Account } from '@/types'

const { getAccount, updateAccount, syncModels, checkMixedChannel, showWarning } = vi.hoisted(() => ({
  getAccount: vi.fn(),
  updateAccount: vi.fn(),
  syncModels: vi.fn(),
  checkMixedChannel: vi.fn(),
  showWarning: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn(), showInfo: vi.fn(), showWarning })
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ isSimpleMode: true }) }))
vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: { getById: getAccount, update: updateAccount, syncUpstreamModels: syncModels, checkMixedChannelRisk: checkMixedChannel },
    settings: {
      getWebSearchEmulationConfig: vi.fn().mockResolvedValue({ enabled: false, providers: [] }),
      getSettings: vi.fn().mockResolvedValue({})
    },
    tlsFingerprintProfiles: { list: vi.fn().mockResolvedValue([]) }
  }
}))
vi.mock('@/api/admin/accounts', () => ({ getAntigravityDefaultModelMapping: vi.fn() }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

import EditAccountModal from '../EditAccountModal.vue'

const BaseDialog = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})
const ModelWhitelistSelector = defineComponent({
  name: 'ModelWhitelistSelector',
  props: ['modelValue', 'accountId'],
  emits: ['sync-result', 'sync-error', 'update:modelValue'],
  template: '<div />'
})

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => { resolve = done })
  return { promise, resolve }
}

function diagnostics() {
  return {
    state: 'fresh', endpoint: 'https://api.x.ai/v1/models',
    fetched_at: new Date().toISOString(), expires_at: new Date(Date.now() + 600_000).toISOString(),
    models: ['grok-4.6'],
    mappings: [{ requested: 'grok-4.6', upstream: 'grok-4.6', status: 'listed' }]
  }
}

function account(id = 5, catalog = diagnostics()): Account {
  return {
    id, name: `Grok ${id}`, platform: 'grok', type: 'oauth', notes: '',
    credentials: { model_mapping: { 'grok-4.6': 'grok-4.6' } },
    credentials_status: { has_access_token: true, has_refresh_token: true },
    extra: {}, proxy_id: null, concurrency: 1, priority: 1, rate_multiplier: 1,
    status: 'active', group_ids: [], expires_at: null, auto_pause_on_expired: false,
    grok_model_diagnostics: catalog
  } as unknown as Account
}

function mountModal(value = account()) {
  return mount(EditAccountModal, {
    props: { show: true, account: value, proxies: [], groups: [] },
    global: { stubs: { BaseDialog, Select: true, Icon: true, ProxySelector: true, GroupSelector: true, ModelWhitelistSelector } }
  })
}

describe('Grok diagnostics in the account editor', () => {
  beforeEach(() => {
    getAccount.mockReset().mockResolvedValue(account())
    updateAccount.mockReset()
    syncModels.mockReset()
    checkMixedChannel.mockReset().mockResolvedValue({ has_risk: false })
    showWarning.mockReset()
  })

  it('renders the saved detail diagnostics without fetching credentials or changing the form', async () => {
    const wrapper = mountModal()
    await flushPromises()
    expect(wrapper.find('[data-testid="grok-model-diagnostics"]').exists()).toBe(true)
    expect(getAccount).not.toHaveBeenCalled()
    expect(updateAccount).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('keeps the last diagnostics and shows failure when sync fails but detail reload succeeds', async () => {
    syncModels.mockRejectedValue(new Error('upstream unavailable'))
    const wrapper = mountModal()
    await wrapper.get('[data-testid="grok-diagnostics-sync"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="grok-diagnostics-refresh-failed"]').exists()).toBe(true)
    expect(getAccount).toHaveBeenCalledWith(5)
    expect(updateAccount).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('refreshes diagnostics without resetting an unsaved endpoint edit', async () => {
    const catalog = diagnostics()
    catalog.models = ['fresh-model']
    getAccount.mockResolvedValue(account(5, catalog))
    syncModels.mockResolvedValue({ models: ['fresh-model'] })
    const wrapper = mountModal()
    await wrapper.get('[data-testid="grok-custom-base-url-toggle"]').trigger('click')
    await wrapper.get('[data-testid="grok-custom-base-url-input"]').setValue('https://edited.example')
    await wrapper.get('[data-testid="grok-diagnostics-sync"]').trigger('click')
    await flushPromises()
    expect((wrapper.get('[data-testid="grok-custom-base-url-input"]').element as HTMLInputElement).value).toBe('https://edited.example')
    expect(wrapper.find('[data-testid="grok-diagnostics-dirty"]').exists()).toBe(true)
    expect(updateAccount).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('ignores an initial detail response after switching accounts', async () => {
    const pending = deferred<Account>()
    getAccount.mockReturnValueOnce(pending.promise)
    const initial = account()
    delete initial.grok_model_diagnostics
    const wrapper = mountModal(initial)
    const nextCatalog = diagnostics()
    nextCatalog.models = ['current-account-model']
    await wrapper.setProps({ account: account(6, nextCatalog) })
    const oldCatalog = diagnostics()
    oldCatalog.models = ['old-account-model']
    pending.resolve(account(5, oldCatalog))
    await flushPromises()
    await wrapper.get('[data-testid="grok-diagnostics-details-toggle"]').trigger('click')
    expect(wrapper.text()).toContain('current-account-model')
    expect(wrapper.text()).not.toContain('old-account-model')
    wrapper.unmount()
  })

  it('keeps the newest issued diagnostic read when responses finish in reverse order', async () => {
    const older = deferred<Account>()
    const newer = deferred<Account>()
    getAccount.mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise)
    const wrapper = mountModal()
    const selector = wrapper.findComponent(ModelWhitelistSelector)
    selector.vm.$emit('sync-result', { models: [] }, 5)
    selector.vm.$emit('sync-result', { models: [] }, 5)
    const current = diagnostics()
    current.models = ['newest-result']
    newer.resolve(account(5, current))
    await flushPromises()
    const previous = diagnostics()
    previous.models = ['older-result']
    older.resolve(account(5, previous))
    await flushPromises()
    await wrapper.get('[data-testid="grok-diagnostics-details-toggle"]').trigger('click')
    expect(wrapper.text()).toContain('newest-result')
    expect(wrapper.text()).not.toContain('older-result')
    wrapper.unmount()
  })

  it('ignores a late sync completion after closing and reopening another account', async () => {
    const pending = deferred<{ models: string[] }>()
    syncModels.mockReturnValue(pending.promise)
    const wrapper = mountModal()
    await wrapper.get('[data-testid="grok-diagnostics-sync"]').trigger('click')
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, account: account(6) })
    pending.resolve({ models: ['old-model'] })
    await flushPromises()
    expect(getAccount).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="grok-diagnostics-refresh-failed"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="grok-diagnostics-sync"]').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('refreshes an empty whitelist sync result and ignores notifications for another account', async () => {
    const wrapper = mountModal()
    const selector = wrapper.findComponent(ModelWhitelistSelector)
    selector.vm.$emit('sync-result', { models: [] }, 6)
    await flushPromises()
    expect(getAccount).not.toHaveBeenCalled()
    selector.vm.$emit('sync-result', { models: [] }, 5)
    await flushPromises()
    expect(getAccount).toHaveBeenCalledTimes(1)
    expect(getAccount).toHaveBeenCalledWith(5)
    wrapper.unmount()
  })
})
