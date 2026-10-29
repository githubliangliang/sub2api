"""Verify Terra's implementation against the immutable actual starting commit."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

out = Path(__file__).resolve().parent
root = out.parents[2]
baseline = "b9e51e27366e8eab20ee2c2746f77065c8ee02b6"
products = [
    "backend/internal/pkg/apicompat/anthropic_to_responses.go",
    "backend/internal/pkg/openai/constants.go",
    "backend/internal/service/billing_service.go",
    "backend/internal/service/openai_codex_models_service.go",
    "backend/internal/service/openai_codex_transform.go",
    "backend/internal/service/openai_compat_model.go",
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    "backend/internal/service/openai_gateway_messages_chat_fallback.go",
    "backend/internal/service/openai_gateway_request_body.go",
    "backend/internal/service/openai_gateway_responses_chat_fallback.go",
    "backend/internal/service/openai_model_alias.go",
    "backend/internal/service/pricing_service.go",
    "backend/resources/model-pricing/model_prices_and_context_window.json",
    "backend/internal/service/openai_gpt61_terra_test.go",
    "backend/internal/service/openai_gpt61_terra_pricing_test.go",
    "backend/internal/handler/openai_gpt61_terra_test.go",
    "backend/internal/handler/openai_gpt61_terra_sqlite_test.go",
    "frontend/src/components/keys/UseKeyModal.vue",
    "frontend/src/components/keys/__tests__/UseKeyModal.spec.ts",
    "frontend/src/composables/useModelWhitelist.ts",
    "frontend/src/composables/__tests__/useModelWhitelist.spec.ts",
]
mutable_docs = {
    "AGENTS.md", "docs/upstream-sync/README.md",
    "openspec/changes/port-upstream-0.2.19-terra/README.md",
    "openspec/changes/port-upstream-0.2.19-terra/tasks.md",
    "openspec/changes/port-upstream-0.2.19-terra/verification.md",
}
allowed = set(products) | mutable_docs


def git(*args):
    return subprocess.check_output(["git", *args], cwd=root)


tree = git("ls-tree", "-rz", "--full-tree", baseline).split(b"\0")
unchanged = []
changed = []
frozen_sha256 = {}
for entry in tree:
    if not entry:
        continue
    metadata, encoded_path = entry.split(b"\t", 1)
    path = encoded_path.decode()
    original_blob = metadata.decode().split()[2]
    contents = (root / path).read_bytes()
    current_blob = hashlib.sha1(b"blob " + str(len(contents)).encode() + b"\0" + contents).hexdigest()
    if current_blob != original_blob:
        assert path in allowed, f"Unexpected change to frozen/baseline file: {path}"
        changed.append(path)
    else:
        unchanged.append(path)
    if (path.startswith("openspec/") or path.startswith("docs/upstream-sync/PORTING-") or path.startswith("docs/upstream-sync/evidence-")) and path not in mutable_docs:
        frozen_sha256[path] = hashlib.sha256(contents).hexdigest()

for path in git("ls-files", "--others", "--exclude-standard", "-z").decode().split("\0"):
    if path:
        assert path in allowed or path.startswith("docs/upstream-sync/implementation-0.2.19-terra/"), f"Unexpected new file: {path}"

hashes = {path: hashlib.sha256((root / path).read_bytes()).hexdigest() for path in sorted(products)}
aggregate = hashlib.sha256("".join(path + "\t" + value + "\n" for path, value in hashes.items()).encode()).hexdigest()
protected_paths = ["backend/go.mod", "backend/go.sum", "backend/cmd/server/VERSION", "frontend/package.json", "frontend/pnpm-lock.yaml", "frontend/pnpm-workspace.yaml"]
protected = {path: hashlib.sha256((root / path).read_bytes()).hexdigest() for path in protected_paths}
maximum_migration = max(int(re.match(r"^\d+", path.name).group()) for path in (root / "backend/migrations").glob("[0-9]*.sql"))
assert maximum_migration == 227
version = (root / "backend/cmd/server/VERSION").read_text().strip()
assert version == "1.1.20"
record = {
    "baseline_head": baseline,
    "snapshot_head": git("rev-parse", "HEAD").decode().strip(),
    "version": version,
    "go_mod_go": next(line for line in (root / "backend/go.mod").read_text().splitlines() if line.startswith("go ")),
    "maximum_sqlite_migration": maximum_migration,
    "code_sha256": aggregate,
    "product_file_sha256": hashes,
    "changed_baseline_files": sorted(changed),
    "unchanged_baseline_file_count": len(unchanged),
    "protected_dependency_version_sha256": protected,
    "frozen_source_spec_evidence_sha256": frozen_sha256,
    "ent_wire_migrations_and_other_product_files_unchanged": True,
}
(out / "scope-facts.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({key: value for key, value in record.items() if not isinstance(value, (dict, list))}, ensure_ascii=False))
