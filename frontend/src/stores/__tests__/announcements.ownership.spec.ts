import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { useAnnouncementStore } from '../announcements'
import type { UserAnnouncement } from '@/types'

const api = vi.hoisted(() => ({ list: vi.fn(), markRead: vi.fn() }))
vi.mock('@/api', () => ({ announcementsAPI: api }))
beforeEach(() => { setActivePinia(createPinia()); vi.resetAllMocks() })
afterEach(() => { vi.restoreAllMocks(); vi.useRealTimers() })
const notice = (id: number): UserAnnouncement => ({ id, title: `Notice ${id}`, content: 'Details', notify_mode: 'popup', created_at: '', updated_at: '' })

it('retains a popup until its read acknowledgement succeeds', async () => {
  vi.spyOn(console, 'error').mockImplementation(() => {})
  const store = useAnnouncementStore()
  store.announcements = [notice(1)]
  store.currentPopup = store.announcements[0]
  api.markRead.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce(undefined)
  await store.dismissPopup()
  expect(store.currentPopup?.id).toBe(1)
  expect(store.unreadCount).toBe(1)
  await store.dismissPopup()
  expect(store.currentPopup).toBeNull()
  expect(store.unreadCount).toBe(0)
})

it('discards a previous session fetch and read completion', async () => {
  const store = useAnnouncementStore()
  let finishFetch!: (value: UserAnnouncement[]) => void
  let finishRead!: () => void
  api.list.mockReturnValueOnce(new Promise(resolve => { finishFetch = resolve }))
  api.markRead.mockReturnValueOnce(new Promise<void>(resolve => { finishRead = resolve }))
  const fetch = store.fetchAnnouncements()
  const read = store.markAsRead(1)
  store.reset()
  store.announcements = [notice(1)]
  finishFetch([notice(9)])
  finishRead()
  await Promise.all([fetch, read])
  expect(store.announcements.map(item => item.id)).toEqual([1])
  expect(store.unreadCount).toBe(1)
  expect(store.currentPopup).toBeNull()
})

it('cancels a queued popup transition when the session resets', async () => {
  vi.useFakeTimers()
  api.list.mockResolvedValue([notice(1), notice(2)])
  api.markRead.mockResolvedValue(undefined)
  const store = useAnnouncementStore()
  await store.fetchAnnouncements()
  await store.dismissPopup()
  store.reset()
  await vi.advanceTimersByTimeAsync(300)
  expect(store.currentPopup).toBeNull()
})
