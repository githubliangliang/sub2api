import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { announcementsAPI } from '@/api'
import type { UserAnnouncement } from '@/types'

const THROTTLE_MS = 20 * 60 * 1000 // 20 minutes

export const useAnnouncementStore = defineStore('announcements', () => {
  // State
  const announcements = ref<UserAnnouncement[]>([])
  const loading = ref(false)
  const lastFetchTime = ref(0)
  const popupQueue = ref<UserAnnouncement[]>([])
  const currentPopup = ref<UserAnnouncement | null>(null)

  // Session-scoped dedup set — not reactive, used as plain lookup only
  let shownPopupIds = new Set<number>()
  let stateVersion = 0
  let fetchVersion = 0
  let loadingVersion = 0
  let dismissingId: number | null = null
  let popupTimer: ReturnType<typeof setTimeout> | undefined

  // Getters
  const unreadCount = computed(() =>
    announcements.value.filter((a) => !a.read_at).length
  )

  // Actions
  async function fetchAnnouncements(force = false) {
    const now = Date.now()
    if (!force && lastFetchTime.value > 0 && now - lastFetchTime.value < THROTTLE_MS) {
      return
    }

    const version = stateVersion
    const request = ++fetchVersion
    const loadingRequest = ++loadingVersion
    // Set immediately to prevent concurrent duplicate requests
    lastFetchTime.value = now

    try {
      loading.value = true
      const all = await announcementsAPI.list(false)
      if (version !== stateVersion || request !== fetchVersion) return
      announcements.value = all.slice(0, 20)
      enqueueNewPopups()
    } catch (err: any) {
      if (version !== stateVersion || request !== fetchVersion) return
      // Revert throttle timestamp on failure so retry is allowed
      lastFetchTime.value = 0
      console.error('Failed to fetch announcements:', err)
    } finally {
      if (version === stateVersion && loadingRequest === loadingVersion) loading.value = false
    }
  }

  function enqueueNewPopups() {
    const newPopups = announcements.value.filter(
      (a) => a.notify_mode === 'popup' && !a.read_at && !shownPopupIds.has(a.id)
    )
    if (newPopups.length === 0) return

    for (const p of newPopups) {
      if (!popupQueue.value.some((q) => q.id === p.id)) {
        popupQueue.value.push(p)
      }
    }

    if (!currentPopup.value) {
      showNextPopup()
    }
  }

  function showNextPopup() {
    if (popupQueue.value.length === 0) {
      currentPopup.value = null
      return
    }
    currentPopup.value = popupQueue.value.shift()!
    shownPopupIds.add(currentPopup.value.id)
  }

  async function dismissPopup() {
    const popup = currentPopup.value
    if (!popup || dismissingId === popup.id) return
    const version = stateVersion
    dismissingId = popup.id
    const marked = await markAsRead(popup.id)
    if (version !== stateVersion) return
    dismissingId = null
    if (!marked || currentPopup.value?.id !== popup.id) return
    currentPopup.value = null
    if (popupQueue.value.length > 0) {
      popupTimer = setTimeout(() => {
        if (version === stateVersion) showNextPopup()
      }, 300)
    }
  }

  async function markAsRead(id: number) {
    const version = stateVersion
    try {
      await announcementsAPI.markRead(id)
      if (version !== stateVersion) return false
      const ann = announcements.value.find((a) => a.id === id)
      if (ann) {
        ann.read_at = new Date().toISOString()
      }
      return true
    } catch (err: any) {
      if (version !== stateVersion) return false
      console.error('Failed to mark announcement as read:', err)
      return false
    }
  }

  async function markAllAsRead() {
    const unread = announcements.value.filter((a) => !a.read_at)
    if (unread.length === 0) return true
    const version = stateVersion
    const loadingRequest = ++loadingVersion
    const unreadIds = new Set(unread.map(item => item.id))

    try {
      loading.value = true
      await Promise.all(unread.map((a) => announcementsAPI.markRead(a.id)))
      if (version !== stateVersion) return false
      announcements.value.forEach((a) => {
        if (unreadIds.has(a.id) && !a.read_at) {
          a.read_at = new Date().toISOString()
        }
      })
      return true
    } catch (err: any) {
      if (version !== stateVersion) return false
      console.error('Failed to mark all as read:', err)
      throw err
    } finally {
      if (version === stateVersion && loadingRequest === loadingVersion) loading.value = false
    }
  }

  function reset() {
    stateVersion++
    fetchVersion++
    loadingVersion++
    clearTimeout(popupTimer)
    dismissingId = null
    announcements.value = []
    lastFetchTime.value = 0
    shownPopupIds = new Set()
    popupQueue.value = []
    currentPopup.value = null
    loading.value = false
  }

  return {
    // State
    announcements,
    loading,
    currentPopup,
    // Getters
    unreadCount,
    // Actions
    fetchAnnouncements,
    dismissPopup,
    markAsRead,
    markAllAsRead,
    reset,
  }
})
