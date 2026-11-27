import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ModelPlazaView from '../ModelPlazaView.vue'

const { listAccountModels, updateReferencePricing, copyToClipboard, showError } = vi.hoisted(() => ({
  listAccountModels: vi.fn(),
  updateReferencePricing: vi.fn(),
  copyToClipboard: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    modelPlaza: {
      listAccountModels,
      updateReferencePricing
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess: vi.fn()
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function sampleItems() {
  return {
    items: [
      {
        client_id: 'gpt-5.4',
        platform: 'openai',
        upstream_ids: ['gpt-5.4'],
        account_count: 1
      },
      {
        client_id: 'grok-4.5',
        platform: 'grok',
        upstream_ids: ['grok-4.5'],
        account_count: 1
      }
    ]
  }
}

function mountView() {
  return mount(ModelPlazaView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /></div>'
        },
        DataTable: {
          props: ['data', 'columns', 'loading'],
          template: `
            <div>
              <div v-if="!data || data.length === 0"><slot name="empty" /></div>
              <div v-for="row in data || []" :key="row.rowKey" class="row">
                <slot name="cell-client_id" :row="row" :value="row.client_id" />
                <slot name="cell-upstream_ids" :row="row" :value="row.upstream_ids" />
                <slot name="cell-platform" :row="row" :value="row.platform" />
                <slot name="cell-pricing" :row="row" :value="row.pricing" />
                <slot name="cell-account_count" :row="row" :value="row.account_count" />
              </div>
            </div>
          `
        },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue'],
          template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option></select>'
        },
        Icon: true,
        BaseDialog: {
          props: ['show', 'title'],
          template: '<div v-if="show" role="dialog"><h3>{{ title }}</h3><slot /><slot name="footer" /></div>'
        }
      }
    }
  })
}

describe('Admin ModelPlazaView', () => {
  beforeEach(() => {
    listAccountModels.mockReset()
    updateReferencePricing.mockReset()
    copyToClipboard.mockReset()
    showError.mockReset()
  })

  it('renders aggregated account models and copies the client ID', async () => {
    listAccountModels.mockResolvedValue(sampleItems())
    const wrapper = mountView()
    await flushPromises()

    expect(listAccountModels).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('gpt-5.4')
    expect(wrapper.text()).toContain('grok-4.5')

    await wrapper.get('button[title="admin.modelPlaza.copyClientId"]').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('gpt-5.4')
  })

  it('filters by search query', async () => {
    listAccountModels.mockResolvedValue(sampleItems())
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input').setValue('grok')
    expect(wrapper.text()).toContain('grok-4.5')
    expect(wrapper.text()).not.toContain('gpt-5.4')
  })

  it('shows empty copy when there are no enabled-account models', async () => {
    listAccountModels.mockResolvedValue({ items: [] })
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('admin.modelPlaza.empty')
  })

  it('saves missing reference prices in per-token units and refreshes the list', async () => {
    listAccountModels.mockResolvedValueOnce(sampleItems()).mockResolvedValue({ items: [
      { ...sampleItems().items[0], pricing: {
        model: 'gpt-5.4', manual: true, input_price: 0, output_price: 0.0000125,
        cache_read_price: 0.00000025, cache_write_price: null, cache_write_1h_price: null
      } }
    ] })
    updateReferencePricing.mockResolvedValue(undefined)
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[title="admin.modelPlaza.addPrice"]').trigger('click')
    expect(wrapper.get('[role="dialog"]').text()).toContain('gpt-5.4')
    await wrapper.get('#plaza-price-input_price').setValue('0')
    await wrapper.get('#plaza-price-output_price').setValue('12.5')
    await wrapper.get('#plaza-price-cache_read_price').setValue('0.25')
    await wrapper.get('#plaza-price-form').trigger('submit')
    await flushPromises()

    expect(updateReferencePricing).toHaveBeenCalledWith({
      client_id: 'gpt-5.4',
      platform: 'openai',
      pricing: {
        input_price: 0,
        output_price: 0.0000125,
        cache_read_price: 0.00000025,
        cache_write_price: null,
        cache_write_1h_price: null
      }
    })
    expect(listAccountModels).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('$12.50')
    expect(wrapper.text()).toContain('admin.modelPlaza.manualPrice')
  })

  it('edits and clears a manually supplied price while retaining automatic prices', async () => {
    listAccountModels.mockResolvedValue({ items: [
      { ...sampleItems().items[0], pricing: {
        model: 'gpt-5.4', manual: true, input_price: 0.0000025, output_price: 0,
        cache_read_price: 0, cache_write_price: null, cache_write_1h_price: null
      } },
      { ...sampleItems().items[1], pricing: {
        model: 'grok-4.5', input_price: 0.000003, output_price: 0.000015,
        cache_read_price: null, cache_write_price: null
      } }
    ] })
    updateReferencePricing.mockResolvedValue(undefined)
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('button[title="admin.modelPlaza.addPrice"]').exists()).toBe(false)
    expect(wrapper.findAll('button[title="admin.modelPlaza.editPrice"]')).toHaveLength(1)
    await wrapper.get('button[title="admin.modelPlaza.editPrice"]').trigger('click')
    expect((wrapper.get('#plaza-price-input_price').element as HTMLInputElement).value).toBe('2.5')
    expect((wrapper.get('#plaza-price-output_price').element as HTMLInputElement).value).toBe('0')
    expect((wrapper.get('#plaza-price-cache_read_price').element as HTMLInputElement).value).toBe('0')
    expect((wrapper.get('#plaza-price-cache_write_price').element as HTMLInputElement).value).toBe('')
    await wrapper.get('button[data-testid="clear-reference-price"]').trigger('click')
    await flushPromises()

    expect(updateReferencePricing).toHaveBeenCalledWith({
      client_id: 'gpt-5.4', platform: 'openai', pricing: null
    })
  })

  it('validates required and nonnegative prices and retains the form after a save failure', async () => {
    listAccountModels.mockResolvedValue(sampleItems())
    updateReferencePricing.mockRejectedValue(new Error('save failed'))
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('button[title="admin.modelPlaza.addPrice"]').trigger('click')
    await wrapper.get('#plaza-price-form').trigger('submit')
    expect(updateReferencePricing).not.toHaveBeenCalled()
    await wrapper.get('#plaza-price-input_price').setValue('-1')
    await wrapper.get('#plaza-price-output_price').setValue('3')
    await wrapper.get('#plaza-price-form').trigger('submit')
    expect(updateReferencePricing).not.toHaveBeenCalled()
    await wrapper.get('#plaza-price-input_price').setValue('1')
    await wrapper.get('#plaza-price-form').trigger('submit')
    await flushPromises()
    expect(showError).toHaveBeenLastCalledWith('admin.modelPlaza.savePriceFailed')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect((wrapper.get('#plaza-price-input_price').element as HTMLInputElement).value).toBe('1')
  })
})
