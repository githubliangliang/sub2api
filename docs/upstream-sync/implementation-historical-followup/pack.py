#!/usr/bin/env python3
"""Compress large raw outputs losslessly and record their uncompressed digests."""
from pathlib import Path
import gzip
import hashlib
import json

out = Path(__file__).resolve().parent
initial = out / 'initial-state.json'
if initial.exists():
    (out / 'initial-state.json.gz').write_bytes(gzip.compress(initial.read_bytes(), mtime=0))
    initial.unlink()
for record in sorted(out.glob('*.json')):
    data = json.loads(record.read_text())
    if not isinstance(data, dict) or not isinstance(data.get('command'), list) or 'exit_code' not in data:
        continue
    name = record.stem
    for stream in ('stdout', 'stderr'):
        original = out / (name + '.' + stream + '.log')
        zipped = out / (original.name + '.gz')
        if original.exists():
            raw = original.read_bytes()
            if len(raw) > 65536:
                zipped.write_bytes(gzip.compress(raw, mtime=0))
                original.unlink()
                chosen = zipped
            else:
                chosen = original
        elif zipped.exists():
            raw = gzip.decompress(zipped.read_bytes())
            chosen = zipped
        else:
            continue
        data[stream + '_log'] = chosen.name
        data[stream + '_sha256'] = hashlib.sha256(raw).hexdigest()
    record.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
print('Preserved raw outputs and their SHA256 digests; large logs compressed.')
