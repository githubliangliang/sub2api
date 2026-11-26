import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ImageUpload from '../ImageUpload.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

enableAutoUnmount(afterEach)

class ControlledReader {
  static instances: ControlledReader[] = []
  onload: ((event: { target: { result: string } }) => void) | null = null
  onerror: (() => void) | null = null
  readyState = 1
  abort = vi.fn(() => { this.readyState = 2 })
  readAsText = vi.fn()
  readAsDataURL = vi.fn()

  constructor() { ControlledReader.instances.push(this) }

  // Deliberately deliver callbacks after abort to cover already queued events too.
  complete(result: string) {
    this.readyState = 2
    this.onload?.({ target: { result } })
  }

  fail() {
    this.readyState = 2
    this.onerror?.()
  }
}

beforeEach(() => {
  ControlledReader.instances = []
  vi.stubGlobal('FileReader', ControlledReader)
})
afterEach(() => vi.unstubAllGlobals())

function uploader(mode: 'image' | 'svg' = 'image', maxSize?: number) {
  const wrapper = mount(ImageUpload, {
    props: { modelValue: 'existing', mode, ...(maxSize === undefined ? {} : { maxSize }) },
    global: { stubs: { Icon: true } },
  })
  const input = wrapper.get('input[type="file"]')
  return {
    wrapper,
    async choose(type = 'image/png', contents = 'image') {
      const file = new File([contents], type === 'image/svg+xml' ? 'upload.svg' : 'upload.png', { type })
      Object.defineProperty(input.element, 'files', { configurable: true, value: [file] })
      await input.trigger('change')
      expect((input.element as HTMLInputElement).value).toBe('')
      return file
    },
  }
}

describe('shared image upload read ownership', () => {
  it.each(['image', 'svg'] as const)('keeps the newest %s selection when reads finish out of order', async mode => {
    const { wrapper, choose } = uploader(mode)
    await choose()
    await choose()
    const [old, current] = ControlledReader.instances
    expect(old.abort).toHaveBeenCalledOnce()
    current.complete('current-image')
    old.complete('old-image')
    old.fail()
    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toEqual([['current-image']])
    expect(wrapper.text()).not.toContain('common.fileReadFailed')
  })

  it('does not let an old error replace the current read state', async () => {
    const { wrapper, choose } = uploader()
    await choose()
    await choose()
    ControlledReader.instances[0].fail()
    await nextTick()
    expect(wrapper.text()).not.toContain('common.fileReadFailed')
    ControlledReader.instances[1].complete('current-image')
    expect(wrapper.emitted('update:modelValue')).toEqual([['current-image']])
  })

  it('does not restore an image or an error after removal', async () => {
    const { wrapper, choose } = uploader()
    await choose()
    await wrapper.get('button').trigger('click')
    const reader = ControlledReader.instances[0]
    expect(reader.abort).toHaveBeenCalledOnce()
    reader.complete('removed-image')
    reader.fail()
    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toEqual([['']])
    expect(wrapper.text()).not.toContain('common.fileReadFailed')
  })

  it('cancels a pending read and ignores queued callbacks after unmount', async () => {
    const { wrapper, choose } = uploader()
    await choose()
    wrapper.unmount()
    const reader = ControlledReader.instances[0]
    expect(reader.abort).toHaveBeenCalledOnce()
    reader.complete('unmounted-image')
    reader.fail()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it.each(['type', 'size'] as const)('rejects a new invalid %s without restoring the previous pending read', async invalid => {
    const { wrapper, choose } = uploader('image', 10)
    await choose()
    await choose(invalid === 'type' ? 'text/plain' : 'image/png', invalid === 'size' ? 'x'.repeat(11) : 'image')
    const reader = ControlledReader.instances[0]
    expect(reader.abort).toHaveBeenCalledOnce()
    expect(ControlledReader.instances).toHaveLength(1)
    reader.complete('old-image')
    reader.fail()
    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.text()).toContain(invalid === 'type' ? 'common.selectImageFile' : 'common.fileTooLargeKb')
    expect(wrapper.text()).not.toContain('common.fileReadFailed')
  })

  it('keeps the 300KiB default bound and accepts a file exactly at the bound', async () => {
    const { wrapper, choose } = uploader()
    await choose('image/png', 'x'.repeat(300 * 1024))
    expect(ControlledReader.instances).toHaveLength(1)
    ControlledReader.instances[0].complete('accepted-image')
    await choose('image/png', 'x'.repeat(300 * 1024 + 1))
    expect(ControlledReader.instances).toHaveLength(1)
    expect(wrapper.emitted('update:modelValue')).toEqual([['accepted-image']])
    expect(wrapper.text()).toContain('common.fileTooLargeKb')
  })

  it('reports the active read error and lets a subsequent image selection recover', async () => {
    const { wrapper, choose } = uploader()
    const file = await choose()
    expect(ControlledReader.instances[0].readAsDataURL).toHaveBeenCalledWith(file)
    ControlledReader.instances[0].fail()
    await nextTick()
    expect(wrapper.text()).toContain('common.fileReadFailed')
    await choose()
    ControlledReader.instances[1].complete('data:image/png;base64,current')
    await nextTick()
    expect(wrapper.text()).not.toContain('common.fileReadFailed')
    expect(wrapper.emitted('update:modelValue')).toEqual([['data:image/png;base64,current']])
  })

  it('retains SVG text trimming and sanitizes the inline preview', async () => {
    const { wrapper, choose } = uploader('svg')
    expect(wrapper.get('input').attributes('accept')).toBe('.svg')
    const file = await choose('image/svg+xml')
    expect(ControlledReader.instances[0].readAsText).toHaveBeenCalledWith(file)
    const svg = '<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"><script>alert(1)</script><path d="M0 0h1"/></svg>'
    ControlledReader.instances[0].complete(`  ${svg}\n`)
    expect(wrapper.emitted('update:modelValue')).toEqual([[svg]])
    await wrapper.setProps({ modelValue: svg })
    const preview = wrapper.get('span').html()
    expect(preview).toContain('<path')
    expect(preview).not.toContain('<script')
    expect(preview).not.toContain('onload')
  })
})
