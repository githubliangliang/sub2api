import { chromium } from '/tmp/sub2api-020-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';

const out = new URL('./browser/', import.meta.url);
await fs.mkdir(out, { recursive: true });
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] });
const results = [], requests = [], errors = [];
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const deferred = () => { let resolve; const promise = new Promise(r => { resolve = r; }); return { promise, resolve }; };
const user = { id: 1, email: 'browser@example.invalid', username: 'fixture', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' };
const key = { id: 1, user_id: 1, key: 'sk-browser-fixture', name: 'tier2-visible-key', group_id: null, status: 'active', quota: 0, quota_used: 0, ip_whitelist: [], ip_blacklist: [], rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0, created_at: '2026-11-11T00:00:00Z', updated_at: '2026-11-11T00:00:00Z' };
const account = { id: 11, name: 'tier2-account', platform: 'anthropic', type: 'oauth', status: 'active', schedulable: true, concurrency: 1, priority: 1, rate_multiplier: 1, credentials: {}, extra: {}, group_ids: [], groups: [], created_at: '2026-11-11T00:00:00Z', updated_at: '2026-11-11T00:00:00Z' };
let keyUsageGate = null, accountListGate = null, quotaGate = null;
let failKeyUsage = false, failAccounts = false, failQuota = false, utilization = 21, suffix = '';

const context = await browser.newContext({ viewport: { width: 1440, height: 1000 }, locale: 'en-US' });
await context.addInitScript(user => {
  localStorage.setItem('auth_token', 'synthetic-browser-only');
  localStorage.setItem('auth_user', JSON.stringify(user));
  localStorage.setItem('locale', 'en');
  localStorage.setItem('admin_guide_1_admin_v4_interactive', 'true');
  localStorage.setItem('token_expires_at', String(Date.now() + 3600000));
}, user);
await context.route('**/api/v1/**', async route => {
  const url = new URL(route.request().url()), path = url.pathname;
  requests.push(path + url.search);
  let data = {}, status = 200;
  if (path.endsWith('/auth/me')) data = user;
  else if (path.endsWith('/settings/public')) data = { site_name: 'Tier2 Browser Acceptance', run_mode: 'standard', registration_enabled: false };
  else if (path === '/api/v1/keys') data = { items: [key], total: 1, page: 1, page_size: 20, pages: 1 };
  else if (path.endsWith('/usage/dashboard/api-keys-usage')) {
    if (keyUsageGate) await keyUsageGate.promise;
    if (failKeyUsage) status = 500;
    else data = { stats: { 1: { api_key_id: 1, today_actual_cost: 0.5, total_actual_cost: 12.25 } } };
  } else if (path === '/api/v1/admin/accounts') {
    if (accountListGate) await accountListGate.promise;
    if (failAccounts) status = 500;
    else data = { items: [{ ...account, name: account.name + suffix }], total: 1, page: 1, page_size: 20, pages: 1 };
  } else if (path.endsWith('/admin/accounts/usage/batch') || /\/admin\/accounts\/\d+\/usage$/.test(path)) {
    if (quotaGate) await quotaGate.promise;
    if (failQuota) status = 500;
    else {
      const usage = { updated_at: '2026-11-11T00:00:00Z', five_hour: { utilization, resets_at: '2026-11-12T00:00:00Z', remaining_seconds: 3600 }, seven_day: null, seven_day_sonnet: null };
      data = path.endsWith('/batch') ? { usage: { 11: usage }, errors: {} } : usage;
    }
  } else if (path.includes('/stats/batch')) data = { stats: {} };
  else if (path.includes('upstream-billing')) data = { items: [], total: 0 };
  else if (path.includes('available') || path === '/api/v1/admin/groups' || path === '/api/v1/admin/proxies') data = [];
  else if (path.includes('announcements') || path.includes('tls-fingerprint')) data = { items: [], total: 0 };
  else if (path.endsWith('/user/group-rates')) data = {};
  try {
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(status === 200 ? { code: 0, message: 'ok', data } : { code: 'FIXTURE_FAILURE', message: 'fixture failed' }) });
  } catch (error) {
    if (!/closed|cancel|intercept|handled/i.test(String(error))) throw error;
  }
});

const page = await context.newPage();
page.on('pageerror', e => errors.push(e.message));
try {
  keyUsageGate = deferred();
  await page.goto('http://127.0.0.1:18763/keys');
  await page.getByText(key.name, { exact: true }).waitFor();
  assert(await page.locator('[data-test="usage-pending"]').count() > 0, 'slow usage must not block the key list');
  await page.screenshot({ path: new URL('keys-slow-desktop.png', out).pathname, fullPage: false });
  keyUsageGate.resolve(); keyUsageGate = null;
  await page.getByText('$12.2500', { exact: true }).waitFor();
  results.push('desktop key list renders while usage is pending, then updates costs');

  failKeyUsage = true;
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await page.getByText(key.name, { exact: true }).waitFor();
  await page.waitForFunction(() => !document.querySelector('[data-test="usage-pending"]'));
  assert(await page.getByText(key.name, { exact: true }).count() > 0, 'usage failure must retain the list');
  results.push('desktop usage failure retains usable key list');

  await page.goto('http://127.0.0.1:18763/admin/accounts');
  await page.getByText(account.name, { exact: true }).waitFor();
  await page.getByText('21%', { exact: true }).first().waitFor();
  const row = page.locator('tr').filter({ hasText: account.name });
  await row.evaluate(el => {
    el.dataset.acceptanceIdentity = 'retained-row';
    window.__acceptanceRow = el;
    window.__rowChanges = [];
    window.__resizeHistory = [];
    window.addEventListener("resize", () => window.__resizeHistory.push({ width: window.innerWidth, desktop: matchMedia("(min-width: 768px)").matches }));
    new MutationObserver(records => {
      for (const record of records) if (Array.from(record.removedNodes).some(n => n === el || n.contains?.(el))) window.__rowChanges.push({ removed: el.outerHTML.slice(0, 300), target: record.target.outerHTML?.slice(0, 300) });
    }).observe(document.body, { childList: true, subtree: true });
  });
  accountListGate = deferred(); quotaGate = deferred();
  // The existing refresh action is an icon-only secondary button.
  await page.locator('button.btn-secondary').filter({ has: page.locator('svg') }).first().click();
  await page.locator('[aria-busy="true"][inert]').waitFor();
  assert(await page.locator('tr[data-acceptance-identity="retained-row"]').count() === 1, 'list refresh must keep the same DOM row');
  assert(await page.getByText('21%', { exact: true }).count() > 0, 'known quota must stay visible');
  await page.screenshot({ path: new URL('accounts-refresh-desktop.png', out).pathname, fullPage: false });
  suffix = '-updated'; accountListGate.resolve(); accountListGate = null;
  await page.getByText(account.name + suffix, { exact: true }).waitFor();
  const identity = await page.evaluate(() => ({ connected: window.__acceptanceRow?.isConnected, resizes: window.__resizeHistory, changes: window.__rowChanges, rows: Array.from(document.querySelectorAll('tr[data-row-id]')).map(n => ({ id: n.dataset.rowId, marker: n.dataset.acceptanceIdentity, text: n.textContent.slice(0, 90) })) }));
  await fs.writeFile(new URL('row-identity.json', out), JSON.stringify(identity, null, 2));
  assert(await page.locator('tr[data-acceptance-identity="retained-row"]').count() === 1, 'same account must preserve its DOM after refresh: ' + JSON.stringify(identity));
  utilization = 44; quotaGate.resolve(); quotaGate = null;
  await page.getByText('44%', { exact: true }).first().waitFor();
  results.push('desktop account refresh retains row and quota, marks table inert, then updates');

  failQuota = true;
  await page.locator('button.btn-secondary').filter({ has: page.locator('svg') }).first().click();
  await page.waitForTimeout(500);
  assert(await page.getByText('44%', { exact: true }).count() > 0, 'failed quota refresh must retain known quota');
  results.push('desktop failed quota refresh keeps known quota');
  failQuota = false;

  failAccounts = true;
  await page.locator('button.btn-secondary').filter({ has: page.locator('svg') }).first().click();
  await page.waitForTimeout(500);
  assert(await page.getByText(account.name + suffix, { exact: true }).count() > 0, 'failed account refresh must retain the row');
  results.push('desktop failed account refresh retains prior row');
  failAccounts = false;

  await page.setViewportSize({ width: 390, height: 844 });
  await page.reload();
  await page.getByText(account.name + suffix, { exact: true }).waitFor();
  await page.screenshot({ path: new URL('accounts-mobile.png', out).pathname, fullPage: false });
  failKeyUsage = false; keyUsageGate = deferred();
  await page.goto('http://127.0.0.1:18763/keys');
  await page.getByText(key.name, { exact: true }).waitFor();
  assert(await page.locator('[data-test="usage-pending"]').count() > 0, 'mobile key list must also render before usage');
  await page.screenshot({ path: new URL('keys-slow-mobile.png', out).pathname, fullPage: false });
  keyUsageGate.resolve(); keyUsageGate = null;
  await page.getByText('$12.2500', { exact: true }).waitFor();
  results.push('390px mobile accounts and independent key usage render successfully');
  assert(errors.length === 0, 'unexpected browser errors: ' + errors.join('; '));
} finally {
  keyUsageGate?.resolve(); accountListGate?.resolve(); quotaGate?.resolve();
  await fs.writeFile(new URL('results.json', out), JSON.stringify({ results, errors, requests, note: 'Real built UI and local Chrome; all API data are synthetic route fixtures.' }, null, 2));
  await browser.close();
}
console.log(JSON.stringify({ results, errors, requests: requests.length }));
