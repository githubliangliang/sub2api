#!/usr/bin/env python3
"""Capture read-only checks; report the child exit code, never imply probe failures passed."""
from pathlib import Path
import collections,datetime,json,os,subprocess,sys,time
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
name,relative,*args=sys.argv[1:]
started=time.time()
p=subprocess.run(['rtk','proxy',*args],cwd=ROOT/relative,
                 env={**os.environ,'TZ':'UTC','CI':'true','GOMAXPROCS':'2'},
                 capture_output=True)
(OUT/(name+'.stdout.log')).write_bytes(p.stdout)
(OUT/(name+'.stderr.log')).write_bytes(p.stderr)
events=[]
for line in p.stdout.decode(errors='replace').splitlines():
    try:e=json.loads(line)
    except ValueError:continue
    if isinstance(e,dict) and e.get('Action') in {'pass','fail','skip'}:events.append(e)
summary={'command':['rtk','proxy',*args],'cwd':str(ROOT/relative),
         'environment':{'TZ':'UTC','CI':'true','GOMAXPROCS':'2'},
         'started_at_utc':datetime.datetime.fromtimestamp(started,datetime.timezone.utc).isoformat(),
         'exit_code':p.returncode,'seconds':round(time.time()-started,3),
         'top_level':dict(collections.Counter(e['Action'] for e in events if e.get('Test') and '/' not in e['Test'])),
         'including_subtests':dict(collections.Counter(e['Action'] for e in events if e.get('Test'))),
         'packages':dict(collections.Counter(e['Action'] for e in events if not e.get('Test')))}
(OUT/(name+'.json')).write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(summary,ensure_ascii=False))
if not events:print((p.stdout+p.stderr).decode(errors='replace')[-1800:])
