import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import AuditLogView from '../AuditLogView.vue'

const { list, clear, getStatus, showError, showSuccess } = vi.hoisted(() => ({
  list: vi.fn(), clear: vi.fn(), getStatus: vi.fn(), showError: vi.fn(), showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({ adminAPI: { audit: { list, clear, get: vi.fn() } } }))
vi.mock('@/api', () => ({ totpAPI: { getStatus } }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string, params?: { count?: number }) => params?.count === undefined ? key : `${key}:${params.count}` })
}))

const wrappers: VueWrapper[] = []
function createView() {
  const wrapper = mount(AuditLogView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /><slot /></div>' },
        BaseDialog: {
          props: ['show', 'title'],
          template: '<section v-if="show"><h2>{{ title }}</h2><slot /><slot name="footer" /></section>'
        },
        DataTable: true, Pagination: true, Select: true, Icon: true
      }
    }
  })
  wrappers.push(wrapper)
  return wrapper
}

function clearButton(wrapper: VueWrapper) {
  const buttons = wrapper.findAll('button').filter(button => button.text() === 'admin.audit.clearAll')
  return buttons[buttons.length - 1]
}

async function enterTotp(wrapper: VueWrapper) {
  await flushPromises()
  await clearButton(wrapper).trigger('click')
  await flushPromises()
  expect(wrapper.text()).toContain('admin.audit.clearConfirm.title')
  await clearButton(wrapper).trigger('click')
  await wrapper.find('input[inputmode="numeric"]').setValue('123456')
}

beforeEach(() => {
  vi.clearAllMocks()
  list.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20 })
  getStatus.mockResolvedValue({ enabled: true })
  clear.mockResolvedValue({ deleted: 2 })
})

afterEach(() => {
  for (const wrapper of wrappers.splice(0)) wrapper.unmount()
})

describe('operation log clear-all', () => {
  it('checks TOTP, confirms, submits a fresh code and refreshes after success', async () => {
    const wrapper = createView()
    await enterTotp(wrapper)
    await clearButton(wrapper).trigger('click')
    await flushPromises()
    expect(getStatus).toHaveBeenCalledOnce()
    expect(clear).toHaveBeenCalledOnce()
    expect(clear).toHaveBeenCalledWith('123456')
    expect(showSuccess).toHaveBeenCalledWith('admin.audit.clearConfirm.success:2')
    expect(list).toHaveBeenCalledTimes(2)
    expect(wrapper.find('input[inputmode="numeric"]').exists()).toBe(false)
  })

  it('requires enabled TOTP before showing confirmation', async () => {
    getStatus.mockResolvedValue({ enabled: false })
    const wrapper = createView()
    await flushPromises()
    await clearButton(wrapper).trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenCalledWith('stepUp.notEnabled')
    expect(wrapper.text()).not.toContain('admin.audit.clearConfirm.title')
    expect(clear).not.toHaveBeenCalled()
  })

  it('keeps the verification prompt open and clears the code after failure', async () => {
    clear.mockRejectedValueOnce(new Error('cleanup failed'))
    const wrapper = createView()
    await enterTotp(wrapper)
    await clearButton(wrapper).trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenCalledWith('cleanup failed')
    const input = wrapper.find<HTMLInputElement>('input[inputmode="numeric"]')
    expect(input.element.value).toBe('')
    expect(list).toHaveBeenCalledOnce()
    await input.setValue('654321')
    await clearButton(wrapper).trigger('click')
    await flushPromises()
    expect(clear).toHaveBeenLastCalledWith('654321')
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('does not submit again when Enter is pressed during cleanup', async () => {
    let finish!: (value: { deleted: number }) => void
    clear.mockImplementation(() => new Promise(resolve => { finish = resolve }))
    const wrapper = createView()
    await enterTotp(wrapper)
    await clearButton(wrapper).trigger('click')
    try {
      await wrapper.find('input[inputmode="numeric"]').trigger('keyup.enter')
      expect(clear).toHaveBeenCalledOnce()
    } finally {
      finish({ deleted: 2 })
      await flushPromises()
    }
  })
})
