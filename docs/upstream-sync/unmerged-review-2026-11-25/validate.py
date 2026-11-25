#!/usr/bin/env python3
"""Validate source/proposal scope and preserve all input files and refs."""
from pathlib import Path
import ast,collections,csv,datetime,gzip,hashlib,json,re,subprocess
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
BASE=json.loads(gzip.decompress((OUT/'initial-hashes.json.gz').read_bytes()))
SLUGS=['port-upstream-historical-backlog-tier1','port-upstream-historical-backlog-tier2']

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(*args):return subprocess.check_output(['rtk','proxy','git',*args],cwd=ROOT,text=True).strip()

allowed={'AGENTS.md','docs/upstream-sync/README.md'}
intentional=[];changed=[];unchanged=0
for rel,h in BASE['tracked'].items():
    p=ROOT/rel;actual=sha(p) if p.is_file() else None
    if actual==h:unchanged+=1
    elif rel in allowed:intentional.append({'path':rel,'before':h,'after':actual})
    else:changed.append(rel)
untracked=[{'path':rel,'matches':(ROOT/rel).is_file() and sha(ROOT/rel)==h} for rel,h in BASE['initial_untracked'].items()]
new=[p for p in git('ls-files','--others','--exclude-standard','-z').split('\0') if p and p not in BASE['initial_untracked']]
def in_scope(p):return p=='docs/upstream-sync/UNMERGED-REVIEW-2026-11-25.md' or p.startswith('docs/upstream-sync/unmerged-review-2026-11-25/') or any(p.startswith('openspec/changes/'+s+'/') for s in SLUGS)
unexpected=[p for p in new if not in_scope(p)]
report=ROOT/'docs/upstream-sync/UNMERGED-REVIEW-2026-11-25.md'
md=[report,OUT/'README.md',OUT/'inventory.md']+[p for s in SLUGS for p in (ROOT/'openspec/changes'/s).rglob('*.md')]
links=[];broken=[];trailing=[]
for p in md:
    text=p.read_text()
    for n,line in enumerate(text.splitlines(),1):
        if line.rstrip()!=line:trailing.append(f'{p.relative_to(ROOT)}:{n}')
    for target in re.findall(r'\]\(([^)]+)\)',text):
        if '://' in target or target.startswith('#'):continue
        f=target.split('#',1)[0];dest=(p.parent/f).resolve()
        if dest.parent==OUT and dest.name in {'validation.json','artifacts.json'}:continue
        links.append({'source':str(p.relative_to(ROOT)),'target':target})
        if not dest.exists():broken.append(links[-1])
sources=json.loads((OUT/'sources.json').read_text());inventory=json.loads((OUT/'inventory.json').read_text())
assert len(sources['sources'])==len(inventory)==134
assert {s['pr'] for s in sources['sources']}=={i['pr'] for i in inventory}
for s in sources['sources']:
    raw=gzip.decompress((OUT/f"patches/PR{s['pr']}.patch.gz").read_bytes());assert hashlib.sha256(raw).hexdigest()==s['sha256']
rows=list(csv.DictReader((OUT/'files.tsv').open(),delimiter='\t'))
assert len(rows)==1843 and dict(collections.Counter(x['state'] for x in rows))==sources['state_count']
assert sources['missing_prs']==[]
tasks={};empty_rows={}
for slug,want in zip(SLUGS,[5,6]):
    d=ROOT/'openspec/changes'/slug;text=(d/'tasks.md').read_text()
    completed=re.findall(r'^\s*- \[[xX]\]',text,re.M);unchecked=re.findall(r'^\s*- \[ \]',text,re.M)
    assert not completed and len(unchecked)==want
    verification=(d/'verification.md').read_text();slots=re.findall(r'^\| [^|]+ \| \|$',verification,re.M)
    assert len(slots)==6;tasks[slug]=len(unchecked);empty_rows[slug]=len(slots)
    assert json.loads((OUT/f'openspec-tier{1 if want==5 else 2}.json').read_text())['exit_code']==0
old=[]
for p in sorted((ROOT/'openspec/changes').glob('*/tasks.md')):
    if p.parent.name in SLUGS:continue
    text=p.read_text();old.append({'change':p.parent.name,'completed':len(re.findall(r'^\s*- \[[xX]\]',text,re.M)),'unchecked':len(re.findall(r'^\s*- \[ \]',text,re.M))})
assert len(old)==15 and sum(x['completed'] for x in old)==456 and sum(x['unchecked'] for x in old)==0
baseline=json.loads((OUT/'baseline.json').read_text());probe=json.loads((OUT/'probes-final.json').read_text());front=json.loads((OUT/'frontend-summary.json').read_text())
assert baseline['exit_code']==0 and baseline['top_level']=={'pass':98} and baseline['including_subtests']=={'pass':214}
assert probe['exit_code']==1 and probe['top_level']=={'pass':5,'fail':6} and probe['including_subtests']=={'pass':9,'fail':14}
assert front['exit_code']==0 and front['passed']==66
assert json.loads((OUT/'performance.json').read_text())['exit_code']==0
for name in ['collect.py','run.py','validate.py']:ast.parse((OUT/name).read_text())
whitespace=subprocess.run(['rtk','proxy','git','diff','--check','--','AGENTS.md','docs/upstream-sync/README.md'],cwd=ROOT,capture_output=True,text=True)
refs=git('for-each-ref','--format=%(refname) %(objectname)','refs/tags','refs/heads','refs/remotes/upstream/main')
result={'checked_at_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
        'head_unchanged':git('rev-parse','HEAD')==BASE['head'],'index_unchanged':git('write-tree')==BASE['index'],'refs_unchanged':refs==BASE['refs'],
        'initial_tracked':len(BASE['tracked']),'unchanged_tracked':unchanged,'intentional_changes':intentional,'unexpected_changes':changed,
        'initial_untracked_checked':len(untracked),'initial_untracked_mismatches':[x['path'] for x in untracked if not x['matches']],
        'new_files_at_validation':len(new),'unexpected_new_files':unexpected,'links_checked':len(links),'broken_links':broken,'trailing_markdown_whitespace':trailing,
        'whitespace_exit_code':whitespace.returncode,'old_active_changes':old,'new_unchecked_tasks':tasks,'empty_implementation_slots':empty_rows,
        'source_prs':134,'candidate_files':1843,'source_states':sources['state_count'],'strict_validation_exit_codes':[0,0],
        'current_baseline':baseline['top_level'],'probe_top_level':probe['top_level'],'frontend_passed':66,
        'protected_scope':'All original backend/frontend/dependencies/VERSION/migrations/generated files, old OpenSpec and frozen evidence, preceding untracked recheck and Fast diagnostic remain hash-identical. Only AGENTS and the upstream README change intentionally.',
        'not_performed':'No product patch application, new full unit/integration/race/build/browser/provider acceptance, commit/push/release/archive or business-instance operation.'}
assert result['head_unchanged'] and result['index_unchanged'] and result['refs_unchanged']
assert not changed and not unexpected and not result['initial_untracked_mismatches'] and not broken and not trailing and whitespace.returncode==0
(OUT/'validation.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
artifacts={str(p.relative_to(OUT)):sha(p) for p in sorted(OUT.rglob('*')) if p.is_file() and p.name!='artifacts.json'}
(OUT/'artifacts.json').write_text(json.dumps({'algorithm':'SHA256','artifacts':artifacts},ensure_ascii=False,indent=2)+'\n')
assert all(sha(OUT/p)==h for p,h in artifacts.items())
print(json.dumps({k:result[k] for k in ['head_unchanged','index_unchanged','refs_unchanged','initial_tracked','unchanged_tracked','initial_untracked_checked','unexpected_changes','unexpected_new_files','broken_links','new_unchecked_tasks','empty_implementation_slots','source_prs','candidate_files']},ensure_ascii=False))
