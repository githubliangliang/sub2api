#!/usr/bin/env python3
"""Losslessly pack large logs from this implementation; record original hashes."""
from pathlib import Path
import gzip
import hashlib
import json

out = Path(__file__).resolve().parent
for record in sorted(out.glob('*.json')):
    data = json.loads(record.read_text())
    if not isinstance(data, dict) or 'command' not in data or 'exit_code' not in data:
        continue
    for stream in ('stdout', 'stderr'):
        log = out / (record.stem + '.' + stream + '.log')
        if not log.exists():
            continue
        raw = log.read_bytes()
        digest = hashlib.sha256(raw).hexdigest()
        if len(raw) > 65536:
            packed = log.with_suffix(log.suffix + '.gz')
            packed.write_bytes(gzip.compress(raw, mtime=0))
            assert gzip.decompress(packed.read_bytes()) == raw
            log.unlink()
            log = packed
        data[stream + '_log'] = log.name
        data[stream + '_sha256'] = digest
    record.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
