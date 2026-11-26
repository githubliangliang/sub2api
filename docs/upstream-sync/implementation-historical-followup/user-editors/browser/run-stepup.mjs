import { chromium } from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs'
import fs from 'node:fs/promises'

const out = new URL('./stepup/', import.meta.url)
await fs.mkdir(out, { recursive: true })
const base = 'http://127.0.0.1:18926'
const results = [], requests = [], errors = []
const assert = (condition, message) => { if (!condition) throw new Error(message) }
const deferred = () => { let resolve; const promise = new Promise(res => { resolve = res }); return { promise, resolve, seen: false } }
const admin = { id: 99, email: 'admin@browser.invalid', username: 'fixture-admin', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' }
const initialUsers = [1, 2].map(id => ({ id, email: `user${id}@browser.invalid`, username: `fixture-${id}`, role: 'user', status: 'active', balance: 100, concurrency: 1, rpm_limit: 0, allowed_groups: id === 1 ? [7] : [], group_rates: id === 1 ? { 7: 0.5, 8: 0.8 } : {}, subscriptions: [], created_at: '2026-11-26T00:00:00Z', updated_at: '2026-11-26T00:00:00Z' }))
const initialDefinitions = [
  { id: 1, key: 'count', name: 'Count', type: 'number', validation: { min: -10, max: 100 }, required: false, enabled: true, description: 'Count saved description', placeholder: 'Count fixture' },
  { id: 2, key: 'note', name: 'Note', type: 'text', required: false, enabled: true, description: 'Note saved description', placeholder: 'Note fixture' }
]
const initialGroups = [
  { id: 7, name: 'Exclusive fixture', platform: 'openai', is_exclusive: true, subscription_type: 'standard', status: 'active', rate_multiplier: 1 },
  { id: 8, name: 'Public fixture', platform: 'openai', is_exclusive: false, subscription_type: 'standard', status: 'active', rate_multiplier: 1 }
]
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
try {
  for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
    const users = structuredClone(initialUsers), definitions = structuredClone(initialDefinitions), groups = structuredClone(initialGroups)
    const values = { 1: { 1: '12', 2: 'user-1-original' }, 2: { 1: '22', 2: 'user-2-original' } }
    let holdValues = null, holdGroups = null, holdUserSave = null, failGroups = false, holdVerification = null, verificationGrant = false
    const context = await browser.newContext({ viewport, locale: 'en-US' })
    await context.addInitScript(admin => {
      localStorage.setItem('auth_token', 'synthetic-browser-only')
      localStorage.setItem('auth_user', JSON.stringify(admin))
      localStorage.setItem('sub2api_locale', 'en')
      localStorage.setItem('admin_guide_99_admin_v4_interactive', 'true')
      localStorage.setItem('token_expires_at', String(Date.now() + 3600000))
    }, admin)
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname, method = request.method()
      const body = request.postData() ? request.postDataJSON() : undefined
      requests.push({ viewport: viewport.name, method, url: request.url(), path, body })
      let data = {}
      if (path.endsWith('/auth/me')) data = admin
      else if (path.endsWith('/settings/public')) data = { site_name: 'Historical Followup Synthetic Browser', run_mode: 'standard', registration_enabled: false }
      else if (path.includes('compliance')) data = { required: false }
      else if (path === '/api/v1/user/totp/step-up') {
        if (holdVerification) { const wait = holdVerification; holdVerification = null; wait.seen = true; await wait.promise }
        verificationGrant = true; data = { verified: true, expires_in: 300 }
      }
      else if (path === '/api/v1/admin/users') data = { items: users, total: users.length, page: 1, page_size: 20, pages: 1 }
      else if (path === '/api/v1/admin/user-attributes') data = definitions
      else if (path === '/api/v1/admin/user-attributes/batch') data = { attributes: values }
      else if (/\/admin\/user-attributes\/\d+$/.test(path)) {
        const id = Number(path.split('/').at(-1)); Object.assign(definitions.find(d => d.id === id), body); data = definitions.find(d => d.id === id)
      } else if (/\/admin\/users\/\d+\/attributes$/.test(path)) {
        const id = Number(path.split('/').at(-2))
        if (method === 'PUT') { values[id] = structuredClone(body.values); data = { message: 'saved synthetic attributes' } }
        else {
          const snapshot = Object.entries(values[id]).map(([attribute_id, value]) => ({ attribute_id: Number(attribute_id), value }))
          if (holdValues) { const wait = holdValues; holdValues = null; wait.seen = true; await wait.promise }
          data = snapshot
        }
      } else if (/\/admin\/users\/\d+$/.test(path)) {
        const id = Number(path.split('/').at(-1)), user = users.find(u => u.id === id)
        if (method === 'PUT') {
          if (holdUserSave) { const wait = holdUserSave; holdUserSave = null; wait.seen = true; await wait.promise }
          if (body.role === 'admin' && !verificationGrant) { await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ code: 'STEP_UP_REQUIRED', message: 'Synthetic step-up required' }) }); return }
          Object.assign(user, body)
        }
        data = user
      } else if (path === '/api/v1/admin/groups') {
        if (holdGroups) { const wait = holdGroups; holdGroups = null; wait.seen = true; await wait.promise }
        if (failGroups) { failGroups = false; await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ code: 'FIXTURE_LOAD_FAILED', message: 'Synthetic load failure' }) }); return }
        data = { items: groups, total: groups.length, page: 1, page_size: 1000, pages: 1 }
      } else if (path === '/api/v1/admin/groups/all') data = groups
      else if (path.endsWith('/stats/batch')) data = { stats: {} }
      else if (path.endsWith('/system/version') || path.endsWith('/system/check-updates')) data = { current_version: '1.1.28', latest_version: '1.1.28', build_type: 'source', has_update: false }
      else if (path.includes('announcements')) data = { items: [], total: 0 }
      else if (path.includes('available') || path.endsWith('/models')) data = []
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, message: 'ok', data }) })
    })
    const page = await context.newPage(); page.setDefaultTimeout(8000)
    page.on('pageerror', e => errors.push({ viewport: viewport.name, message: e.message }))
    const record = name => { results.push({ viewport: viewport.name, name, status: 'passed' }); console.log(viewport.name + ': ' + name) }
    const edit = async index => { await page.getByRole('button', { name: 'Edit', exact: true }).nth(index).click(); await page.locator('#edit-user-form').waitFor() }
    const editDialog = () => page.getByRole('dialog', { name: 'Edit User', exact: true })
    const closeEdit = async () => { await editDialog().getByRole('button', { name: 'Cancel', exact: true }).click(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' }) }
    const groupsDialog = () => page.getByRole('dialog', { name: 'User Group Configuration', exact: true })
    const openGroups = async index => { const more = page.getByRole('button', { name: 'More', exact: true }).nth(index); await more.scrollIntoViewIfNeeded(); await page.waitForTimeout(100); await more.click(); const action = page.locator('.action-menu-content').getByRole('button', { name: 'Groups', exact: true }); if (!(await action.isVisible())) await more.click(); await action.click(); await groupsDialog().waitFor() }
    const closeGroups = async () => { await groupsDialog().getByRole('button', { name: 'Cancel', exact: true }).click(); await groupsDialog().waitFor({ state: 'hidden' }) }
    const awaitSeen = async wait => { for (let i = 0; i < 100 && !wait.seen; i++) await new Promise(res => setTimeout(res, 10)); assert(wait.seen, 'expected delayed route was not reached') }
    try {
      await page.goto(base + '/admin/users'); await page.getByText(users[0].email, { exact: true }).first().waitFor()
      const heading = page.getByRole('heading', { name: 'Two-Factor Verification Required', exact: true })
      const note = () => page.getByPlaceholder('Note fixture', { exact: true })
      const promote = async () => { await editDialog().locator('select').selectOption('admin'); await editDialog().getByRole('button', { name: 'Update', exact: true }).click() }
      const enterCode = async () => { for (let index = 0; index < 6; index++) await page.locator('input[maxlength="1"]').nth(index).fill(String(index + 1)) }
      const userPuts = id => requests.filter(r => r.viewport === viewport.name && r.method === 'PUT' && r.path === `/api/v1/admin/users/${id}`).length

      await edit(0); await note().fill('originating-promotion')
      const old = deferred(); holdUserSave = old
      await promote(); await awaitSeen(old); await page.keyboard.press('Escape')
      await page.locator('#edit-user-form').waitFor({ state: 'hidden' }); await edit(1); await note().fill('current-draft')
      const rejected = page.waitForResponse(r => new URL(r.url()).pathname === '/api/v1/admin/users/1' && r.request().method() === 'PUT')
      old.resolve(); await rejected; await page.waitForTimeout(100)
      assert(await heading.count() === 0, 'late old required response opened a prompt')
      assert(await note().inputValue() === 'current-draft', 'late old required response changed current draft')
      assert(userPuts(1) === 1, 'old promotion was retried')
      record('late old STEP_UP_REQUIRED cannot prompt or retry into a new user editor')

      await promote(); await heading.waitFor(); await page.keyboard.press('Escape')
      await heading.waitFor({ state: 'hidden' }); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
      await edit(1); await note().fill('normal-promotion'); const currentCount = userPuts(2)
      await promote(); await heading.waitFor(); await enterCode(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
      assert(userPuts(2) === currentCount + 2, 'current promotion did not retry exactly once')
      assert(users[1].role === 'admin' && values[2][2] === 'normal-promotion', 'current step-up did not save current user and values')
      assert(userPuts(1) === 1, 'current verification retried the closed original action')
      record('local Escape cancels its prompt; current verification retains one normal retry')

      verificationGrant = false
      await edit(0); await note().fill('old-verification'); await promote(); await heading.waitFor()
      const oldVerification = deferred(); holdVerification = oldVerification
      await enterCode(); await awaitSeen(oldVerification); await page.keyboard.press('Escape')
      await page.locator('#edit-user-form').waitFor({ state: 'hidden' }); await heading.waitFor({ state: 'hidden' })
      await edit(1); await note().fill('new-verification-draft'); await promote(); await heading.waitFor()
      const putsBeforeOldVerification = userPuts(2)
      await page.screenshot({ path: new URL(`${viewport.name}-new-stepup-prompt.png`, out).pathname, fullPage: true })
      const verifiedOld = page.waitForResponse(r => new URL(r.url()).pathname === '/api/v1/user/totp/step-up')
      oldVerification.resolve(); await verifiedOld; await page.waitForTimeout(100)
      assert(await heading.isVisible(), 'old verification closed a newer prompt')
      assert(userPuts(2) === putsBeforeOldVerification, 'old verification retried the newer operation')
      await enterCode(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
      assert(values[2][2] === 'new-verification-draft', 'current verification failed after ignoring old verification')
      record('unmounted old verification success cannot verify or retry a newer session')
      await page.screenshot({ path: new URL(`${viewport.name}-stepup-complete.png`, out).pathname, fullPage: true })
    } catch (error) {
      await page.screenshot({ path: new URL(`${viewport.name}-failure.png`, out).pathname, fullPage: true }).catch(() => {})
      await fs.writeFile(new URL(`${viewport.name}-failure.txt`, out), String(error) + '\n' + (await page.locator('body').innerText()).slice(-10000))
      throw error
    } finally { await context.close() }
  }
  assert(errors.length === 0, 'unexpected page errors: ' + JSON.stringify(errors))
} finally {
  await fs.writeFile(new URL('results.json', out), JSON.stringify({ environment: 'Google Chrome headless; final production preview real /admin/users component chain; all API responses and writes are intercepted synthetic fixtures, no business instance', results, errors }, null, 2) + '\n')
  await fs.writeFile(new URL('requests.json', out), JSON.stringify(requests, null, 2) + '\n')
  await browser.close()
}
