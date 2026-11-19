"""Validate GPT-6.2 scope, frozen source bytes, local links and final hashes."""
import difflib
import hashlib
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
OUT = Path(__file__).resolve().parent
CHANGE = ROOT / "openspec/changes/port-upstream-0.2.23-gpt62"
EXTERNAL = {".github/workflows/release.yml", "deploy/tests/release-publish-config-test.sh"}
DOC_PREFIXES = ("openspec/changes/port-upstream-0.2.23-gpt62/", "docs/upstream-sync/implementation-gpt62/")
OWNED_CODE = {
    "backend/internal/handler/sol_ultrafast_sqlite_test.go",
    "backend/internal/pkg/apicompat/anthropic_to_responses.go",
    "backend/internal/pkg/openai/constants.go",
    "backend/internal/pkg/openai/codex_gpt62_astra.json",
    "backend/internal/pkg/openai/codex_gpt62_sol.json",
    "backend/internal/service/account_stats_pricing.go",
    "backend/internal/service/account_stats_pricing_test.go",
    "backend/internal/service/billing_service.go",
    "backend/internal/service/billing_service_test.go",
    "backend/internal/service/gateway_usage_billing.go",
    "backend/internal/service/openai_codex_model_metadata.go",
    "backend/internal/service/openai_codex_models_service.go",
    "backend/internal/service/openai_codex_transform.go",
    "backend/internal/service/openai_compat_model.go",
    "backend/internal/service/openai_compat_prompt_cache_key.go",
    "backend/internal/service/openai_gateway_messages_chat_fallback.go",
    "backend/internal/service/openai_gateway_request_body.go",
    "backend/internal/service/openai_gateway_responses_chat_fallback.go",
    "backend/internal/service/openai_gateway_usage.go",
    "backend/internal/service/openai_gpt62_compatibility_test.go",
    "backend/internal/service/openai_gpt62_pricing_test.go",
    "backend/internal/service/openai_model_alias.go",
    "backend/internal/service/pricing_service.go",
    "backend/resources/model-pricing/model_prices_and_context_window.json",
    "frontend/src/components/keys/UseKeyModal.vue",
    "frontend/src/components/keys/__tests__/UseKeyModal.spec.ts",
    "frontend/src/composables/__tests__/useModelWhitelist.spec.ts",
    "frontend/src/composables/useModelWhitelist.ts",
}
INDEXES = {"AGENTS.md", "docs/upstream-sync/README.md", "docs/upstream-sync/PORTING-gpt62.md"}

def git(*args):
    return subprocess.check_output(["rtk", "proxy", "git", *args], cwd=ROOT, text=True)

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

tracked = git("diff", "HEAD", "--name-only").splitlines()
untracked = git("ls-files", "--others", "--exclude-standard").splitlines()
changed = sorted(set(tracked + untracked))
unexpected = [name for name in changed if name not in OWNED_CODE | INDEXES | EXTERNAL
              and not name.startswith(DOC_PREFIXES)]
assert not unexpected, unexpected
for name in ["AGENTS.md", "docs/upstream-sync/README.md"]:
    original = git("show", "38990cfa716e337740ae8aa3c316940368a3f9ce:" + name).splitlines()
    current = (ROOT / name).read_text().splitlines()
    operations = difflib.SequenceMatcher(None, original, current, autojunk=False).get_opcodes()
    assert all(operation[0] in ("equal", "insert") for operation in operations), name

frozen = {
    "docs/upstream-sync/evidence-0.2.23/patches/PR8906.patch": "146eaea850df80d133fcc6caa42bfd0d9a065e2710bbbd1bc117643300298c4e",
    "backend/internal/pkg/openai/codex_gpt62_sol.json": "62469f05d46e7108e573a4746490edc6acd8931fb9b259e326c33e61f0c9cf49",
    "backend/internal/pkg/openai/codex_gpt62_astra.json": "d5e1e3e3dcc06e9648bc5eefc3b3d09bd5412046d60e1be3b3f3a45ec3f6b16c",
}
assert all(digest(ROOT / name) == expected for name, expected in frozen.items())
assert (ROOT / "backend/cmd/server/VERSION").read_text().strip() == "1.1.26"
assert re.search(r"^go 1\.27\.4$", (ROOT / "backend/go.mod").read_text(), re.M)
preserved_paths = ["backend/go.mod", "backend/go.sum", "frontend/package.json", "frontend/pnpm-lock.yaml",
                   "frontend/pnpm-workspace.yaml", "backend/ent", "backend/migrations",
                   "backend/cmd/server/wire.go", "backend/cmd/server/wire_gen.go",
                   "backend/internal/config", "deploy/config.personal.sqlite.yaml",
                   "backend/internal/service/api_key_auth_cache.go", "docs/releases"]
assert not git("diff", "38990cfa716e337740ae8aa3c316940368a3f9ce", "--name-only", "--", *preserved_paths).strip()
tasks = (CHANGE / "tasks.md").read_text()
assert len(re.findall(r"^- \[x\]", tasks, re.M)) == 12
assert not re.search(r"^- \[ \]", tasks, re.M)
hashes = {name: digest(ROOT / name) for name in sorted(OWNED_CODE)}
(OUT / "product-hashes.json").write_text(json.dumps(hashes, indent=2) + "\n")
(OUT / "scope-validation.json").write_text("{}\n")

files = [ROOT / "docs/upstream-sync/PORTING-gpt62.md", *CHANGE.rglob("*.md"), *OUT.glob("*.md")]
links = 0
for file in files:
    for target in re.findall(r"\]\(([^)]+)\)", file.read_text()):
        if target.startswith(("https:", "http:", "#")):
            continue
        destination = target.split("#", 1)[0]
        if destination:
            assert (file.parent / destination).exists(), (str(file), target)
            links += 1
result = {
    "entry_head": "38990cfa716e337740ae8aa3c316940368a3f9ce",
    "current_head": git("rev-parse", "HEAD").strip(),
    "owned_product_and_test_files": len(OWNED_CODE),
    "unexpected_changes": unexpected,
    "external_changes_observed": [name for name in changed if name in EXTERNAL],
    "external_file_snapshot_sha256": {name: digest(ROOT / name) for name in EXTERNAL},
    "frozen_source_hashes_match": True,
    "historical_index_content_preserved": True,
    "old_changes_and_evidence_preserved": True,
    "version_dependencies_schema_generated_defaults_unchanged": True,
    "tasks_completed": 12,
    "local_links_checked": links,
    "broken_links": [],
    "index_staged_files_observed": git("diff", "--cached", "--name-only").splitlines(),
}
(OUT / "scope-validation.json").write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(result, ensure_ascii=False, indent=2))
