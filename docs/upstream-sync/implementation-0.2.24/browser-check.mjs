import {chromium} from '/tmp/sub2api-023-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';
import path from 'node:path';
import http from 'node:http';
const root=process.cwd(), out=path.join(root,'docs/upstream-sync/implementation-0.2.24/browser');
await fs.mkdir(out,{recursive:true});
const dist=path.join(root,'backend/internal/web/dist');
const server=http.createServer(async(req,res)=>{ try {let file=path.join(dist,new URL(req.url,'http://localhost').pathname);try{if(!(await fs.stat(file)).isFile())throw Error()}catch{file=path.join(dist,'index.html')}const ext=path.extname(file);res.setHeader('Content-Type',({'.html':'text/html','.js':'application/javascript','.css':'text/css','.svg':'image/svg+xml'})[ext]||'application/octet-stream');res.end(await fs.readFile(file))}catch{res.statusCode=500;res.end()} });
await new Promise(r=>server.listen(0,'127.0.0.1',r));const base=`http://127.0.0.1:${server.address().port}`;
const browser=await chromium.launch({executablePath:'/usr/bin/google-chrome',headless:true,args:['--no-sandbox']});
const results=[],errors=[],requests=[];const check=(v,msg)=>{if(!v)throw Error(msg)};
const user={id:1,email:'local-fixture@example.invalid',role:'admin',status:'active',balance:100,concurrency:1,run_mode:'standard'};
try {
for(const viewport of [{name:'desktop',width:1440,height:1000},{name:'mobile',width:390,height:844}]) {
 let snapshot=0,usageCalls=0,bulkBodies=[];
 let account={id:1,name:'Local 024 browser fixture',platform:'openai',type:'oauth',status:'active',schedulable:true,concurrency:1,priority:1,rate_multiplier:1,group_ids:[],groups:[],credentials:{},credentials_status:{has_access_token:true},extra:{openai_passthrough_model_filter:true},created_at:'2026-11-21T00:00:00Z',updated_at:'2026-11-21T00:00:00Z'};
 const context=await browser.newContext({viewport:{width:viewport.width,height:viewport.height},locale:'en-US'});
 await context.addInitScript(user=>{localStorage.setItem('auth_token','local-synthetic-fixture');localStorage.setItem('auth_user',JSON.stringify(user));localStorage.setItem('sub2api_locale','en');localStorage.setItem('admin_guide_1_admin_v4_interactive','true');localStorage.setItem('token_expires_at',String(Date.now()+3600000));localStorage.setItem('account-auto-refresh',JSON.stringify({enabled:true,interval_seconds:5}));},user);
 await context.route('**/api/v1/**',async route=>{
  const req=route.request(),url=new URL(req.url()),p=url.pathname;let body;try{body=req.postDataJSON()}catch{};requests.push({viewport:viewport.name,path:p,method:req.method(),query:url.search,body});let data={};
  if(p.endsWith('/auth/me')) data=user;
  else if(p.endsWith('/settings/public')) data={site_name:'Local 024 Acceptance',run_mode:'standard',registration_enabled:false};
  else if(p==='/api/v1/admin/accounts/bulk-update') {bulkBodies.push(body);account.extra={...account.extra,...body.extra};data={success:1,failed:0};}
  else if(p==='/api/v1/admin/accounts/usage/batch') {usageCalls++;data={usage:{1:{five_hour:{utilization:snapshot?44:21},seven_day:{utilization:snapshot?54:31},gemini_shared_daily:{utilization:snapshot?54:31}}},errors:{}};}
  else if(p==='/api/v1/admin/accounts/1/usage') {usageCalls++;data={five_hour:{utilization:snapshot?44:21},seven_day:{utilization:snapshot?54:31},gemini_shared_daily:{utilization:snapshot?54:31}};}
  else if(p==='/api/v1/admin/accounts/1') data=account;
  else if(p==='/api/v1/admin/accounts') data={items:[account],total:1,page:1,page_size:20,pages:1};
  else if(p.endsWith('/stats/batch')) data={stats:{}};
  else if(p.includes('/admin/groups')||p.includes('/admin/proxies')||p.includes('/models')) data=[];
  else if(p.includes('tls-fingerprint')||p.includes('announcements'))data={items:[],total:0};
  await route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({code:0,message:'ok',data})});
 });
 const page=await context.newPage();page.on('pageerror',e=>errors.push({viewport:viewport.name,message:e.message}));
 await page.goto(base+'/admin/accounts');await page.getByText(account.name,{exact:true}).waitFor();
 const openBulk=async()=>{await page.locator('input[type="checkbox"]:visible').first().check();await page.getByRole('button',{name:'Bulk Edit',exact:true}).click();await page.locator('#bulk-edit-account-form').waitFor();};
 await openBulk();await page.locator('#bulk-edit-openai-passthrough-enabled').check();await page.locator('#bulk-edit-openai-passthrough-toggle').click();await page.locator('button[type="submit"][form="bulk-edit-account-form"]').click();await page.locator('#bulk-edit-account-form').waitFor({state:'hidden'});
 check(bulkBodies.length===1&&!Object.hasOwn(bulkBodies[0].extra,'openai_passthrough_model_filter'),'passthrough-only must omit model filter');check(account.extra.openai_passthrough_model_filter===true,'persisted filter preserved');
 await openBulk();check(!(await page.locator('#bulk-edit-openai-passthrough-model-filter-enabled').isChecked()),'reopen resets opt-in');await page.locator('#bulk-edit-openai-passthrough-model-filter-enabled').check();await page.locator('[data-testid="bulk-openai-passthrough-model-filter"]').uncheck();await page.locator('button[type="submit"][form="bulk-edit-account-form"]').click();await page.locator('#bulk-edit-account-form').waitFor({state:'hidden'});check(bulkBodies[1].extra.openai_passthrough_model_filter===false,'explicit false submitted');check(!Object.hasOwn(bulkBodies[1].extra,'openai_passthrough'),'filter-only must omit passthrough');
 results.push(viewport.name+': bulk passthrough preserves filter; independent explicit false; reopen resets opt-in');
 await page.screenshot({path:path.join(out,'bulk-'+viewport.name+'.png')});
 for(const platform of ['anthropic','gemini']) {
  snapshot=0;account={...account,platform,type:'oauth',last_used_at:null,extra:{}};await page.goto(base+'/admin/accounts');await page.getByText(account.name,{exact:true}).waitFor();await page.getByText('31%',{exact:true}).first().waitFor();const before=usageCalls;
  snapshot=1;account={...account,last_used_at:'2026-11-21T09:30:00Z',extra:{passive_usage_7d_utilization:0.54,session_window_utilization:0.44}};
  await page.getByText('54%',{exact:true}).first().waitFor({timeout:15000});check(usageCalls>before,'snapshot refresh must issue another usage read');check(requests.filter(r=>r.path.endsWith('/usage/batch')).every(r=>!r.body?.force),'passive snapshot refresh must not force upstream');
  await page.screenshot({path:path.join(out,platform+'-'+viewport.name+'.png')});results.push(viewport.name+': '+platform+' auto refresh replaces same-id snapshot and reads passive usage');
 }
 await context.close();
}
check(errors.length===0,'browser pageerrors: '+JSON.stringify(errors));
}finally{await fs.writeFile(path.join(out,'results.json'),JSON.stringify({results,errors,requests,scope:'actual built application and local Chrome; synthetic intercepted API; no business instance or upstream network'},null,2));await browser.close();await new Promise(r=>server.close(r));}
console.log(JSON.stringify({results,errors,requests:requests.length},null,2));
