import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ChannelsView from '../ChannelsView.vue'

const { listChannels, updateChannel, createChannel, getGroups, showError } = vi.hoisted(() => ({
  listChannels: vi.fn(), updateChannel: vi.fn(), createChannel: vi.fn(), getGroups: vi.fn(), showError: vi.fn()
}))

vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn() }) }))
vi.mock('@/api/admin', () => ({ adminAPI: {
  channels: { list: listChannels, update: updateChannel, create: createChannel },
  groups: { getAll: getGroups },
  accounts: { search: vi.fn().mockResolvedValue([]) },
  settings: { getWebSearchEmulationConfig: vi.fn().mockResolvedValue({ enabled: true, providers: [{}] }) }
} }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

const ToggleStub = defineComponent({
  props: ['modelValue'], emits: ['update:modelValue'],
  template: '<input type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />'
})
const channel = { id: 11, name: 'shared search', status: 'active', group_ids: [1, 2], model_pricing: [], model_mapping: {}, features_config: { web_search_emulation: { anthropic: true, openai: true } } }
function mountView() {
  return mount(ChannelsView, { global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    TablePageLayout: { template: '<div><slot name="filters"/><slot name="table"/><slot name="pagination"/></div>' },
    DataTable: { props: ['data'], emits: ['sort'], template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row"/></div></div>' },
    BaseDialog: { props: ['show'], template: '<div v-if="show"><slot/><slot name="footer"/></div>' },
    Toggle: ToggleStub, Select: true, Icon: true, PlatformIcon: true, Pagination: true, ConfirmDialog: true, PricingEntryCard: true
  } } })
}

describe('channel OpenAI Web Search inheritance', () => {
  beforeEach(() => {
    showError.mockReset()
    localStorage.clear()
    listChannels.mockReset().mockResolvedValue({ items: [channel], total: 1 })
    getGroups.mockReset().mockResolvedValue([{ id: 1, name: 'claude', platform: 'anthropic' }, { id: 2, name: 'search-group', platform: 'openai' }])
    updateChannel.mockReset().mockResolvedValue(channel)
    createChannel.mockReset().mockResolvedValue(channel)
  })

  it('reads and saves OpenAI independently while retaining the Anthropic flag', async () => {
    const wrapper = mountView()
    await flushPromises()
    const edit = wrapper.findAll('button').find(button => button.text() === 'common.edit')!
    await edit.trigger('click')
    await flushPromises()
    const toggle = wrapper.get<HTMLInputElement>('[data-testid="channel-web-search-openai"]')
    expect(toggle.element.checked).toBe(true)
    expect(wrapper.get<HTMLInputElement>('[data-testid="channel-web-search-anthropic"]').element.checked).toBe(true)
    await toggle.setValue(false)
    await wrapper.get('form#channel-form').trigger('submit.prevent')
    await flushPromises()
    expect(updateChannel).toHaveBeenCalledWith(11, expect.objectContaining({
      features_config: expect.objectContaining({ web_search_emulation: { anthropic: true, openai: false } })
    }))
    wrapper.unmount()
  })

  it('creates an OpenAI channel with an enabled search flag', async () => {
    const wrapper = mountView()
    await flushPromises()
    listChannels.mockResolvedValueOnce({ items: [], total: 0 })
    await wrapper.findAll('button').find(button => button.text().includes('admin.channels.createChannel'))!.trigger('click')
    await flushPromises()
    await wrapper.get('form#channel-form input[type="text"]').setValue('new search channel')
    const openai = wrapper.findAll('label').find(label => label.text() === 'admin.groups.platforms.openai')!
    await openai.get('input').setValue(true)
    await flushPromises()
    await wrapper.get('[data-testid="channel-web-search-openai"]').setValue(true)
    const group = wrapper.findAll('label').find(label => label.text().startsWith('search-group'))!
    expect(group.get('input').attributes('disabled')).toBeUndefined()
    await group.get('input').setValue(true)
    await wrapper.get('form#channel-form').trigger('submit.prevent')
    await flushPromises()
    expect(showError).not.toHaveBeenCalled()
    expect(createChannel).toHaveBeenCalledWith(expect.objectContaining({
      group_ids: [2], features_config: expect.objectContaining({ web_search_emulation: { openai: true } })
    }))
    wrapper.unmount()
  })
})
