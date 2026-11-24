#!/usr/bin/env python3
import sys,subprocess,time,json,collections,os
from pathlib import Path
out=Path(__file__).resolve().parent; root=out.parents[3]
name,*args=sys.argv[1:]; started=time.time(); cmd=["rtk","proxy",*args]
r=subprocess.run(cmd,cwd=root/"backend",env=dict(os.environ,TZ="UTC",CI="true"),stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
(out/(name+".log")).write_bytes(r.stdout)
events=[]
for line in r.stdout.decode(errors="replace").splitlines():
 try:item=json.loads(line)
 except ValueError:continue
 if isinstance(item,dict) and item.get("Action") in ("pass","fail","skip"):
  events.append({k:item[k] for k in ("Action","Package","Test","Elapsed") if k in item})
summary={"command":cmd,"cwd":str(root/"backend"),"exit_code":r.returncode,"seconds":round(time.time()-started,3),"events":events,"tests":dict(collections.Counter(x["Action"] for x in events if "Test" in x)),"top_level":dict(collections.Counter(x["Action"] for x in events if "Test" in x and "/" not in x["Test"])),"packages":dict(collections.Counter(x["Action"] for x in events if "Test" not in x))}
(out/(name+".json")).write_text(json.dumps(summary,ensure_ascii=False,indent=2)+"\n")
print(json.dumps({k:v for k,v in summary.items() if k not in ("events",)},ensure_ascii=False))
if r.returncode:print(r.stdout.decode(errors="replace")[-5000:])
sys.exit(r.returncode)
