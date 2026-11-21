#!/usr/bin/env python3
"""Verify candidate coverage, docs, empty implementation evidence, and protected files."""
from pathlib import Path
import ast,collections,csv,hashlib,json,re,subprocess
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
def git(*args): return subprocess.check_output(['rtk','proxy','git',*args],cwd=ROOT,text=True)
def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()
facts=json.loads((OUT/'baseline.json').read_text())
protected=json.loads((OUT/'protected-initial-hashes.json').read_text())
changed=sorted(f for f,h in protected.items() if not (ROOT/f).is_file() or digest(ROOT/f)!=h)
allowed=['AGENTS.md','docs/upstream-sync/README.md']
assert changed==allowed,changed
assert git('rev-parse','HEAD').strip()==facts['fork_head']
assert git('diff','--cached','--name-only')==''
assert sorted(git('diff','--name-only').splitlines())==allowed
assert facts['fork_tree']==facts['snapshot_tree']
candidates=list(csv.DictReader((OUT/'candidates.tsv').open(),delimiter='\t'))
commits=list(csv.DictReader((OUT/'commits.tsv').open(),delimiter='\t'))
assert len(candidates)==41 and len(commits)==65
covered=set()
for c in candidates:
 members=set(git('rev-list','--no-merges',c['parent1']+'..'+c['sha']).splitlines())
 assert not (covered & members)
 assert len(members)==int(c['nonmerge_count'])
 covered |= members
 assert digest(OUT/'patches'/(c['candidate']+'.patch'))==c['patch_sha256']
assert covered=={c['sha'] for c in commits}
files=list(csv.DictReader((OUT/'files.tsv').open(),delimiter='\t'))
assert len(files)==217
assert collections.Counter(r['state'] for r in files)=={'CLEAN':135,'CONFLICT':62,'NOBASE':20}
classification=json.loads((OUT/'classification.json').read_text())
assert set(classification)=={c['candidate'] for c in candidates}
review=json.loads((OUT/'review.json').read_text())
assert {r['id'] for r in review if r['tier']==1}=={f'F{i:02}' for i in range(1,11)}
assert {r['id'] for r in review if r['tier']==2}=={f'S{i:02}' for i in range(1,16)}
assert len({c for r in review for c in r['candidates']})==31
assert len({c for r in review for c in r['candidates'] if c.startswith('PR')})==30
summary=json.loads((OUT/'probe-summary.json').read_text())
assert summary['local_failed_top_level']==17 and summary['local_passed_top_level']==10
for n in ['behaviour-probes','behaviour-probes-additional']:
 assert summary[n]['compiled'] and not summary[n]['panic']
assert summary['frontend_probes']['temporary_file_removed']
for n in ['baseline-protocol','baseline-consumers','baseline-frontend','baseline-test-list','source-scanner','source-scanner-benchmark','scheduler-projection-poc']:
 assert json.loads((OUT/(n+'.json')).read_text())['exit_code']==0,n
for script in OUT.glob('*.py'): ast.parse(script.read_text())
proposals=[];task_counts={};empty_rows=0
for tier,count in [(1,37),(2,52)]:
 folder=ROOT/'openspec/changes'/f'port-upstream-0.2.24-tier{tier}'
 for f in ['README.md','proposal.md','source-baseline.md','source-feature-map.md','design.md','tasks.md','verification.md']:
  assert (folder/f).is_file()
 tasks=(folder/'tasks.md').read_text();assert not re.search(r'\[[xX]\]',tasks)
 task_counts[str(tier)]=len(re.findall(r'^- \[ \]',tasks,re.M));assert task_counts[str(tier)]==count
 for line in (folder/'verification.md').read_text().splitlines():
  if not line.startswith('|') or line.startswith('|---'):continue
  cells=[x.strip() for x in line.split('|')[1:-1]]
  if cells[0] in ['项目','簇']:continue
  assert all(not c for c in cells[1:]),line
  empty_rows+=1
 source=(folder/'source-baseline.md').read_text()
 assert facts['fork_head'] in source
 assert all(facts['refs'][x] in source for x in ['v0.2.23','v0.2.24'])
 assert json.loads((OUT/f'openspec-tier{tier}.json').read_text())['exit_code']==0
 proposals+=list(folder.rglob('*.md'))
files_to_scan=proposals+[ROOT/'docs/upstream-sync/PORTING-0.2.24.md',OUT/'MANIFEST.md']
broken=[];links=0
for p in files_to_scan:
 for target in re.findall(r'\]\(([^)]+)\)',p.read_text()):
  if target.startswith(('http:','https:','#')):continue
  target=target.split('#')[0]
  if not target:continue
  links+=1
  if not (p.parent/target).exists() and (p.parent/target).resolve() != (OUT/'document-validation.json').resolve():broken.append([str(p.relative_to(ROOT)),target])
assert not broken,broken
prefixes=('docs/upstream-sync/PORTING-0.2.24.md','docs/upstream-sync/evidence-0.2.24/','openspec/changes/port-upstream-0.2.24-tier1/','openspec/changes/port-upstream-0.2.24-tier2/')
untracked=git('ls-files','--others','--exclude-standard').splitlines()
assert all(f.startswith(prefixes) for f in untracked),untracked
assert not (ROOT/'frontend/src/components/account/__tests__/intake024-bulk.spec.ts').exists()
assert git('diff','--check')==''
result={'fork_head':facts['fork_head'],'protected_initial_tracked_files':len(protected),'allowed_modified':changed,'unexpected_modified':[],'product_and_frozen_history_unchanged':True,'version_go_dependencies_sqlite_ent_wire_unchanged':True,'real_index_unchanged':True,'private_index_matches_fork_tree':True,'candidates':41,'nonmerge_commits':65,'coverage_unique_and_complete':True,'candidate_files':217,'states':dict(collections.Counter(r['state'] for r in files)),'patch_hashes_valid':True,'first_tier_clusters':10,'second_tier_clusters':15,'selected_prs':30,'selected_security_merges':1,'tasks_unchecked':task_counts,'verification_empty_rows':empty_rows,'openspec_strict':{'tier1':0,'tier2':0},'local_links_checked':links,'broken_links':[],'untracked_files':len(untracked),'frontend_temporary_removed':True,'go_probe_failed_top_level':17,'go_probe_controls_or_observations_passed':10,'implementation_performed':False}
(OUT/'document-validation.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False,indent=2))
