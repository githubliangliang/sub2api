import { chromium } from '/tmp/sub2api-020-browser/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';
const out=new URL('./browser/',import.meta.url);await fs.mkdir(out,{recursive:true});
const browser=await chromium.launch({executablePath:'/usr/bin/google-chrome',headless:true,args:['--no-sandbox']});
const context=await browser.newContext({acceptDownloads:true,locale:'en-US'});
const user={id:1,email:'fixture@example.test',username:'fixture',role:'user',status:'active',balance:100,concurrency:1,run_mode:'standard'};
await context.addInitScript(user=>{localStorage.setItem('auth_token','synthetic-browser-only');localStorage.setItem('auth_user',JSON.stringify(user));localStorage.setItem('locale','en');localStorage.setItem('token_expires_at',String(Date.now()+3600000));},user);
const requests=[],errors=[];let mode='success',exportCalls=0;
const row={id:1,api_key_id:2,model:'gpt-6.1-terra',input_tokens:100,output_tokens:10,cache_read_tokens:0,cache_creation_tokens:0,rate_multiplier:1,actual_cost:0.001,total_cost:0.001,created_at:'2026-11-10T12:00:00Z',duration_ms:100,first_token_ms:10,api_key:{id:2,name:'fixture-key'},stream:false};
await context.route('**/api/v1/**',async route=>{
 const url=new URL(route.request().url());requests.push(url.pathname+url.search);let data={};
 if(url.pathname.endsWith('/auth/me'))data=user;
 else if(url.pathname.endsWith('/settings/public'))data={registration_enabled:false,site_name:'Tier2 Local Acceptance',run_mode:'standard',allow_user_view_error_requests:false};
 else if(url.pathname==='/api/v1/usage'){
  if(url.searchParams.get('page_size')==='1000'){
   exportCalls++;
   if(mode==='failure'||mode==='cancel'||(mode==='success'&&exportCalls===1)) {await route.fulfill({status:429,headers:{'Retry-After':mode==='cancel'?'17':'1'},contentType:'application/json',body:JSON.stringify({code:'RATE_LIMITED',message:'fixture rate limit'})});return;}
  }
  data={items:[row],total:1,page:1,page_size:20,pages:1};
 }else if(url.pathname.includes('snapshot'))data={stats:{total_requests:1,total_tokens:110,total_cost:0.001},trend:[],models:[],groups:[],endpoints:[]};
 else if(url.pathname.includes('models'))data={models:[]};
 else if(url.pathname.includes('stats'))data={total_requests:1,total_tokens:110,total_cost:0.001};
 else if(url.pathname.includes('available')||url.pathname.includes('/keys')||url.pathname.includes('/announcements'))data={items:[],total:0};
 await route.fulfill({contentType:'application/json',body:JSON.stringify({code:0,message:'ok',data})});
});
const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));
try {
 await page.goto('http://127.0.0.1:18762/usage');
 const button=page.getByRole('button',{name:/Export CSV|导出 CSV|导出CSV/i});await button.waitFor({timeout:15000});
 await page.screenshot({path:new URL('desktop.png',out).pathname,fullPage:true});
 const downloadPromise=page.waitForEvent('download');await button.click();const download=await downloadPromise;
 await download.saveAs(new URL('usage.csv',out).pathname);const csv=await fs.readFile(new URL('usage.csv',out),'utf8');
 if(exportCalls!==2||!csv.includes('fixture-key'))throw new Error(`success export calls=${exportCalls}, CSV invalid`);
 mode='failure';exportCalls=0;let unexpectedDownload=false;page.on('download',()=>{unexpectedDownload=true});await button.click();await page.waitForTimeout(3700);
 if(exportCalls!==4||unexpectedDownload)throw new Error(`failure export calls=${exportCalls} download=${unexpectedDownload}`);
 mode='cancel';exportCalls=0;await button.click();await page.waitForTimeout(200);await page.goto('http://127.0.0.1:18762/dashboard');await page.waitForTimeout(500);
 if(exportCalls!==1||unexpectedDownload)throw new Error('cancel produced another request or download');
 await page.setViewportSize({width:390,height:844});await page.goto('http://127.0.0.1:18762/usage');await button.waitFor();await page.screenshot({path:new URL('mobile.png',out).pathname,fullPage:true});
 await fs.writeFile(new URL('result.json',out),JSON.stringify({result:'PASS',successCalls:2,failureCalls:4,cancelCalls:exportCalls,errors,requests,scope:'actual built UI and browser; intercepted synthetic API; no business instance'},null,2));
 console.log('PASS: real browser export retry/download, retry exhaustion, navigation cancellation, mobile render');
} catch(e){await fs.writeFile(new URL('failure.json',out),JSON.stringify({error:String(e),errors,requests,url:page.url(),body:await page.locator('body').innerText()},null,2));await page.screenshot({path:new URL('failure.png',out).pathname,fullPage:true});throw e;}finally{await browser.close();}
