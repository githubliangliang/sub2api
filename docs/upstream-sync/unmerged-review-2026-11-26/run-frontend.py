#!/usr/bin/env python3
"""Run a temporary test against real Vue components; always remove the test file."""
from pathlib import Path
import json,subprocess
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
target=ROOT/'frontend/src/__tests__/backlog-next-probe.spec.ts'
assert not target.exists()
try:
    target.write_bytes((OUT/'probes/frontend.spec.ts.txt').read_bytes())
    result=subprocess.run(['rtk','proxy','python3',str(OUT/'run.py'),'frontend-probes-final-behavior','frontend','pnpm','exec','vitest','run','src/__tests__/backlog-next-probe.spec.ts','--reporter=json'],cwd=ROOT)
finally:
    target.unlink(missing_ok=True)
    (OUT/'frontend-probe-cleanup.json').write_text(json.dumps({'temporary_test':str(target.relative_to(ROOT)),'source':'probes/frontend.spec.ts.txt','removed':not target.exists(),'production_components_changed':False},indent=2)+'\n')
