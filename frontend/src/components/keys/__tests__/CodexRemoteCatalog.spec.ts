import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
import UseKeyModal from '../UseKeyModal.vue'

function mountModal() {
  return mount(UseKeyModal, {
    props: { show: true, platform: 'openai', apiKey: 'sk-test', baseUrl: 'https://example.com/prefix/' },
    global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }, Icon: true } }
  })
}
const configs = (wrapper: ReturnType<typeof mountModal>) => wrapper.findAll('pre code').map(c => c.text()).join('\n')

describe('Codex remote catalogs', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('generates provider remote URL and switches to a local file for older clients', async () => {
    const wrapper = mountModal()
    expect(configs(wrapper)).toContain('model_catalog_url = "https://example.com/prefix/v1/models"')
    expect(configs(wrapper)).not.toContain('model_catalog_json')
    await wrapper.get('[data-testid="codex-model-catalog-mode"]').setValue('file')
    expect(configs(wrapper)).toContain('model_catalog_json = "~/.codex/codex-models.json"')
    expect(configs(wrapper)).not.toContain('model_catalog_url')
    wrapper.unmount()
  })

  it('keeps remote catalogs available at exactly 1 MiB', async () => {
    const raw = JSON.stringify({ models: [] }).padEnd(1024 * 1024, ' ')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(raw)))
    const wrapper = mountModal()
    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await flushPromises()
    expect(configs(wrapper)).toContain('model_catalog_url')
    expect(wrapper.get('option[value="remote"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).not.toContain('keys.useKeyModal.codexModelCatalog.oversized')
    wrapper.unmount()
  })

  it('falls back using UTF-8 response size and ignores a stale request from another group', async () => {
    const oversized = JSON.stringify({ models: [{ slug: 'model', description: '中'.repeat(360000) }] })
    const fetchMock = vi.fn().mockResolvedValue(new Response(oversized))
    vi.stubGlobal('fetch', fetchMock)
    const wrapper = mountModal()
    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await flushPromises()
    expect(configs(wrapper)).toContain('model_catalog_json')
    expect(wrapper.text()).toContain('keys.useKeyModal.codexModelCatalog.oversized')
    expect(wrapper.get('option[value="remote"]').attributes('disabled')).toBeDefined()
    expect(fetchMock.mock.calls[0][0]).toBe('https://example.com/prefix/v1/models?client_version=0.158.0')
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBe('Bearer sk-test')
    await wrapper.setProps({ baseUrl: 'https://second.example/v1' })
    await wrapper.get('[data-testid="codex-model-catalog-mode"]').setValue('remote')
    let resolveResponse!: (response: Response) => void
    fetchMock.mockImplementationOnce(() => new Promise<Response>(resolve => { resolveResponse = resolve }))
    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await wrapper.setProps({ apiKey: 'sk-next' })
    resolveResponse(new Response(oversized))
    await flushPromises()
    expect(configs(wrapper)).toContain('model_catalog_url')
    expect(wrapper.text()).not.toContain('keys.useKeyModal.codexModelCatalog.oversized')
    wrapper.unmount()
  })
})
