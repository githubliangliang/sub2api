import { afterEach, describe, expect, it, vi } from 'vitest'
import { createApp, defineComponent, h, type App } from 'vue'
import { useTableLoader } from '@/composables/useTableLoader'

let app: App | undefined

afterEach(() => {
  app?.unmount()
  app = undefined
  document.body.innerHTML = ''
  vi.clearAllTimers()
  vi.useRealTimers()
})

describe('useTableLoader owner lifecycle', () => {
  it('does not start a queued debounced request after the owner unmounts', async () => {
    vi.useFakeTimers()
    const fetchFn = vi.fn().mockResolvedValue({ items: [], total: 0, pages: 0 })
    let scheduleReload!: () => Promise<void>
    app = createApp(defineComponent({
      setup() {
        const loader = useTableLoader({ fetchFn })
        scheduleReload = loader.debouncedReload
        return () => h('div')
      }
    }))
    const container = document.createElement('div')
    document.body.appendChild(container)
    app.mount(container)

    const pending = scheduleReload()
    app.unmount()
    app = undefined
    await vi.advanceTimersByTimeAsync(301)
    await pending

    expect(fetchFn).not.toHaveBeenCalled()
  })
})
