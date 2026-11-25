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
PRS = [9052, 9008, 8794, 8801, 8803, 8804, 8900, 8902, 8976, 8962, 8816, 8833, 8644, 8779, 8678, 8606, 8634, 8639, 8568, 8610, 8590, 8367, 8389, 8319, 8476, 8504, 8462, 8557, 8543, 8521, 8522, 8380, 8495, 8518, 8549, 8323, 8348, 8287, 8395, 8251, 8269, 8078, 8093, 8114, 8166, 8167, 7957, 7684, 7726, 7646, 7681, 7679, 7752, 6960, 7049, 6575, 6890, 6974, 6769, 6747, 7091, 6691, 7153, 6989, 7057, 6968, 7011, 6945, 6988, 7129, 7085, 6929, 6917, 6847, 7001, 6845, 6913, 6915, 6602, 6662, 6688, 6772, 7526, 7595, 7540, 8661, 8746, 8827, 8812, 7237, 7261, 7173, 7215, 7224, 7246, 7313, 7294, 7247, 7387, 7401, 7412, 7369, 7361, 7419, 7363, 7366, 7373, 7375, 7365, 7442, 7354, 7289, 7394, 5443, 7320, 7169, 7501, 7500, 7494, 7492, 7308, 7060, 7160, 7491, 7513, 4867, 7288, 7090, 7478, 7532, 5616, 7636, 7524, 7610]

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
