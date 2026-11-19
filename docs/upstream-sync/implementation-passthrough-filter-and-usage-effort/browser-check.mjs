const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright');
import fs from 'node:fs/promises';
import { createRequire } from 'node:module';
const XLSX = createRequire(process.cwd()+'/frontend/package.json')('xlsx');

const browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] });
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const results = [], requests = [], errors = [];
const base = process.env.APP_BASE_URL || 'http://127.0.0.1:18927';
const user = { id: 1, email: 'browser@example.invalid', username: 'fixture', role: 'admin', status: 'active', balance: 100, concurrency: 1, run_mode: 'standard' };
const groups = [{ id: 11, name: 'Anthropic group', platform: 'anthropic', status: 'active', rate_multiplier: 1 }, { id: 22, name: 'OpenAI group', platform: 'openai', status: 'active', rate_multiplier: 1 }, { id: 33, name: 'Anthropic spare group', platform: 'anthropic', status: 'active', rate_multiplier: 1 }];
const initialAccount = { id: 11, name: 'Filtered OpenAI account', platform: 'openai', type: 'apikey', status: 'active', schedulable: true, concurrency: 1, priority: 1, rate_multiplier: 1,
  credentials_status: {has_api_key:true}, credentials: {base_url:'https://upstream.example.invalid',model_mapping:{'gpt-6.1-sol':'gpt-6.1-sol'}}, extra:{}, group_ids:[], groups:[], created_at:'2026-11-18T00:00:00Z',updated_at:'2026-11-18T00:00:00Z' };

try {
  for (const viewport of [{ name: 'desktop', width: 1440, height: 1000 }, { name: 'mobile', width: 390, height: 844 }]) {
    let account = structuredClone(initialAccount);
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
      else if (path.endsWith('/settings/public')) data = { site_name: 'Browser Acceptance', run_mode: 'standard', registration_enabled: false };
      else if (path === '/api/v1/admin/accounts' && request.method()==='POST') data = {...account,...body,id:12};
      else if (path === '/api/v1/admin/accounts/bulk-update') data = {success:1,failed:0,success_ids:[11],results:[{account_id:11,success:true}]};
      else if (path === '/api/v1/admin/accounts') data = { items: [account], total: 1, page: 1, page_size: 20, pages: 1 };
      else if (path === '/api/v1/admin/accounts/11') {
        if (request.method() !== 'GET') account = { ...account, ...body, load_factor: body.load_factor > 0 ? body.load_factor : null };
        data = account;
      } else if (path === '/api/v1/usage' || path === '/api/v1/admin/usage') data = {items:[{id:1,created_at:'2026-11-19T00:00:00Z',model:'gpt-5.4',reasoning_effort:'xhigh',...(path.includes('/admin/')?{upstream_reasoning_effort:'high'}:{}),input_tokens:10,output_tokens:5,total_cost:0.1,actual_cost:0.1,rate_multiplier:1,request_type:1}],total:1,pages:1,page:1,page_size:20};
      else if (path.endsWith('/stats')) data = {total_requests:1,total_input_tokens:10,total_output_tokens:5, total_actual_cost:0.1,endpoints:[],upstream_endpoints:[],endpoint_paths:[]};
      else if (path.includes('/admin/groups')) data = groups;
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
    page.setDefaultTimeout(6000);
    page.on('pageerror', error => errors.push({ viewport: viewport.name, message: error.message }));
    await page.goto(base + '/admin/accounts');
    await page.getByText(account.name, { exact: true }).waitFor();
    await page.getByRole('button', { name: 'Edit', exact: true }).first().click();
    const dialog = page.getByRole('dialog');
    await dialog.getByTestId('edit-openai-passthrough-toggle').click();
    const check = dialog.getByTestId('edit-openai-passthrough-model-filter');
    await check.check();
    await dialog.getByRole('button',{name:/^(Save|Update)/}).last().click();
    await dialog.waitFor({state:'hidden'});
    assert(account.extra.openai_passthrough_model_filter === true, 'filter persisted');
    await page.getByRole('button',{name:'Edit',exact:true}).first().click();
    assert(await check.isChecked(),'reopen enabled');
    await check.uncheck();
    await dialog.getByRole('button',{name:/^(Save|Update)/}).last().click();
    await dialog.waitFor({state:'hidden'});
    assert(!account.extra.openai_passthrough_model_filter, 'filter disabled');
    await page.getByRole('button',{name:'Create Account',exact:true}).click();
    const createForm=page.locator('#create-account-form');
    await createForm.getByRole('button',{name:'OpenAI',exact:true}).click();
    await createForm.getByRole('button',{name:/^API Key/}).first().click();
    await createForm.locator('input[type=text]').first().fill('new filtered');
    await createForm.locator('input[type=password]').first().fill('synthetic-only');
    await createForm.getByTestId('create-openai-passthrough-toggle').click();
    await createForm.getByTestId('create-openai-passthrough-model-filter').check();
    await page.getByRole('button',{name:'Create',exact:true}).click();
    await createForm.waitFor({state:'hidden'});
    assert(requests.some(x=>x.path==='/api/v1/admin/accounts'&&x.method==='POST'&&x.body.extra.openai_passthrough_model_filter===true),'create filter');
    await page.locator('input[type=checkbox]:visible').first().check();
    await page.getByRole('button',{name:'Bulk Edit',exact:true}).click();
    await page.locator('#bulk-edit-openai-passthrough-enabled').check();
    await page.locator('#bulk-edit-openai-passthrough-toggle').click();
    await page.getByTestId('bulk-openai-passthrough-model-filter').check();
    await page.getByRole('dialog').getByRole('button',{name:'Update Accounts',exact:true}).click();
    await page.getByRole('dialog').waitFor({state:'hidden'});
    assert(requests.some(x=>x.path==='/api/v1/admin/accounts/bulk-update'&&x.body.extra.openai_passthrough_model_filter===true),'bulk filter');
    for (const path of ['/admin/usage','/usage']) {
      await page.goto(base + path);
      await page.getByTestId('usage-column-settings').click();
      const toggle = page.getByTestId('usage-column-toggle-reasoning_effort');
      if (!(await page.getByTestId('reasoning-effort-cell').count())) await toggle.click();
      await page.getByTestId('usage-column-settings').click();
      const cell = page.getByTestId('reasoning-effort-cell').first();
      await cell.waitFor();
      const content=await cell.innerText();
      assert(/Xhigh/i.test(content), 'requested '+content);
      if (path.includes('admin')) assert(/\bhigh\b/i.test(content),'forwarded '+content); else assert(content.trim() === 'XHigh', 'user saw a second value: '+content);
      const downloadEvent=page.waitForEvent('download',{timeout:6000});
      await page.getByRole('button',{name:path.includes('admin')?'Export Excel':'Export CSV',exact:true}).click();
      const exported=await downloadEvent;
      const book=XLSX.read(await fs.readFile(await exported.path()),{type:'buffer'});
      const [head,first]=XLSX.utils.sheet_to_json(book.Sheets[book.SheetNames[0]],{header:1});
      const requested=head.findIndex(v=>/requested reasoning effort/i.test(v));
      const effort=head.findIndex(v=>v==='Reasoning Effort');
      if(path.includes('admin')) assert(first[requested]==='XHigh'&&first[effort]==='High','admin export '+JSON.stringify([head,first]));
      else assert(requested===-1&&first[effort]==='XHigh','user export '+JSON.stringify([head,first]));
    }
    results.push(viewport.name+' account create/edit/reopen/disable/bulk, role-specific usage and downloaded XLSX/CSV passed');
    await context.close();
  }
  console.log(JSON.stringify({results,errors}));
  assert(errors.length === 0,JSON.stringify(errors));
} finally { await browser.close(); }
