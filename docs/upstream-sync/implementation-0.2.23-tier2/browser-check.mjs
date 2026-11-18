import { chromium } from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';

const out = new URL('./browser/', import.meta.url);
await fs.mkdir(out, { recursive: true });
const assert = (value, message) => { if (!value) throw new Error(message); };
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] });
const results = [], requests = [], errors = [];
const base = 'http://127.0.0.1:18793';
const user = { id: 1, email: 'browser@example.invalid', username: 'fixture', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' };
const account = { id: 1, name: 'Tier2 quota account', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, concurrency: 1, priority: 1, rate_multiplier: 0.5, group_ids: [], groups: [],
  credentials: { access_token: 'synthetic-browser-only', refresh_token: 'synthetic-browser-only' }, extra: {}, created_at: '2026-11-18T00:00:00Z', updated_at: '2026-11-18T00:00:00Z' };
const endpoints = [{ endpoint: '/v1/messages', requests: 4, total_tokens: 1000, cost: 17, actual_cost: 4.1, account_cost: 2.9 }];
const model = { model: 'gpt-5.4', requests: 4, input_tokens: 700, output_tokens: 100, cache_creation_tokens: 100, cache_read_tokens: 100, total_tokens: 1000, cost: 17, actual_cost: 4.1, account_cost: 2.9 };
const stats = { summary: { days: 30, actual_days_used: 1, total_cost: 2.9, total_user_cost: 4.1, total_standard_cost: 17, total_requests: 4, total_tokens: 1000, avg_daily_cost: 2.9, avg_daily_user_cost: 4.1, avg_daily_requests: 4, avg_daily_tokens: 1000, avg_duration_ms: 100, today: {}, highest_cost_day: {}, highest_request_day: {} }, history: [], models: [model], endpoints, upstream_endpoints: endpoints, endpoint_paths: endpoints };
const oldUsage = { source: 'passive', error_code: 'quota_refresh_failed', error: 'old observation retained', five_hour: { utilization: 0, utilization_unknown: true, window_stats: { requests: 4, tokens: 1000, cost: 4.1 } }, seven_day: { utilization: 25, resets_at: '2026-11-25T00:00:00Z' } };
const newUsage = { source: 'active', updated_at: '2026-11-18T11:30:00Z', five_hour: { utilization: 12.5, resets_at: '2026-11-18T16:30:00Z' }, seven_day: { utilization: 25, resets_at: '2026-11-25T00:00:00Z' } };

try {
  for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
    let group = { id: 11, name: 'Tier2 Anthropic group', platform: 'anthropic', status: 'active', rate_multiplier: 1, subscription_type: 'standard', max_reasoning_effort: 'low', max_reasoning_effort_over_limit: 'downgrade', reasoning_effort_mappings: [], account_count: 0 };
    let activeQueries = 0, usage = structuredClone(oldUsage);
    const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height }, locale: 'en-US' });
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'synthetic-browser-only');
      localStorage.setItem('auth_user', JSON.stringify(user));
      localStorage.setItem('sub2api_locale', 'en');
      localStorage.setItem('admin_guide_1_admin_v4_interactive', 'true');
      localStorage.setItem('token_expires_at', String(Date.now() + 3600000));
    }, user);
    await context.route('**/api/v1/**', async route => {
      const req = route.request(), url = new URL(req.url()), path = url.pathname, body = req.postDataJSON();
      requests.push({ viewport: viewport.name, method: req.method(), path, query: url.search, body });
      let data = {};
      if (path.endsWith('/auth/me')) data = user;
      else if (path.endsWith('/settings/public')) data = { site_name: 'Tier2 Browser Acceptance', run_mode: 'standard', registration_enabled: false };
      else if (path === '/api/v1/admin/groups') data = { items: [group], total: 1, page: 1, page_size: 20, pages: 1 };
      else if (path === '/api/v1/admin/groups/11') { if (req.method() !== 'GET') group = { ...group, ...body }; data = group; }
      else if (path.includes('/admin/groups')) data = [group];
      else if (path === '/api/v1/admin/accounts') data = { items: [account], total: 1, page: 1, page_size: 20, pages: 1 };
      else if (path === '/api/v1/admin/accounts/1') data = account;
      else if (path === '/api/v1/admin/accounts/1/stats') data = stats;
      else if (path === '/api/v1/admin/accounts/1/usage') {
        if (url.searchParams.get('force') === 'true') {
          activeQueries++;
          if (activeQueries === 1) { await route.fulfill({ status: 502, contentType: 'application/json', body: JSON.stringify({ code: 502, message: 'fixture quota refresh failure' }) }); return; }
          usage = structuredClone(newUsage);
        }
        data = usage;
      } else if (path.endsWith('/admin/accounts/usage/batch')) data = { usage: { 1: usage }, errors: {} };
      else if (path.endsWith('/stats/batch')) data = { stats: {} };
      else if (path === '/api/v1/usage/stats') data = { total_requests: 4, total_tokens: 1000, total_cost: 17, total_actual_cost: 4.1, total_account_cost: 2.9, endpoints };
      else if (path.includes('/usage/models') || path.includes('/usage/dashboard/models')) data = [model];
      else if (path.includes('/usage/trend')) data = [];
      else if (path === '/api/v1/usage') data = { items: [], total: 0, page: 1, page_size: 20, pages: 1 };
      else if (path.includes('api-keys')) data = { items: [], total: 0 };
      else if (path.includes('available') || path.includes('/admin/proxies') || path.includes('/models')) data = [];
      else if (path.includes('announcements') || path.includes('upstream-billing') || path.includes('tls-fingerprint')) data = { items: [], total: 0 };
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, message: 'ok', data }) });
    });
    const page = await context.newPage();
    page.on('pageerror', error => errors.push({ viewport: viewport.name, message: error.message }));
    await page.goto(base + '/admin/groups');
    await page.getByText(group.name, { exact: true }).waitFor();
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    await page.locator('#edit-group-reasoning-max-effort').waitFor();
    await page.getByText(/Native Anthropic Messages use their own thinking settings/).waitFor();
    await page.locator('#edit-group-reasoning-max-effort').click();
    await page.getByRole('option', { name: 'medium', exact: true }).click();
    await page.locator('#edit-group-reasoning-over-limit').click();
    await page.getByRole('option', { name: 'Deny access', exact: true }).click();
    await page.getByRole('dialog').getByRole('button', { name: /^(Save|Update)/ }).last().click();
    await page.getByRole('dialog').waitFor({ state: 'hidden' });
    assert(group.max_reasoning_effort === 'medium' && group.max_reasoning_effort_over_limit === 'deny', 'Anthropic policy must persist');
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    await page.locator('#edit-group-reasoning-max-effort').waitFor();
    assert((await page.locator('#edit-group-reasoning-max-effort').textContent()).includes('medium'), 'Anthropic ceiling must reopen');
    await page.screenshot({ path: new URL('anthropic-effort-' + viewport.name + '.png', out).pathname });
    await page.getByRole('dialog').getByRole('button', { name: 'Cancel', exact: true }).last().click();
    results.push(viewport.name + ': Anthropic compatibility policy hint, save and reopen');

    await page.goto(base + '/admin/accounts');
    await page.getByText(account.name, { exact: true }).waitFor();
    await page.getByRole('button', { name: 'Query', exact: true }).first().waitFor();
    await page.getByText('Unknown (cached)', { exact: true }).waitFor();
    assert((await page.locator('body').innerText()).includes('25%'), 'failed observation must retain the known weekly value');
    await page.getByRole('button', { name: 'Query', exact: true }).first().click();
    await page.getByRole('alert').filter({ hasText: /Showing a previous observation/ }).waitFor();
    assert((await page.locator('body').innerText()).includes('25%'), 'manual refresh failure must retain readings');
    await page.getByRole('button', { name: 'Query', exact: true }).first().click();
    await page.locator('time[datetime="2026-11-18T11:30:00Z"]').waitFor();
    assert((await page.locator('body').innerText()).includes('13%'), 'a successful retry must replace the unknown window');
    await page.screenshot({ path: new URL('usage-freshness-' + viewport.name + '.png', out).pathname });
    results.push(viewport.name + ': unknown legacy quota, retained warning/readings, failed query and successful retry timestamp');

    await fs.writeFile(new URL('account-buttons-' + viewport.name + '.json', out), JSON.stringify(await page.getByRole('button').evaluateAll(nodes => nodes.map(node => ({ text: node.textContent, title: node.title, label: node.getAttribute('aria-label') }))), null, 2));
    await page.getByRole('button', { name: 'More', exact: true }).first().click();
    await page.getByRole('button', { name: 'View Stats', exact: true }).last().click();
    await page.getByRole('dialog').getByText('/v1/messages', { exact: true }).first().waitFor();
    const rows = page.getByRole('dialog').locator('tbody tr').filter({ hasText: '/v1/messages' });
    const row = await rows.first().innerText();
    assert(row.includes('$4.10') && row.includes('$2.90') && row.includes('$17.00'), 'admin endpoint columns must retain three distinct costs: ' + row);
    await page.screenshot({ path: new URL('admin-costs-' + viewport.name + '.png', out).pathname });
    results.push(viewport.name + ': admin account endpoint actual deduction, account cost and standard cost');
    await page.goto(base + '/usage');
    await page.getByText('/v1/messages', { exact: true }).waitFor();
    const userText = await page.locator('body').innerText();
    assert(!userText.includes('$2.90'), 'user usage must hide account cost even from a stale payload');
    assert(userText.includes('$4.10') && userText.includes('$17.00'), 'user actual and standard cost must remain distinct');
    await page.screenshot({ path: new URL('user-costs-' + viewport.name + '.png', out).pathname });
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), 'page must fit the viewport');
    results.push(viewport.name + ': user account-cost hiding and responsive layout');
    await context.close();
  }
  assert(errors.length === 0, 'browser errors: ' + JSON.stringify(errors));
} finally {
  await fs.writeFile(new URL('results.json', out), JSON.stringify({ results, requests, errors, fixture: 'built application and real Chrome with synthetic intercepted API; backend HTTP/SQLite acceptance is separate' }, null, 2));
  await browser.close();
}
console.log(JSON.stringify({ results, requests: requests.length, errors }, null, 2));
