"""Temporary overlay feasibility check; product migrations are not modified."""
from pathlib import Path
import json, subprocess, tempfile
out=Path(__file__).resolve().parent
root=out.parents[2]
with tempfile.TemporaryDirectory(prefix='sub2api-021-index-') as temp:
    overlay=Path(temp)/'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(root/'backend/internal/repository/intake021_sqlite_index_test.go'):str(out/'probes/sqlite-index.go.txt')}}))
    result=subprocess.run(['rtk','proxy','python3',str(out/'run-check.py'),'sqlite-index-feasibility','backend','go','test','-tags=unit','-overlay='+str(overlay),'-run','TestIntake021_SQLite','-count=1','-v','./internal/repository'])
    raise SystemExit(result.returncode)
