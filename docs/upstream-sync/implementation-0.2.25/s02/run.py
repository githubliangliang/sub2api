#!/usr/bin/env python3
import os,sys,subprocess,time,json,collections
from pathlib import Path
out=Path(__file__).resolve().parent
name,cwd,*cmd=sys.argv[1:]
env=dict(os.environ,CI='true',TZ='UTC',GOTOOLCHAIN='go1.27.4')
started=time.time(); p=subprocess.run(['rtk','proxy',*cmd],cwd=cwd,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
(out/(name+'.log')).write_bytes(p.stdout)
events=[]
for line in p.stdout.decode(errors='replace').splitlines():
 try: e=json.loads(line)
 except ValueError: continue
 if isinstance(e,dict) and e.get('Action') in ('pass','fail','skip'):events.append({k:e[k] for k in ('Action','Package','Test','Elapsed') if k in e})
summary={'command':['rtk','proxy',*cmd],'cwd':cwd,'exit_code':p.returncode,'seconds':round(time.time()-started,3),'events':events,'test_counts':dict(collections.Counter(e['Action'] for e in events if 'Test' in e)),'top_level_counts':dict(collections.Counter(e['Action'] for e in events if 'Test' in e and '/' not in e['Test'])),'package_counts':dict(collections.Counter(e['Action'] for e in events if 'Test' not in e))}
(out/(name+'.json')).write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ('command','events')},indent=2))
if not events:print(p.stdout.decode(errors='replace')[-2500:])
