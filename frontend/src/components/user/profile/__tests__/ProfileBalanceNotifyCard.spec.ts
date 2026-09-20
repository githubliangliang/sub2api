import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ProfileBalanceNotifyCard from '../ProfileBalanceNotifyCard.vue'

const { sendNotifyEmailCode, verifyNotifyEmail, getProfile } = vi.hoisted(() => ({
  sendNotifyEmailCode: vi.fn(), verifyNotifyEmail: vi.fn(), getProfile: vi.fn(),
}))
vi.mock('@/api', () => ({ userAPI: { sendNotifyEmailCode, verifyNotifyEmail, getProfile } }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: null }) }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('ProfileBalanceNotifyCard pending email identity', () => {
  beforeEach(() => {
    vi.useFakeTimers(); vi.resetAllMocks(); sendNotifyEmailCode.mockResolvedValue({})
    getProfile.mockResolvedValue({ balance_notify_extra_emails: [] })
  })
  afterEach(() => { vi.clearAllTimers(); vi.useRealTimers() })

  it('removes verified pending emails by email identity', async () => {
    verifyNotifyEmail.mockResolvedValue({})
    const wrapper = mount(ProfileBalanceNotifyCard, {
      props: { enabled: true, threshold: null, extraEmails: [], systemDefaultThreshold: 5, userEmail: '' },
    })
    for (const email of ['a@example.com', 'b@example.com']) {
      await wrapper.get('input[type="email"]').setValue(email)
      await wrapper.findAll('button').find(button => button.text() === 'common.add')!.trigger('click')
    }
    for (const row of wrapper.findAll('.bg-yellow-50')) {
      await row.findAll('button').find(button => button.text() === 'profile.balanceNotify.sendCode')!.trigger('click')
    }
    await flushPromises()
    for (const row of wrapper.findAll('.bg-yellow-50')) await row.get('input').setValue('123456')
    await wrapper.findAll('.bg-yellow-50')[1].findAll('button').find(button => button.text() === 'profile.balanceNotify.verify')!.trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.bg-yellow-50').map(row => row.get('span').text())).toEqual(['a@example.com'])
    wrapper.unmount()
  })
})
