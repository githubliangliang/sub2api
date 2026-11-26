import { chromium } from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';

const out = new URL('./', import.meta.url);
const base = 'http://127.0.0.1:18926';
const pngBase64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+a2ioAAAAASUVORK5CYII=';
const initialLogo = `data:image/png;base64,${pngBase64}`;
const user = { id: 1, email: 'browser-followup@example.invalid', username: 'browser fixture', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' };
const templates = [1, 2].map(id => ({ id, name: `Browser Template ${id}`, provider: 'anthropic', api_mode: 'chat_completions', description: '', extra_headers: {}, body_override_mode: 'off', body_override: null, associated_monitors: 2 }));
const monitors = id => ({ items: [id * 10, id * 10 + 1].map(monitorId => ({ id: monitorId, name: `Browser Monitor ${monitorId}`, provider: 'anthropic', api_mode: 'chat_completions', enabled: true })) });
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const results = [], requests = [], pageErrors = [];
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] });

async function fulfill(route, data) {
  await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, message: 'ok', data }) });
}

async function flushBrowser(page) {
  await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
}

try {
  for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
    let holdFirstList = true, heldList = null, holdNextApply = true, heldApply = null;
    const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height }, locale: 'en-US' });
    await context.addInitScript(user => {
      localStorage.setItem('auth_token', 'synthetic-browser-only');
      localStorage.setItem('auth_user', JSON.stringify(user));
      localStorage.setItem('sub2api_locale', 'en');
      localStorage.setItem('admin_guide_1_admin_v4_interactive', 'true');
      localStorage.setItem('token_expires_at', String(Date.now() + 3600000));
      window.__nativeFileReader = window.FileReader;
      window.__browserReaders = [];
      window.FileReader = class {
        readyState = 1;
        result = null;
        onload = null;
        onerror = null;
        abortCount = 0;
        constructor() { window.__browserReaders.push(this); }
        readAsDataURL(file) { this.file = file; }
        readAsText(file) { this.file = file; }
        abort() { this.abortCount++; this.readyState = 2; }
        complete(value) { this.readyState = 2; this.result = value; this.onload?.({ target: this }); }
        fail() { this.readyState = 2; this.onerror?.({ target: this }); }
      };
    }, user);
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname;
      const body = request.postData() ? request.postDataJSON() : null;
      requests.push({ viewport: viewport.name, method: request.method(), path, body });
      let data = {};
      if (path.endsWith('/auth/me')) data = user;
      else if (path.endsWith('/settings/public')) data = { site_name: 'Historical Followup Browser', run_mode: 'standard', registration_enabled: false, channel_monitor_enabled: true };
      else if (path === '/api/v1/admin/settings') data = { site_name: 'Historical Followup Browser', site_logo: initialLogo, custom_menu_items: [{ id: 'browser-menu', label: 'Browser menu fixture', icon_svg: '', url: 'https://example.invalid', visibility: 'user', sort_order: 0 }], backend_mode_enabled: false, channel_monitor_enabled: true, smtp_port: 587 };
      else if (path === '/api/v1/admin/settings/web-search-emulation') data = { enabled: false, search_proxy_url: '' };
      else if (path.includes('/admin/groups')) data = [];
      else if (path === '/api/v1/admin/proxies') data = { items: [], total: 0 };
      else if (path === '/api/v1/admin/channel-monitor-templates') data = { items: templates };
      else if (/\/admin\/channel-monitor-templates\/\d+\/monitors$/.test(path)) {
        const id = Number(path.split('/').at(-2));
        if (id === 1 && holdFirstList) { holdFirstList = false; heldList = route; return; }
        data = monitors(id);
      } else if (/\/admin\/channel-monitor-templates\/\d+\/apply$/.test(path)) {
        if (holdNextApply) { holdNextApply = false; heldApply = route; return; }
        data = { affected: body.monitor_ids.length };
      } else if (path === '/api/v1/admin/channel-monitors') data = { items: [], total: 0, page: 1, page_size: 20, pages: 0 };
      else if (/announcements|providers|accounts|channels/.test(path)) data = { items: [], total: 0 };
      await fulfill(route, data);
    });
    const page = await context.newPage();
    page.setDefaultTimeout(10000);
    page.on('pageerror', error => pageErrors.push({ viewport: viewport.name, message: error.message }));

    await page.goto(`${base}/admin/settings`);
    const logo = page.getByText('Site Logo', { exact: true }).locator('..');
    const logoInput = logo.locator('input[type=file]');
    await logoInput.waitFor({ state: 'attached' });
    assert(await logo.locator('img').getAttribute('src') === initialLogo, 'initial Settings logo loaded');
    const choose = (name, buffer = Buffer.from('synthetic controlled image')) => logoInput.setInputFiles({ name, mimeType: 'image/png', buffer });
    await choose('first.png');
    await choose('second.png');
    await page.evaluate(() => {
      window.__browserReaders[1].complete('data:image/png;base64,bmV3');
      window.__browserReaders[0].complete('data:image/png;base64,b2xk');
      window.__browserReaders[0].fail();
    });
    await flushBrowser(page);
    assert(await logo.locator('img').getAttribute('src') === 'data:image/png;base64,bmV3', 'late image load cannot overwrite current Settings logo');
    assert(await page.evaluate(() => window.__browserReaders[0].abortCount) === 1, 'previous Settings reader aborted');
    assert(!(await logo.innerText()).includes('Failed to read'), 'old image error ignored');
    await choose('remove-pending.png');
    await logo.getByRole('button', { name: 'Remove', exact: true }).click();
    await page.evaluate(() => { window.__browserReaders[2].complete('data:image/png;base64,bGF0ZQ=='); window.__browserReaders[2].fail(); });
    await flushBrowser(page);
    assert(await logo.locator('img').count() === 0, 'late read cannot restore removed Settings logo');
    await choose('before-rejection.png');
    await choose('oversized.png', Buffer.alloc(300 * 1024 + 1));
    await page.evaluate(() => { window.__browserReaders[3].complete('data:image/png;base64,bGF0ZQ=='); window.__browserReaders[3].fail(); });
    await flushBrowser(page);
    assert(await logo.locator('img').count() === 0, 'rejected new file cancels pending Settings image');
    assert((await logo.innerText()).includes('max 300 KB'), '300KiB limit retained in Settings');
    assert(await page.evaluate(() => window.__browserReaders.length) === 4, 'oversized file starts no new read');
    await page.evaluate(() => { window.FileReader = window.__nativeFileReader; });
    await choose('native.png', Buffer.from(pngBase64, 'base64'));
    await page.waitForFunction(expected => [...document.querySelectorAll('img')].some(img => img.getAttribute('src') === expected), initialLogo);
    assert(!(await logo.innerText()).includes('File too large'), 'native successful read clears rejected-file error');
    const svgInput = page.locator('input[accept=".svg"]').first();
    const svg = '<svg xmlns="http://www.w3.org/2000/svg" onload="window.__unsafeSvgRan=true"><script>window.__unsafeSvgRan=true</script><path d="M0 0h1"/></svg>';
    await svgInput.setInputFiles({ name: 'browser-icon.svg', mimeType: 'image/svg+xml', buffer: Buffer.from(`  ${svg}\n`) });
    const svgComponent = svgInput.locator('xpath=ancestor::div[contains(@class,"items-start")][1]');
    await svgComponent.locator('span svg path').waitFor({ state: 'attached' });
    assert(await svgComponent.locator('script,[onload]').count() === 0, 'actual Settings SVG preview sanitized');
    assert(await page.evaluate(() => window.__unsafeSvgRan === true) === false, 'SVG injected code did not run');
    await svgComponent.scrollIntoViewIfNeeded();
    await page.screenshot({ path: new URL(`${viewport.name}-settings-svg.png`, out).pathname });
    results.push({ viewport: viewport.name, area: 'SettingsView', passed: ['controlled double-read ordering', 'reader abort', 'late error ignored', 'delete during read', 'rejected next file cancels prior read', '300KiB limit', 'native FileReader image', 'native SVG read and DOMPurify preview'] });

    await page.goto(`${base}/admin/channels/monitor`);
    await page.getByRole('button', { name: 'Templates', exact: true }).click();
    const manager = page.getByRole('dialog', { name: 'Request template manager', exact: true });
    await manager.getByText('Browser Template 1', { exact: true }).waitFor();
    const openPicker = async id => {
      const row = manager.getByText(`Browser Template ${id}`, { exact: true }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
      await row.getByRole('button', { name: 'Apply to monitors', exact: true }).click();
      return page.getByRole('dialog', { name: `Apply template "Browser Template ${id}"`, exact: true });
    };
    const firstListRequest = page.waitForRequest(request => request.url().endsWith('/channel-monitor-templates/1/monitors'));
    let picker = await openPicker(1);
    await firstListRequest;
    await flushBrowser(page);
    assert(heldList !== null, 'old association request held');
    await picker.getByRole('button', { name: 'Cancel', exact: true }).click();
    picker = await openPicker(2);
    await picker.getByText('Browser Monitor 20', { exact: true }).waitFor();
    const lateList = page.waitForResponse(response => response.url().endsWith('/channel-monitor-templates/1/monitors'));
    await fulfill(heldList, monitors(1));
    await lateList;
    await flushBrowser(page);
    assert((await picker.innerText()).includes('Browser Monitor 20'), 'current template association survives late old list');
    assert(!(await picker.innerText()).includes('Browser Monitor 10'), 'old template monitors stay absent');
    await picker.locator('li').filter({ hasText: 'Browser Monitor 21' }).locator('input[type=checkbox]').uncheck();
    await picker.getByRole('button', { name: 'Apply to 1 monitor(s)', exact: true }).click();
    assert(heldApply !== null, 'old apply request held');
    const oldApply = requests.filter(item => item.viewport === viewport.name && item.path.endsWith('/2/apply')).at(-1);
    assert(JSON.stringify(oldApply.body.monitor_ids) === '[20]', 'apply captures current template selected subset');
    await picker.getByRole('button', { name: 'Cancel', exact: true }).click();
    picker = await openPicker(1);
    await picker.getByText('Browser Monitor 10', { exact: true }).waitFor();
    const lateApply = page.waitForResponse(response => response.url().endsWith('/channel-monitor-templates/2/apply'));
    await fulfill(heldApply, { affected: 1 });
    await lateApply;
    await flushBrowser(page);
    assert(await picker.isVisible(), 'old apply completion cannot close reopened template picker');
    assert(await picker.locator('input[type=checkbox]:checked').count() === 2, 'old apply leaves current selection intact');
    await picker.locator('li').filter({ hasText: 'Browser Monitor 11' }).locator('input[type=checkbox]').uncheck();
    await page.screenshot({ path: new URL(`${viewport.name}-current-template.png`, out).pathname });
    const normalApply = page.waitForResponse(response => response.url().endsWith('/channel-monitor-templates/1/apply'));
    await picker.getByRole('button', { name: 'Apply to 1 monitor(s)', exact: true }).click();
    await normalApply;
    await picker.waitFor({ state: 'hidden' });
    const currentApply = requests.filter(item => item.viewport === viewport.name && item.path.endsWith('/1/apply')).at(-1);
    assert(JSON.stringify(currentApply.body.monitor_ids) === '[10]', 'normal current apply uses only current selected IDs');
    results.push({ viewport: viewport.name, area: 'ChannelMonitorView → manager → picker', passed: ['actual registered route', 'late association rows ignored after switch', 'selected IDs snapshot', 'old apply completion cannot close reopened picker', 'current selection preserved', 'normal current apply closes'], applySnapshots: [oldApply, currentApply] });
    await context.close();
  }
  assert(pageErrors.length === 0, JSON.stringify(pageErrors));
} catch (error) {
  await fs.writeFile(new URL('failure.json', out), JSON.stringify({ message: error.message, stack: error.stack, results, requests, pageErrors }, null, 2));
  throw error;
} finally {
  await fs.writeFile(new URL('result.json', out), JSON.stringify({ base, browser: 'Google Chrome via Playwright', api: 'Synthetic HTTP interception; no business instance, object store or real monitor apply', imageRace: 'Controlled FileReader callbacks in Chrome, followed by native image/SVG reads', results, requests, pageErrors }, null, 2));
  await browser.close();
}
console.log(JSON.stringify({ results, pageErrors }));
