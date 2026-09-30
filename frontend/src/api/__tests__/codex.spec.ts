import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  buildCodexModelsManifestUrl,
  buildCodexModelCatalogUrl,
  fetchCodexModelsManifest
} from '../codex'

describe('Codex models API', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('builds the authenticated Codex manifest endpoint from the public API base', () => {
    expect(buildCodexModelsManifestUrl('https://example.com/api/v1/')).toBe(
      'https://example.com/api/v1/models?client_version=0.158.0'
    )
  })

  it.each([
    ['https://example.com', 'https://example.com/v1/models'],
    ['https://example.com/v1/', 'https://example.com/v1/models'],
    ['https://example.com/prefix/', 'https://example.com/prefix/v1/models'],
    ['https://example.com/prefix/v1/', 'https://example.com/prefix/v1/models']
  ])('normalizes the remote catalog URL %s', (base, expected) => {
    expect(buildCodexModelCatalogUrl(base)).toBe(expected)
  })

  it('measures raw UTF-8 bytes while preserving the full manifest for download', async () => {
    const raw = '{"models":[{"slug":"测试","description":"你好"}]}'
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(raw)))
    const result = await fetchCodexModelsManifest('https://example.com', 'sk-test')
    expect(result.responseBytes).toBe(new TextEncoder().encode(raw).byteLength)
    expect(result.responseBytes).toBeGreaterThan(raw.length)
    expect(JSON.parse(result.content)).toEqual(JSON.parse(raw))
  })

  it('fetches a manifest with the current API key without adding it to the catalog', async () => {
    const manifest = {
      models: [
        {
          slug: 'grok-4.6',
          default_reasoning_level: 'high',
          supported_reasoning_levels: [
            { effort: 'low', description: 'Fast responses' },
            { effort: 'xhigh', description: 'Extra-high reasoning depth' }
          ],
          input_modalities: ['text', 'image'],
          model_messages: { instructions_template: 'Use the routed model.' }
        },
        {
          slug: 'deepseek-v4-pro',
          default_reasoning_level: 'high',
          supported_reasoning_levels: [
            { effort: 'low', description: 'Fast responses' },
            { effort: 'max', description: 'Maximum reasoning depth' }
          ],
          input_modalities: ['text'],
          model_messages: { instructions_template: 'Use the routed model.' }
        }
      ]
    }
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => JSON.stringify(manifest)
    })
    vi.stubGlobal('fetch', fetchMock)

    const result = await fetchCodexModelsManifest('https://example.com/v1', 'sk-user-test')

    expect(fetchMock).toHaveBeenCalledWith(
      'https://example.com/v1/models?client_version=0.158.0',
      expect.objectContaining({
        headers: {
          Accept: 'application/json',
          Authorization: 'Bearer sk-user-test'
        }
      })
    )
    expect(result.modelCount).toBe(2)
    expect(JSON.parse(result.content)).toEqual(manifest)
    expect(result.content).toContain('"effort": "xhigh"')
    expect(result.content).toContain('"input_modalities"')
    expect(result.content).toContain('"instructions_template"')
    expect(result.content).not.toContain('sk-user-test')
  })

  it('rejects a successful response that is not a Codex manifest', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ object: 'list', data: [] })
    }))

    await expect(fetchCodexModelsManifest('https://example.com/v1', 'sk-user-test'))
      .rejects.toThrow('valid manifest')
  })
})
