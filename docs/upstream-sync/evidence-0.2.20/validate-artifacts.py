"""Validate intake inventory, proposed documents, preserved baseline and scope."""
import collections,csv,hashlib,json,re,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
root=out.parents[2]
base=json.loads((out/'baseline.json').read_text())
protected=json.loads((out/'protected-before.json').read_text())
errors=[]
changed=[name for name,digest in protected.items() if not (root/name).is_file() or hashlib.sha256((root/name).read_bytes()).hexdigest()!=digest]
if changed:errors.append({'protected_changes':changed})
candidates=list(csv.DictReader((out/'candidates.tsv').open(),delimiter='\t'))
files=list(csv.DictReader((out/'files.tsv').open(),delimiter='\t'))
assert len(candidates)==42 and sum(int(c['nonmerge_count']) for c in candidates)==72
assert collections.Counter(f['state'] for f in files)=={'ALREADY':3,'CLEAN':330,'CONFLICT':157,'NOBASE':72}
for c in candidates:
 if hashlib.sha256((out/'patches'/(c['candidate']+'.patch')).read_bytes()).hexdigest()!=c['patch_sha256']:errors.append({'patch_hash':c['candidate']})
proposals=[root/'openspec/changes'/('port-upstream-0.2.20-'+tier) for tier in ['tier1','tier2']]
markdown=[root/'docs/upstream-sync/PORTING-0.2.20.md',out/'MANIFEST.md']
counts={}
for proposal in proposals:
 for name in ['README.md','proposal.md','source-baseline.md','source-feature-map.md','design.md','tasks.md','verification.md']:
  if not (proposal/name).is_file():errors.append({'missing':str(proposal/name)})
 tasks=(proposal/'tasks.md').read_text()
 if re.search(r'\[[xX]\]',tasks):errors.append({'checked_tasks':str(proposal)})
 specs=list((proposal/'specs').rglob('spec.md'))
 requirements=sum(s.read_text().count('### Requirement:') for s in specs)
 scenarios=sum(s.read_text().count('#### Scenario:') for s in specs)
 counts[proposal.name]={'unchecked_tasks':tasks.count('[ ]'),'requirements':requirements,'scenarios':scenarios,'capabilities':len(specs)}
 markdown+=list(proposal.rglob('*.md'))
 if base['fork_head'] not in (proposal/'source-baseline.md').read_text():errors.append({'source_head':str(proposal)})
 # Implementation slots must remain empty, distinct from assessment evidence.
 for row in (proposal/'verification.md').read_text().splitlines():
  if re.match(r'\| [FS]\d\d ',row) and not row.endswith('| | | |'):errors.append({'nonempty_verification':row})
for p in markdown:
 for target in re.findall(r'\]\(([^)]+)\)',p.read_text()):
  target=target.split('#')[0]
  if not target or re.match(r'\w+://',target):continue
  path=(p.parent/target).resolve()
  if path==out/'document-validation.json':continue
  if not path.exists():errors.append({'broken_link':str(p.relative_to(root)),'target':target})
checks={}
for label in ['baseline-test-list','baseline-targeted','baseline-frontend','performance-neighbor-list','performance-neighbors','openspec-tier1','openspec-tier2']:
 result=json.loads((out/(label+'.json')).read_text()); checks[label]=result['exit_code']
 if result['exit_code']!=0:errors.append({'check_failed':label})
probe_counts={}
for label,want_fail,want_pass in [('apicompat',3,0),('service',10,1),('ws',1,0)]:
 s=(out/(label+'-probe.log')).read_text();failed=len(re.findall(r'^--- FAIL:',s,re.M));passed=len(re.findall(r'^--- PASS:',s,re.M)); probe_counts[label]={'expected_behavior_failures':failed,'control_passes':passed}
 if (failed,passed)!=(want_fail,want_pass) or '[build failed]' in s:errors.append({'probe_unexpected':label})
index_clean=subprocess.run(['git','diff','--cached','--quiet'],cwd=root).returncode==0
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
if not index_clean or head!=base['fork_head']:errors.append({'index_clean':index_clean,'head':head})
allowed=['AGENTS.md','docs/upstream-sync/README.md','docs/upstream-sync/PORTING-0.2.20.md','docs/upstream-sync/evidence-0.2.20/','openspec/changes/port-upstream-0.2.20-tier1/','openspec/changes/port-upstream-0.2.20-tier2/']
status=subprocess.check_output(['git','status','--porcelain=v1','--untracked-files=all'],cwd=root,text=True)
for row in status.splitlines():
 name=row[3:]
 if not any(name==a or (a.endswith('/') and name.startswith(a)) for a in allowed):errors.append({'unexpected_scope':name})
report={'result':'PASS' if not errors else 'FAIL','errors':errors,'protected_files':len(protected),'protected_changes':changed,'head':head,'real_index_clean':index_clean,'counts':counts,'checks':checks,'probes':probe_counts,'candidates':len(candidates),'candidate_file_states':dict(collections.Counter(f['state'] for f in files)),'validated_markdown_files':len(markdown),'product_changes':[] if not changed else changed,'implementation_status':'Not started'}
(out/'document-validation.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(report,ensure_ascii=False,indent=2))
raise SystemExit(bool(errors))
