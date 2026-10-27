import { createApp, defineComponent, h } from 'vue'
import { createPinia } from 'pinia'
import i18n, { initI18n } from '@/i18n'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import '@/style.css'

async function mountAnnouncements() {
  const app = createApp(defineComponent({
    setup() {
      return () => h('main', { class: 'min-h-screen bg-gray-50 p-6' }, [
        h('h1', { class: 'mb-4 text-2xl font-bold' }, '公告'),
        h(AnnouncementBell),
        h(AnnouncementPopup)
      ])
    }
  }))
  const pinia = createPinia()
  app.use(pinia)
  useAppStore(pinia).initFromInjectedConfig()
  await initI18n()
  app.use(i18n)
  app.mount('#app')
  await useAnnouncementStore(pinia).fetchAnnouncements(true)
}

mountAnnouncements()
