# S01 implementation evidence (2026-11-24)

Implemented from local HEAD `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`; initial repository/product/protected hashes are in `../initial.json`. F01 account redaction is independently implemented and tested by the parent task. No commit, push, release, archive or instance operations.

## Implementation

- Existing runtime wrappers share the detailed rule result; final passthrough rules use the original status/body and account platform. Explicit administrator text stays verbatim; inherited upstream text uses the F01 helper after existing agent-identity redaction. Monitoring gets the same original-body match without retaining raw bodies when detail logging is disabled.
- Only allowed deterministic 400 errors (invalid encrypted content/context window) or an explicit passthrough-body rule expose the standard scalar fields. Numeric fields become strings; objects/arrays/bools and identifiers over 256 bytes before or after redaction are dropped. Missing/null/blank outer codes use the existing embedded-JSON boundary with the same scalar restrictions on nested values. Existing non-400 context errors preserve their sanitized message without new field disclosure; unrelated 403 stays generic.
- After a compact heartbeat, the error writers emit one ordered `response.failed`, retain deterministic codes and rewrite capacity codes to the existing retryable value. The local sequence allocator and internal upstream classification remain intact. Both custom mapped 2xx and non-2xx errors stay failures after heartbeat; before heartbeat they retain the configured HTTP status and JSON error envelope. Repeated failure writes do not consume sequence numbers or append bytes.

## Commands and results

Exact commands, exit codes, elapsed times, test names and complete output are preserved in the JSON/log pairs.

| Evidence | Result |
|---|---|
| `before.json` / `.log` and `baseline-overlay.json` | Frozen test-only intake probe overlay against unmodified implementation: five desired top-level failures, four controls passed. Includes F01 leakage, structured fields, complete rule decisions, compact code and raw-body-only monitoring. |
| `monitoring-priority-before.json` / `.log` | Additional regression reproduced: message-only fallback selected a lower-priority skip rule before the final raw-body match; failed as intended before the narrow optional-original-body fix. |
| `selected.json` / `.log` | First expanded run: nine top-level passed, two failed because new fixtures omitted API-key credentials and Forward rejected before reaching the synthetic upstream. Fixed test credentials only; no product change for these two failures. |
| `selected-final.json` / `.log` | 12 top-level / 77 including subcases passed. Covers source-adapted PR9045 forwarding cases plus scalar byte/type boundaries, embedded malformed codes, raw priority/platform rules, real local HTTP compact matrix (12 cases), single write/sequence, failed-result accounting and canceled-before-heartbeat control. |
| `race.json` / `.log` | `go test -json -race -tags=unit ./internal/service ./internal/handler -run 'Test.*(Passthrough\|Compact\|ErrorPassthrough\|OpsStream\|SkipMonitoring\|CapacityShed\|ResponseCommitted\|ContextWindow\|EncryptedContent\|ResponsesSequence\|StructuredError)' -count=1`: both packages passed;392 top-level /800 including subcases passed. |
| Parent combined gates | Full backend unit, build/lint, additional race, actual SQLite usage/settlement/dedup/fixed-quota, dependency gates and final protected hashes are recorded separately in the parent evidence. |

An initial interactive gofmt invocation used root-relative paths while its working directory was backend and reported five `lstat ... no such file` errors. It changed no files; rerunning from repository root formatted the intended files, then selected tests passed. Initial lint identified an unchecked keepalive type assertion in the test fixture; only that test assertion was corrected after the selected/race runs. Parent final targeted checks and lint cover that test-only correction. `final-hashes.json` records the resulting file hashes.

The new forwarding test file adapts the frozen PR9045 tests, excluding source redaction-helper tests that have independent local F01 coverage. The lifecycle tests add local HTTP/sequence/priority/type-boundary controls. Actual HTTP uses fresh local synthetic upstream/downstream servers; this is not an external provider or business instance test. No new frontend files, migrations, generated files, defaults, model/billing logic or success-path copies/caches were introduced by S01.
