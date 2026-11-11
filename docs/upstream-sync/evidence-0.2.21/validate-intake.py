"""Check intake-only scope, frozen hashes, source inventory and proposal state."""
from pathlib import Path
import ast,csv,hashlib,json,re,subprocess
out=Path(__file__).resolve().parent;root=out.parents[2]
def git(*args):return subprocess.check_output(['rtk','proxy','git',*args],cwd=root).decode()
def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
protected=json.loads((out/'protected-files.json').read_text())
allowed={'AGENTS.md','docs/upstream-sync/README.md'}
changed=[p for p,h in protected.items() if not (root/p).is_file() or digest(root/p)!=h]
assert set(changed)<=allowed,changed
assert git('diff','--cached','--name-only')=='','real index must remain clean'
assert git('rev-parse','HEAD').strip()==json.loads((out/'baseline.json').read_text())['fork_head']
assert set(git('diff','--name-only').splitlines())<=allowed
prefixes=('docs/upstream-sync/evidence-0.2.21/','docs/upstream-sync/PORTING-0.2.21.md','openspec/changes/port-upstream-0.2.21-tier1/','openspec/changes/port-upstream-0.2.21-tier2/')
untracked=git('ls-files','--others','--exclude-standard').splitlines()
assert all(p.startswith(prefixes) for p in untracked),untracked
candidates=list(csv.DictReader((out/'candidates.tsv').open(),delimiter='\t'))
assert len(candidates)==34
for c in candidates:assert digest(out/'patches'/(c['candidate']+'.patch'))==c['patch_sha256'],c['candidate']
classification=json.loads((out/'classification.json').read_text());assert set(classification)=={c['candidate'] for c in candidates}
clusters=json.loads((out/'review.json').read_text());assert len(clusters)==23
summary=json.loads((out/'probe-summary.json').read_text());assert summary['total_failed_top_level']==16
assert all(v['compiled'] for k,v in summary.items() if isinstance(v,dict))
for p in out.glob('*.py'):ast.parse(p.read_text())
files=[root/'docs/upstream-sync/PORTING-0.2.21.md',out/'MANIFEST.md']
task_counts={}
for tier in [1,2]:
 p=root/'openspec/changes'/f'port-upstream-0.2.21-tier{tier}'
 for name in ['README.md','proposal.md','source-baseline.md','source-feature-map.md','design.md','tasks.md','verification.md']:assert(p/name).is_file()
 tasks=(p/'tasks.md').read_text();assert not re.search(r'\[[xX]\]',tasks)
 task_counts[str(tier)]=len(re.findall(r'^- \[ \]',tasks,re.M))
 verification=(p/'verification.md').read_text()
 for line in verification.splitlines():
  if line.startswith('|') and not line.startswith('|---'):
   cells=[x.strip() for x in line.split('|')[1:-1]]
   if cells[0] not in ['项目','簇']:assert all(not c for c in cells[1:]),line
 source=(p/'source-baseline.md').read_text();assert json.loads((out/'baseline.json').read_text())['fork_head'] in source
 record=json.loads((out/f'openspec-tier{tier}.json').read_text());assert record['exit_code']==0
 files+=list(p.rglob('*.md'))
broken=[];checked=0
for p in files:
 for target in re.findall(r'\]\(([^)]+)\)',p.read_text()):
  if target.startswith(('http:','https:','#')):continue
  target=target.split('#')[0]
  if not target:continue
  checked+=1
  if not (p.parent/target).exists():broken.append([str(p.relative_to(root)),target])
assert not broken,broken
result=dict(fork_head=git('rev-parse','HEAD').strip(),protected_files=len(protected),allowed_modified_tracked=changed,unexpected_modified=[],real_index_unchanged=True,untracked_files=len(untracked),candidate_count=34,patch_hashes_valid=True,tasks_unchecked=task_counts,verification_evidence_empty=True,openspec_strict={'tier1':0,'tier2':0},local_links_checked=checked,broken_links=broken,product_changed=False,history_changed=False)
(out/'document-validation.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False,indent=2))
