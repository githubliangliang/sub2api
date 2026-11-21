#!/usr/bin/env python3
"""Record exact commands, full output, timing and Go test events for local implementation verification."""
import sys,subprocess,time,json,collections,os
from pathlib import Path
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
name,relative,*args=sys.argv[1:]
started=time.time()
env=dict(os.environ,TZ="UTC",CI="true")
r=subprocess.run(["rtk","proxy",*args],cwd=ROOT/relative,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
(OUT/(name+".log")).write_bytes(r.stdout)
events=[]
for line in r.stdout.decode(errors="replace").splitlines():
 try: item=json.loads(line)
 except ValueError: continue
 if isinstance(item,dict) and item.get("Action") in ("pass","fail","skip"):
  events.append({k:item[k] for k in ("Action","Package","Test","Elapsed") if k in item})
summary={"command":["rtk","proxy",*args],"cwd":str(ROOT/relative),"exit_code":r.returncode,"seconds":round(time.time()-started,3),"events":events,"test_counts":dict(collections.Counter(x["Action"] for x in events if "Test" in x)),"package_counts":dict(collections.Counter(x["Action"] for x in events if "Test" not in x))}
(OUT/(name+".json")).write_text(json.dumps(summary,ensure_ascii=False,indent=2)+"\n")
print(json.dumps({k:v for k,v in summary.items() if k not in ("events","command")},ensure_ascii=False))
if not events: print(r.stdout.decode(errors="replace")[-1800:])
