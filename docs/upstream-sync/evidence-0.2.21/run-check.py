"""Run an intake check, keeping exact argv, timestamps, exit status and output."""
import datetime,json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
label,folder,*command=sys.argv[1:]
env=dict(os.environ)
env['PATH']='/home/liang/.nvm/versions/node/v24.16.0/bin:'+env['PATH']
start=time.monotonic()
with (out/(label+'.log')).open('w') as log:
    result=subprocess.run(['rtk','proxy',*command],cwd=out.parents[2]/folder,env=env,stdout=log,stderr=subprocess.STDOUT)
record=dict(command=command,cwd=folder,finished_at_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),duration_seconds=round(time.monotonic()-start,3),exit_code=result.returncode,log=label+'.log')
(out/(label+'.json')).write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record))
sys.exit(result.returncode)
