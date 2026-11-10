"""Read-only Go overlays. Desired new behaviour fails on the frozen baseline."""
import json
from pathlib import Path
import subprocess
import tempfile

out = Path(__file__).resolve().parent
root = out.parents[2]
packages = [('apicompat','pkg/apicompat'),('service','service'),('ws','service/openai_ws_v2')]
with tempfile.TemporaryDirectory(prefix='sub2api-020-probes-') as temp:
    overlay = Path(temp)/'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(root/'backend/internal'/package/'release_intake_020_probe_test.go'):str(out/(label+'-probe.go.txt'))
        for label,package in packages
    }}))
    for label, package in packages:
        subprocess.run(['python3',str(out/'run-check.py'),label+'-probe','backend',
            'go','test','-tags=unit','-overlay='+str(overlay),'-run','^TestIntake020',
            '-count=1','-v','./internal/'+package],check=False)
