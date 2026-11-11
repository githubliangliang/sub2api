"""Run desired-behavior tests against the frozen product through Go overlays only."""
import json,subprocess,tempfile
from pathlib import Path
out=Path(__file__).resolve().parent
root=out.parents[2]
entries=json.loads((out/'probe-sources.json').read_text())
entries += [dict(label='service',package='internal/service'),dict(label='apicompat',package='internal/pkg/apicompat'),dict(label='antigravity',package='internal/pkg/antigravity')]
with tempfile.TemporaryDirectory(prefix='sub2api-021-probes-') as temp:
    overlay=Path(temp)/'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(root/'backend'/e['package']/('intake021_'+e['label']+'_test.go')):str(out/'probes'/(e['label']+'.go.txt')) for e in entries}}))
    for package in sorted(set(e['package'] for e in entries)):
        label=package.replace('/','-')+'-probe'
        subprocess.run(['rtk','proxy','python3',str(out/'run-check.py'),label,'backend','go','test','-tags=unit','-overlay='+str(overlay),'-run','^TestIntake021','-count=1','-v','./'+package],check=False)
