#!/usr/bin/env python3
"""Check followup proposals and preservation of every original tracked input."""
from pathlib import Path
import ast,collections,csv,datetime,gzip,hashlib,json,re,subprocess
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
BASE=json.loads(gzip.decompress((OUT/'initial-hashes.json.gz').read_bytes()))
SLUGS=['port-upstream-historical-followup-tier1','port-upstream-historical-followup-tier2']

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(*args):return subprocess.check_output(['rtk','proxy','git',*args],cwd=ROOT,text=True).strip()

front=json.loads((OUT/'frontend-probes-final-behavior.stdout.log').read_text())
summary={k:front[k] for k in ['numTotalTests','numPassedTests','numFailedTests','numPendingTests','success']}
summary['assertions']=[{'title':a['fullName'],'status':a['status'],'errors':[s.splitlines()[0] for s in a['failureMessages']]} for r in front['testResults'] for a in r.get('assertionResults',[])]
(OUT/'frontend-probes-final-behavior-summary.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
assert summary['numTotalTests']==11 and summary['numFailedTests']==7 and summary['numPassedTests']==4
assert all(a['errors'][0].startswith('AssertionError:') for a in summary['assertions'] if a['status']=='failed')

intentional=[];unexpected=[];unchanged=0
for rel,h in BASE['tracked'].items():
    actual=sha(ROOT/rel) if (ROOT/rel).is_file() else None
    if actual==h:unchanged+=1
    elif rel in {'AGENTS.md','docs/upstream-sync/README.md'}:intentional.append({'path':rel,'before':h,'after':actual})
    else:unexpected.append(rel)
untracked_bad=[rel for rel,h in BASE['untracked'].items() if not (ROOT/rel).is_file() or sha(ROOT/rel)!=h]
new=[p for p in git('ls-files','--others','--exclude-standard','-z').split('\0') if p and p not in BASE['untracked']]
def allowed(p):return p=='docs/upstream-sync/UNMERGED-REVIEW-2026-11-26.md' or p.startswith('docs/upstream-sync/unmerged-review-2026-11-26/') or any(p.startswith('openspec/changes/'+s+'/') for s in SLUGS)
new_bad=[p for p in new if not allowed(p)]
report=ROOT/'docs/upstream-sync/UNMERGED-REVIEW-2026-11-26.md'
docs=[report,OUT/'README.md',OUT/'decisions.md']+[p for s in SLUGS for p in (ROOT/'openspec/changes'/s).rglob('*.md')]
links=[];broken=[];trailing=[]
for p in docs:
    text=p.read_text()
    for n,line in enumerate(text.splitlines(),1):
        if line.rstrip()!=line:trailing.append(f'{p.relative_to(ROOT)}:{n}')
    for target in re.findall(r'\]\(([^)]+)\)',text):
        if '://' in target or target.startswith('#'):continue
        dest=(p.parent/target.split('#',1)[0]).resolve()
        if dest.parent==OUT and dest.name in {'validation.json','artifacts.json'}:continue
        links.append({'source':str(p.relative_to(ROOT)),'target':target})
        if not dest.exists():broken.append(links[-1])
sources=json.loads((OUT/'sources.json').read_text());prior=json.loads((ROOT/'docs/upstream-sync/unmerged-review-2026-11-25/sources.json').read_text());prior={x['pr']:x for x in prior['sources']}
decisions=json.loads((OUT/'decisions.json').read_text())['decisions'];assert len(sources['sources'])==len(decisions)==24
assert {s['pr'] for s in sources['sources']}=={d['pr'] for d in decisions}
for s in sources['sources']:
    raw=gzip.decompress((OUT/f"patches/PR{s['pr']}.patch.gz").read_bytes());assert hashlib.sha256(raw).hexdigest()==s['sha256']==prior[s['pr']]['sha256']
files=list(csv.DictReader((OUT/'files.tsv').open(),delimiter='\t'));assert len(files)==61
assert dict(collections.Counter(r['state'] for r in files))==sources['state_count']
new_tasks={};empty_slots={}
for n,(slug,count) in enumerate(zip(SLUGS,[13,14]),1):
    folder=ROOT/'openspec/changes'/slug;text=(folder/'tasks.md').read_text()
    assert not re.findall(r'^\s*- \[[xX]\]',text,re.M)
    tasks=re.findall(r'^\s*- \[ \]',text,re.M);assert len(tasks)==count;new_tasks[slug]=len(tasks)
    slots=re.findall(r'^\| [^|]+ \| \|$',(folder/'verification.md').read_text(),re.M);assert len(slots)==6;empty_slots[slug]=len(slots)
    assert json.loads((OUT/f'openspec-tier{n}.json').read_text())['exit_code']==0
old=[]
for p in sorted((ROOT/'openspec/changes').glob('*/tasks.md')):
    if p.parent.name in SLUGS:continue
    text=p.read_text();old.append({'change':p.parent.name,'completed':len(re.findall(r'^\s*- \[[xX]\]',text,re.M)),'unchecked':len(re.findall(r'^\s*- \[ \]',text,re.M))})
assert len(old)==17 and sum(c['completed'] for c in old)==467 and sum(c['unchecked'] for c in old)==0
backend=json.loads((OUT/'backend-baseline.json').read_text());probe=json.loads((OUT/'backend-probes.json').read_text());wire=json.loads((OUT/'wire-controls-final.json').read_text())
assert backend['exit_code']==0 and backend['top_level']=={'pass':75} and backend['including_subtests']=={'pass':121}
assert probe['exit_code']==1 and probe['top_level']=={'fail':2,'pass':2} and probe['including_subtests']=={'fail':6,'pass':2}
assert wire['exit_code']==0 and wire['top_level']=={'pass':1} and wire['including_subtests']=={'pass':3}
assert json.loads((OUT/'frontend-baseline-summary.json').read_text())['numPassedTests']==15
assert json.loads((OUT/'frontend-probe-cleanup.json').read_text())['removed']
assert not (ROOT/'frontend/src/__tests__/backlog-next-probe.spec.ts').exists()
for path in json.loads((OUT/'overlay.json').read_text())['Replace']:assert not Path(path).exists()
for name in ['collect.py','run.py','run-frontend.py','validate.py']:ast.parse((OUT/name).read_text())
ws=subprocess.run(['rtk','proxy','git','diff','--check','--','AGENTS.md','docs/upstream-sync/README.md'],cwd=ROOT,capture_output=True,text=True)
refs=git('for-each-ref','--format=%(refname) %(objectname)','refs/heads','refs/tags','refs/remotes/upstream/main')
result={'checked_at_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'head':git('rev-parse','HEAD'),
        'head_unchanged':git('rev-parse','HEAD')==BASE['head'],'index_unchanged':git('write-tree')==BASE['index'],'refs_unchanged':refs==BASE['refs'],
        'initial_tracked':len(BASE['tracked']),'unchanged_tracked':unchanged,'intentional_changes':intentional,'unexpected_changes':unexpected,
        'initial_untracked_checked':len(BASE['untracked']),'initial_untracked_mismatches':untracked_bad,'new_files_before_validation_manifest':len(new),'unexpected_new_files':new_bad,
        'links_checked':len(links),'broken_links':broken,'trailing_markdown_whitespace':trailing,'git_diff_check_exit_code':ws.returncode,
        'source_prs':24,'source_candidate_files':61,'states':sources['state_count'],'old_completed_changes':old,'new_unchecked_tasks':new_tasks,'empty_implementation_slots':empty_slots,
        'frontend_goal_failures':7,'frontend_controls_or_observations_passed':4,'backend_goal_top_level_failures':2,'backend_top_level_controls_passed':2,
        'protected_scope':'All existing product/VERSION/dependencies/schema/Ent/Wire/resources/old OpenSpec/evidence files remain SHA256-identical; the only tracked changes are AGENTS and the upstream README. Original Fast diagnostic preserved.',
        'not_performed':'No new product implementation, full unit/integration/race/build/browser/security audit/provider/production-capacity acceptance, commit/push/release/archive or business-instance action.'}
assert result['head_unchanged'] and result['index_unchanged'] and result['refs_unchanged']
assert len(intentional)==2 and not unexpected and not untracked_bad and not new_bad and not broken and not trailing and ws.returncode==0
(OUT/'validation.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
artifacts={str(p.relative_to(OUT)):sha(p) for p in sorted(OUT.rglob('*')) if p.is_file() and p.name!='artifacts.json'}
(OUT/'artifacts.json').write_text(json.dumps({'algorithm':'SHA256','artifacts':artifacts},ensure_ascii=False,indent=2)+'\n')
assert all(sha(OUT/p)==h for p,h in artifacts.items())
print(json.dumps({k:result[k] for k in ['head_unchanged','index_unchanged','refs_unchanged','initial_tracked','unchanged_tracked','initial_untracked_checked','unexpected_changes','unexpected_new_files','broken_links','new_unchecked_tasks','empty_implementation_slots','source_prs','source_candidate_files']},ensure_ascii=False))
