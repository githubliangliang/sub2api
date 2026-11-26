import { chromium } from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs'
import fs from 'node:fs/promises'

const out = new URL('./', import.meta.url)
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
    let holdValues = null, holdGroups = null, holdUserSave = null, failGroups = false
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
      for (const number of ['0', '-3']) {
        await edit(0)
        await page.getByPlaceholder('Note fixture', { exact: true }).fill('normal text ' + number)
        const numberInput = page.getByPlaceholder('Count fixture', { exact: true }); await numberInput.fill(number)
        assert(await numberInput.getAttribute('min') === '-10' && await numberInput.getAttribute('max') === '100', 'integer bounds remain')
        await editDialog().getByRole('button', { name: 'Update', exact: true }).click(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
        assert(values[1][1] === number && values[1][2] === 'normal text ' + number, 'actual values request preserves strings/text')
      }
      record('actual /admin/users attribute JSON keeps zero/negative strings, text and bounds')

      await page.getByTitle('Attributes', { exact: true }).click()
      const config = page.getByRole('dialog', { name: 'User Attributes', exact: true })
      await config.getByTitle('Edit', { exact: true }).last().click()
      const definitionForm = page.locator('#attribute-form')
      await definitionForm.getByPlaceholder('Description text for the attribute', { exact: true }).fill('')
      await definitionForm.getByPlaceholder('Placeholder text for input field', { exact: true }).fill('')
      await page.getByRole('dialog', { name: 'Edit Attribute', exact: true }).getByRole('button', { name: 'Update', exact: true }).click()
      await definitionForm.waitFor({ state: 'hidden' })
      assert(definitions[1].description === '' && definitions[1].placeholder === '', 'definition clears reach actual JSON')
      await config.getByRole('button', { name: 'Close', exact: true }).click(); await config.waitFor({ state: 'hidden' })
      // Use the stable input positions now that the note placeholder has deliberately been cleared.
      const noteInput = () => page.locator('#edit-user-form').getByText('Note', { exact: true }).locator('..').locator('input')
      const countInput = () => page.getByPlaceholder('Count fixture', { exact: true })
      record('actual definition editor sends explicit empty description and placeholder')

      const oldValues = deferred(); holdValues = oldValues
      await edit(0); await awaitSeen(oldValues); await closeEdit(); await edit(1)
      await noteInput().fill('current-user-2')
      const oldValuesDone = page.waitForResponse(r => /\/admin\/users\/1\/attributes$/.test(new URL(r.url()).pathname) && r.request().method() === 'GET'); oldValues.resolve(); await oldValuesDone
      assert(await noteInput().inputValue() === 'current-user-2', 'late old values replaced current form')
      await editDialog().getByRole('button', { name: 'Update', exact: true }).click(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
      assert(values[2][2] === 'current-user-2', 'current user attribute save')
      record('close/open another user ignores delayed original values and saves current attributes')

      failGroups = true
      const putCount = requests.filter(r => r.viewport === viewport.name && r.method === 'PUT' && /\/admin\/users\/\d+$/.test(r.path)).length
      await openGroups(0); await groupsDialog().locator('[data-test=retry-groups]').waitFor()
      assert(await groupsDialog().getByRole('button', { name: 'Save', exact: true }).isDisabled(), 'failed load save must be disabled')
      assert(requests.filter(r => r.viewport === viewport.name && r.method === 'PUT' && /\/admin\/users\/\d+$/.test(r.path)).length === putCount, 'failure emitted a replacement PUT')
      await page.screenshot({ path: new URL(`${viewport.name}-failed-group-load.png`, out).pathname, fullPage: true })
      await groupsDialog().locator('[data-test=retry-groups]').click(); await groupsDialog().getByText('Exclusive fixture', { exact: true }).waitFor()
      await groupsDialog().locator('label:has(input[type=checkbox])').click()
      await groupsDialog().locator('input[type=number]').first().fill('')
      await groupsDialog().getByRole('button', { name: 'Save', exact: true }).click(); await groupsDialog().waitFor({ state: 'hidden' })
      assert(users[0].allowed_groups.length === 0 && users[0].group_rates[7] === null && users[0].group_rates[8] === 0.8, 'public/exclusive or rate-clear semantics changed')
      record('failed group load blocks PUT; retry preserves public/exclusive semantics and explicit rate removal')

      const oldGroups = deferred(); holdGroups = oldGroups
      await openGroups(0); await awaitSeen(oldGroups); await closeGroups(); await openGroups(1)
      await groupsDialog().getByText('Exclusive fixture', { exact: true }).waitFor()
      await groupsDialog().locator('label:has(input[type=checkbox])').click()
      await groupsDialog().locator('input[type=number]').first().fill('0.4')
      const oldGroupsDone = page.waitForResponse(r => /\/admin\/groups(?:\?|$)/.test(new URL(r.url()).pathname)); oldGroups.resolve(); await oldGroupsDone
      assert(await groupsDialog().locator('input[type=checkbox]').isChecked(), 'old group load replaced current selection')
      assert(await groupsDialog().locator('input[type=number]').first().inputValue() === '0.4', 'old group load replaced current rate')
      await groupsDialog().getByRole('button', { name: 'Save', exact: true }).click(); await groupsDialog().waitFor({ state: 'hidden' })
      assert(users[1].allowed_groups.join() === '7' && users[1].group_rates[7] === 0.4, 'current group save did not reach user 2')
      record('delayed closed-session group read cannot replace current selection or rate')

      await edit(0); await noteInput().fill('originating-save-snapshot')
      const oldSave = deferred(); holdUserSave = oldSave
      await editDialog().getByRole('button', { name: 'Update', exact: true }).click(); await awaitSeen(oldSave)
      await closeEdit(); await edit(1); await noteInput().fill('new-session-draft')
      const attrSaved = page.waitForResponse(r => /\/admin\/users\/1\/attributes$/.test(new URL(r.url()).pathname) && r.request().method() === 'PUT')
      oldSave.resolve(); await attrSaved
      assert(await editDialog().isVisible(), 'old save closed the new editor')
      assert(await noteInput().inputValue() === 'new-session-draft', 'old save changed the new editor')
      assert(values[1][2] === 'originating-save-snapshot', 'old save did not retain submitted snapshot')
      await page.screenshot({ path: new URL(`${viewport.name}-new-user-session.png`, out).pathname, fullPage: true })
      await editDialog().getByRole('button', { name: 'Update', exact: true }).click(); await page.locator('#edit-user-form').waitFor({ state: 'hidden' })
      assert(values[2][2] === 'new-session-draft', 'new session normal save failed')
      record('pending old save retains its user/snapshot and cannot close or overwrite a new editor')
    } catch (error) {
      await page.screenshot({ path: new URL(`${viewport.name}-failure.png`, out).pathname, fullPage: true }).catch(() => {})
      await fs.writeFile(new URL(`${viewport.name}-failure.txt`, out), String(error) + '\n' + (await page.locator('body').innerText()).slice(-10000))
      throw error
    } finally { await context.close() }
  }
  assert(errors.length === 0, 'unexpected page errors: ' + JSON.stringify(errors))
} finally {
  await fs.writeFile(new URL('results.json', out), JSON.stringify({ environment: 'Google Chrome headless; Vite real /admin/users component chain; all API responses and writes are intercepted synthetic fixtures, no business instance', results, errors }, null, 2) + '\n')
  await fs.writeFile(new URL('requests.json', out), JSON.stringify(requests, null, 2) + '\n')
  await browser.close()
}
