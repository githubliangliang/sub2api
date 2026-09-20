import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import SubscriptionsView from '../SubscriptionsView.vue'

const { listSubscriptions, assignSubscription, getAllGroups, listUsers, searchUsageUsers, showError } = vi.hoisted(() => ({
  listSubscriptions: vi.fn(),
  assignSubscription: vi.fn(),
  showError: vi.fn(),
  getAllGroups: vi.fn(),
  listUsers: vi.fn(),
  searchUsageUsers: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptions: { list: listSubscriptions, assign: assignSubscription },
    groups: { getAll: getAllGroups },
    users: { list: listUsers },
    usage: { searchUsers: searchUsageUsers }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess: vi.fn() })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: { to: { type: Object, required: true } },
  template: '<a><slot /></a>'
})

describe('admin subscription assignment search', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listSubscriptions.mockResolvedValue({ items: [], total: 0, pages: 1 })
    assignSubscription.mockResolvedValue({})
    getAllGroups.mockResolvedValue([])
    listUsers.mockResolvedValue({ items: [{ id: 42, email: 'reader@example.com' }], total: 1, pages: 1 })
    searchUsageUsers.mockResolvedValue([{ id: 42, email: 'reader@example.com' }])
  })

  it('clears the assignment target before the debounced search runs', async () => {
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] })
    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
          DataTable: { template: '<div />' },
          RouterLink: RouterLinkStub,
          Pagination: true,
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Icon: true,
          Teleport: true
        }
      }
    })
    try {
      await flushPromises()
      await wrapper.findAll('button').find((button) => button.text() === 'admin.subscriptions.assignSubscription')!.trigger('click')
      const form = wrapper.get('#assign-subscription-form')
      form.getComponent({ name: 'Select' }).vm.$emit('update:modelValue', 3)
      const search = wrapper.get('[data-assign-user-search] input')
      await search.trigger('focus')
      await search.setValue('reader')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()
      await wrapper.get('[data-assign-user-search] button').trigger('click')

      await search.setValue('another')
      await form.trigger('submit')
      await flushPromises()

      expect(assignSubscription).not.toHaveBeenCalled()
      expect(showError).toHaveBeenCalledWith('admin.subscriptions.pleaseSelectUser')
    } finally {
      wrapper.unmount()
      vi.useRealTimers()
    }
  })

  it('searches active users instead of usage history when assigning', async () => {
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] })
    searchUsageUsers.mockResolvedValue([{ id: 14, email: 'deleted@example.com' }])
    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
          DataTable: { template: '<div />' },
          RouterLink: RouterLinkStub,
          Pagination: true,
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Icon: true,
          Teleport: true
        }
      }
    })
    try {
      await flushPromises()
      await wrapper.findAll('button').find((button) => button.text() === 'admin.subscriptions.assignSubscription')!.trigger('click')
      const search = wrapper.get('[data-assign-user-search] input')
      await search.trigger('focus')
      await search.setValue('  example.com  ')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(listUsers).toHaveBeenCalledWith(1, 30, {
        search: 'example.com', sort_by: 'email', sort_order: 'asc'
      })
      expect(searchUsageUsers).not.toHaveBeenCalled()
      expect(wrapper.get('[data-assign-user-search]').text()).toContain('reader@example.com')
      expect(wrapper.get('[data-assign-user-search]').text()).not.toContain('deleted@example.com')
    } finally {
      wrapper.unmount()
      vi.useRealTimers()
    }
  })
})
