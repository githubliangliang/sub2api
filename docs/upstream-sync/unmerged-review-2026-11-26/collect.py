#!/usr/bin/env python3
"""Freeze historical whole-PR sources and check a private index, without applying."""
from pathlib import Path
import collections
import csv
import datetime
import gzip
import hashlib
import json
import os
import re
import subprocess
import tempfile

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
PRS = [7308, 7491, 7313, 7224, 7246, 7354, 7394, 7057, 7085, 7215, 7320, 7442, 7060, 7419, 7365, 7492, 7500, 7363, 7366, 7373, 7375, 7494, 4867, 7636]

def git(*args, **kw):
    return subprocess.check_output(['rtk','proxy','git',*args], cwd=ROOT, **kw)

def dump(name, value):
    (OUT/name).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')

def tsv(name, rows, fields):
    with (OUT/name).open('w') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(rows)

def gz(name, data):
    with (OUT/name).open('wb') as f:
        with gzip.GzipFile(filename='',mode='wb',fileobj=f,mtime=0) as z:z.write(data)

merges={}
for first_parent in (True,False):
    args=['log','--format=%H\t%P\t%s']
    if first_parent:args.append('--first-parent')
    for line in git(*args,'upstream/main').decode().splitlines():
        sha,parents,subject=line.split('\t',2)
        m=re.search(r'Merge pull request #(\d+)\b',subject)
        if m:merges.setdefault(int(m[1]),(sha,parents.split()[0],subject))

sources=[];files=[];patches={};missing=[]
(OUT/'patches').mkdir(exist_ok=True)
for pr in PRS:
    if pr not in merges:missing.append(pr);continue
    sha,parent,subject=merges[pr]
    patch=git('diff','--binary',parent,sha)
    key=f'PR{pr}';patches[key]=patch
    gz(f'patches/{key}.patch.gz',patch)
    source={'candidate':key,'pr':pr,'merge':sha,'parent':parent,'subject':subject,
            'sha256':hashlib.sha256(patch).hexdigest(),
            'shortstat':git('diff','--shortstat',parent,sha).decode().strip(),
            'nonmerge_commits':int(git('rev-list','--count','--no-merges',f'{parent}..{sha}'))}
    sources.append(source)
    chunks=re.split(rb'(?=^diff --git )',patch,flags=re.M)
    for chunk in chunks:
        if not chunk:continue
        m=re.match(rb'diff --git a/(.+) b/(.+)\n',chunk)
        if not m:raise ValueError(chunk[:100])
        path=m[2].decode()
        files.append({'candidate':key,'path':path,'base_exists':(ROOT/path).exists(),
                      'patch':chunk,'new_file':b'new file mode ' in chunk.split(b'@@',1)[0]})

with tempfile.TemporaryDirectory(prefix='sub2api-backlog-index-') as d:
    env={**os.environ,'GIT_INDEX_FILE':str(Path(d)/'index')}
    git('read-tree','HEAD',env=env)
    for reverse in (True,False):
        prefix='reverse' if reverse else 'forward'
        for item in files:
            cmd=['rtk','proxy','git','apply','--cached','--check','--whitespace=nowarn']
            if reverse:cmd.append('--reverse')
            p=subprocess.run(cmd,input=item['patch'],env=env,cwd=ROOT,capture_output=True)
            item[prefix+'_exit']=p.returncode
            item[prefix+'_error']=p.stderr.decode(errors='replace').strip().replace('\n',' | ')

for item in files:
    item['state']=('ALREADY' if item['reverse_exit']==0 else 'CLEAN' if item['forward_exit']==0 else 'CONFLICT' if item['base_exists'] else 'NOBASE')
    del item['patch']
tsv('files.tsv',files,['candidate','path','base_exists','new_file','reverse_exit','forward_exit','reverse_error','forward_error','state'])
for source in sources:
    source['states']=dict(collections.Counter(x['state'] for x in files if x['candidate']==source['candidate']))

decl=re.compile(r'^\s*func\s+(?:\([^)]*\)\s*)?(\w+)\s*(?:\[[^]]*\])?\(',re.M)
declarations=collections.defaultdict(list)
for p in (ROOT/'backend/internal').rglob('*.go'):
    text=p.read_text();path=str(p.relative_to(ROOT))
    for m in decl.finditer(text):declarations[m[1]].append(f'{path}:{text.count(chr(10),0,m.start())+1}')
calls=[]
for key,patch in patches.items():
    path='';seen=set()
    added='\n'.join(line[1:] for line in patch.decode(errors='replace').splitlines() if line.startswith('+') and not line.startswith('+++'))
    added_declarations=set(decl.findall(added))
    for line in patch.decode(errors='replace').splitlines():
        if line.startswith('+++ b/'):path=line[6:]
        elif line.startswith('+') and not line.startswith('+++') and path.endswith('.go'):
            for symbol in re.findall(r'\b([A-Za-z_]\w*)\s*\(',line[1:]):
                if symbol in {'if','for','switch','func'} or (path,symbol) in seen:continue
                seen.add((path,symbol));calls.append({'candidate':key,'path':path,'symbol':symbol,'added_by_candidate':symbol in added_declarations,'local_declarations':' | '.join(declarations[symbol][:4]),'line':line[1:].strip()})
tsv('added-calls.tsv',calls,['candidate','path','symbol','added_by_candidate','local_declarations','line'])
tags=[]
for p in sorted((ROOT/'backend').rglob('*_integration_test.go')):
    text=p.read_text();tags.append({'path':str(p.relative_to(ROOT)),'first_line':text.splitlines()[0] if text else ''})
dump('integration-build-tags.json',tags)
dump('sources.json',{'checked_at_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
                    'head':git('rev-parse','HEAD').decode().strip(),
                    'upstream_main':git('rev-parse','upstream/main').decode().strip(),
                    'source_limit':'Previously assessed releases through v0.2.25; no new release intake.',
                    'sources':sources,'missing_prs':missing,
                    'state_count':dict(collections.Counter(x['state'] for x in files)),
                    'candidate_files':len(files),'added_call_entries':len(calls),
                    'caveat':'Whole-PR size/states include already selected subitems and generated files; they are not remaining implementation estimates or compilation proof.'})
print(json.dumps({'candidates':len(sources),'missing':missing,'files':len(files),
                  'states':dict(collections.Counter(x['state'] for x in files)),
                  'calls':len(calls)},ensure_ascii=False))
