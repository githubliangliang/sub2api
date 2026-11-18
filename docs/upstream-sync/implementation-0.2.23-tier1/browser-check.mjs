import { chromium } from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';

const out = new URL('./browser/', import.meta.url);
await fs.mkdir(out, { recursive: true });
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] });
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const results = [], requests = [], errors = [];
const base = 'http://127.0.0.1:18783';
const user = { id: 1, email: 'browser@example.invalid', username: 'fixture', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' };
const groups = [{ id: 11, name: 'Anthropic group', platform: 'anthropic', status: 'active', rate_multiplier: 1 }, { id: 22, name: 'OpenAI group', platform: 'openai', status: 'active', rate_multiplier: 1 }, { id: 33, name: 'Anthropic spare group', platform: 'anthropic', status: 'active', rate_multiplier: 1 }];
const initialAccount = { id: 11, name: 'Tier1 whitelist account', platform: 'openai', type: 'apikey', status: 'active', schedulable: true, concurrency: 1, priority: 1, rate_multiplier: 1,
  credentials: { api_key: 'synthetic-browser-key', base_url: 'https://upstream.example.invalid', model_mapping: { 'gpt-6.1-sol': 'gpt-6.1-sol', 'public-alias': 'gpt-6.1-sol' } }, extra: {}, group_ids: [], groups: [], created_at: '2026-11-18T00:00:00Z', updated_at: '2026-11-18T00:00:00Z' };
const initialChannel = { id: 1, name: 'Tier1 Bedrock channel', description: '', status: 'active', billing_model_source: 'channel_mapped', restrict_models: false,
  features_config: { bedrock_cc_compat: true, future_feature: { enabled: true } }, group_ids: [11, 22], model_pricing: [], model_mapping: {}, apply_pricing_to_account_stats: false, account_stats_pricing_rules: [], created_at: '2026-11-18T00:00:00Z', updated_at: '2026-11-18T00:00:00Z' };

try {
  for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
    let account = structuredClone(initialAccount), channels = [structuredClone(initialChannel)];
    const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height }, locale: 'en-US' });
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'synthetic-browser-only');
      localStorage.setItem('auth_user', JSON.stringify(user));
      localStorage.setItem('sub2api_locale', 'en');
      localStorage.setItem('admin_guide_1_admin_v4_interactive', 'true');
      localStorage.setItem('token_expires_at', String(Date.now() + 3600000));
    }, user);
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname;
      const body = request.postDataJSON();
      requests.push({ viewport: viewport.name, method: request.method(), path, body });
      let data = {};
      if (path.endsWith('/auth/me')) data = user;
      else if (path.endsWith('/settings/public')) data = { site_name: 'Tier1 Browser Acceptance', run_mode: 'standard', registration_enabled: false };
      else if (path === '/api/v1/admin/accounts') data = { items: [account], total: 1, page: 1, page_size: 20, pages: 1 };
      else if (path === '/api/v1/admin/accounts/11') {
        if (request.method() !== 'GET') account = { ...account, ...body };
        data = account;
      } else if (path === '/api/v1/admin/channels') {
        if (request.method() === 'POST') { const channel = { ...structuredClone(initialChannel), ...body, id: channels.length + 1 }; channels.push(channel); data = channel; }
        else data = { items: channels, total: channels.length, page: 1, page_size: 20, pages: 1 };
      } else if (/\/admin\/channels\/\d+$/.test(path)) {
        const id = Number(path.split('/').at(-1)), index = channels.findIndex(channel => channel.id === id);
        if (request.method() !== 'GET') channels[index] = { ...channels[index], ...body };
        data = channels[index];
      } else if (path.includes('/admin/groups')) data = groups;
      else if (path.includes('/stats/batch')) data = { stats: {} };
      else if (path.endsWith('/admin/accounts/usage/batch')) data = { usage: {}, errors: {} };
      else if (path.includes('upstream-billing')) data = { items: [], total: 0 };
      else if (path.includes('/models')) data = { models: [] };
      else if (path.includes('announcements') || path.includes('tls-fingerprint')) data = { items: [], total: 0 };
      else if (path.includes('available') || path.endsWith('/admin/proxies')) data = [];
      else if (path.includes('web-search-emulation')) data = { enabled: false };
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, message: 'ok', data }) });
    });
    const page = await context.newPage();
    page.on('pageerror', error => errors.push({ viewport: viewport.name, message: error.message }));
    await page.goto(base + '/admin/accounts');
    await page.getByText(account.name, { exact: true }).waitFor();
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    const input = page.getByPlaceholder('Enter custom model name', { exact: true });
    await input.waitFor();
    const add = input.locator('..').getByRole('button', { name: 'Add', exact: true });
    await input.fill(' claude-* ');
    await add.click();
    await page.getByText('Wildcards are not supported in the model whitelist. Enter an exact model name.', { exact: true }).waitFor();
    assert(await input.inputValue() === ' claude-* ', 'wildcard draft must be retained');
    await input.fill(' gpt-*-mini ');
    await input.press('Enter');
    assert(await input.inputValue() === ' gpt-*-mini ', 'Enter must retain an invalid draft');
    await input.fill('public-alias');
    await add.click();
    await page.getByText(/A mapping already exists for public-alias/).waitFor();
    assert(await input.inputValue() === 'public-alias', 'mapping conflict must retain its draft');
    await input.fill(' exact-browser-model ');
    await input.press('Enter');
    assert(await input.inputValue() === '', 'corrected exact model must be accepted');
    await input.fill('exact-browser-model');
    await add.click();
    await page.getByText('Model already exists', { exact: true }).waitFor();
    await page.waitForTimeout(500); // Let the existing toast/dialog CSS transitions settle for the visual record.
    await page.screenshot({ path: new URL('whitelist-' + viewport.name + '.png', out).pathname, fullPage: false });
    const dialog = page.getByRole('dialog');
    await fs.writeFile(new URL('account-buttons-' + viewport.name + '.json', out), JSON.stringify(await dialog.getByRole('button').allTextContents(), null, 2));
    await dialog.getByRole('button', { name: /^(Save|Update)/ }).last().click();
    await input.waitFor({ state: 'hidden' });
    assert(account.credentials.model_mapping['exact-browser-model'] === 'exact-browser-model', 'saved whitelist must contain only the corrected exact model');
    assert(!Object.keys(account.credentials.model_mapping).some(name => name.includes('*')), 'invalid wildcard must never be saved');
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    await page.getByPlaceholder('Enter custom model name', { exact: true }).waitFor();
    assert(await page.getByText('exact-browser-model', { exact: true }).count() > 0, 'saved whitelist must reopen');
    results.push(viewport.name + ': wildcard click/Enter, draft, correction, duplicate/mapping conflict, save/reopen passed');

    await page.goto(base + '/admin/channels');
    await page.getByText(initialChannel.name, { exact: true }).waitFor();
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    const label = page.getByText('Bedrock CC Compatibility', { exact: true });
    await page.getByRole('dialog').getByRole('button', { name: 'Anthropic', exact: true }).click();
    await label.waitFor();
    const toggle = label.locator('..').locator('..').getByRole('switch');
    assert(await toggle.getAttribute('aria-checked') === 'true', 'saved boolean true must reopen enabled');
    await toggle.click();
    assert(await toggle.getAttribute('aria-checked') === 'false', 'the toggle must be off before saving');
    await page.waitForTimeout(500);
    await page.screenshot({ path: new URL('bedrock-' + viewport.name + '.png', out).pathname, fullPage: false });
    await page.getByRole('dialog').getByRole('button', { name: /^(Save|Update)/ }).last().click();
    await label.waitFor({ state: 'hidden' });
    assert(channels[0].features_config.bedrock_cc_compat === false, 'update must save boolean false');
    assert(channels[0].features_config.future_feature.enabled === true, 'other features must survive');
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    await page.getByRole('dialog').getByRole('button', { name: 'Anthropic', exact: true }).click();
    await label.waitFor();
    assert(await toggle.getAttribute('aria-checked') === 'false', 'false must reopen disabled');
    await toggle.click();
    await page.getByRole('dialog').getByRole('button', { name: /^(Save|Update)/ }).last().click();
    await label.waitFor({ state: 'hidden' });
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    const platformCheckbox = () => page.getByRole('dialog').locator('label').filter({ hasText: /^Anthropic$/ }).getByRole('checkbox');
    await platformCheckbox().setChecked(false);
    await page.getByRole('dialog').getByRole('button', { name: /^(Save|Update)/ }).last().click();
    await page.getByRole('dialog').waitFor({ state: 'hidden' });
    assert(channels[0].features_config.bedrock_cc_compat === false, 'removing Anthropic must clear a previous true');
    assert(channels[0].group_ids.length === 1 && channels[0].group_ids[0] === 22, 'other platform groups must remain');
    assert(channels[0].features_config.future_feature.enabled === true, 'removal must retain unrelated features');
    for (const enabled of [true, false]) {
      await page.getByRole('button', { name: 'Create Channel', exact: true }).click();
      await page.getByRole('dialog').locator('input[required]').fill('Created ' + viewport.name + ' ' + enabled);
      await platformCheckbox().setChecked(true);
      await page.getByRole('dialog').getByRole('button', { name: 'Anthropic', exact: true }).click();
      await page.getByRole('dialog').locator('label').filter({ hasText: enabled ? 'Anthropic group' : 'Anthropic spare group' }).getByRole('checkbox').setChecked(true);
      if (enabled) await toggle.click();
      await page.getByRole('dialog').getByRole('button', { name: /^(Create|Save)/ }).last().click();
      await page.getByRole('dialog').waitFor({ state: 'hidden' });
      assert(channels.at(-1).features_config.bedrock_cc_compat === enabled, 'creation must save a boolean ' + enabled);
      await page.getByRole('button', { name: 'Edit', exact: true }).last().click();
      await page.getByRole('dialog').getByRole('button', { name: 'Anthropic', exact: true }).click();
      assert(await toggle.getAttribute('aria-checked') === String(enabled), 'created channel must reopen correctly');
      await page.getByRole('dialog').getByRole('button', { name: 'Cancel', exact: true }).click();
    }
    results.push(viewport.name + ': channel creation true/false, updates, unrelated features, reopen and Anthropic removal passed');
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), 'page must fit the viewport');
    await context.close();
  }
  assert(errors.length === 0, 'browser page errors: ' + JSON.stringify(errors));
} finally {
  await fs.writeFile(new URL('results.json', out), JSON.stringify({ results, requests, errors }, null, 2));
  await browser.close();
}
console.log(JSON.stringify({ results, requestCount: requests.length, errors }));
