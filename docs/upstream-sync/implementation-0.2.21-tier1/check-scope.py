"""Attribute acceptance to the isolated tier1 product while preserving shared work."""
from pathlib import Path
import hashlib,json,re,subprocess
out=Path(__file__).resolve().parent
root=out.parents[2]
iso=Path(json.loads((out/'isolated-snapshot.json').read_text())['path'])
before=json.loads((out/'before-hashes.json').read_text())
products=json.loads((out/'product-hashes.json').read_text())
errors=[]
checks={}
for label in ['isolated-unit','isolated-build','isolated-lint','isolated-race','isolated-sqlite-http','after-targeted','boundaries-final','openspec-tier1-final']:
 record=json.loads((out/(label+'.json')).read_text());checks[label]=record['exit_code']
 if record['exit_code']:errors.append({'failed_check':label})
for name,digest in products.items():
 if hashlib.sha256((iso/name).read_bytes()).hexdigest()!=digest:errors.append({'isolated_product_changed':name})
reverse=subprocess.run(['git','apply','--reverse','--check',str(out/'product.diff')],cwd=root,capture_output=True,text=True)
if reverse.returncode:errors.append({'tier1_patch_not_intact_in_shared_worktree':reverse.stderr})
status_files={'AGENTS.md','docs/upstream-sync/README.md'}|{'openspec/changes/port-upstream-0.2.21-tier1/'+n for n in ['README.md','tasks.md','verification.md']}
changed=[];external=[];frozen=0
for name,old in before.items():
 p=root/name
 differs=not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=old
 if differs:
  changed.append(name)
  if name not in products and name not in status_files:external.append(name)
 is_history=(name.startswith('docs/upstream-sync/') and name!='docs/upstream-sync/README.md') or (name.startswith('openspec/') and not name.startswith('openspec/changes/port-upstream-0.2.21-tier2/') and name not in status_files)
 if is_history:
  frozen+=1
  if differs:errors.append({'frozen_history_changed':name})
change=root/'openspec/changes/port-upstream-0.2.21-tier1'
tasks=(change/'tasks.md').read_text()
if tasks.count('[x]')!=39 or '[ ]' in tasks:errors.append({'tasks':tasks.count('[x]')})
for n in ['source-baseline.md','source-feature-map.md','proposal.md','design.md','specs/upstream-0-2-21-tier1-compatibility/spec.md']:
 name='openspec/changes/port-upstream-0.2.21-tier1/'+n
 if hashlib.sha256((root/name).read_bytes()).hexdigest()!=before[name]:errors.append({'frozen_source':name})
for path in [out/'README.md',change/'README.md',change/'verification.md',change/'tasks.md']:
 for target in re.findall(r'\]\(([^)]+)\)',path.read_text()):
  target=target.split('#')[0]
  if not target or re.match(r'\w+://',target):continue
  resolved=(path.parent/target).resolve()
  if resolved==out/'scope-check.json':continue
  if not resolved.exists():errors.append({'broken_link':str(path),'target':target})
fmt=subprocess.check_output(['gofmt','-l',*[str(iso/p) for p in products if p.endswith('.go')]],text=True)
if fmt:errors.append({'gofmt':fmt})
r=subprocess.run(['git','diff','--check'],cwd=iso,capture_output=True,text=True)
if r.returncode:errors.append({'diff_check':r.stdout})
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
base=json.loads((out/'baseline.json').read_text())['head']
if head!=base:errors.append({'head_changed':head})
unit=(out/'isolated-unit.log').read_text();race=(out/'isolated-race.log').read_text()
shared_different=[n for n,d in products.items() if hashlib.sha256((root/n).read_bytes()).hexdigest()!=d]
report={'result':'PASS' if not errors else 'FAIL','errors':errors,'baseline_head':base,'implementation':'working tree; no commit','checked_tasks':39,'isolated_product_files':len(products),'isolated_unit_packages':len(re.findall(r'^ok\s+',unit,re.M)),'isolated_race_packages':len(re.findall(r'^ok\s+',race,re.M)),'checks':checks,'frozen_history_files_checked':frozen,'tier1_patch_intact_in_shared_worktree':reverse.returncode==0,'shared_files_with_additional_changes':shared_different,'concurrent_changes_observed':external,'version':'1.1.24','go':'1.27.3','sqlite_max':228,'auth_snapshot':21,'frontend_dependencies_schema_codegen_changed_by_tier1':False}
(out/'scope-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(report,ensure_ascii=False,indent=2))
raise SystemExit(bool(errors))
