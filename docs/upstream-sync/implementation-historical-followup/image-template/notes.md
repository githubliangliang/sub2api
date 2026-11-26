# S02 / S03 implementation checks

The root task owns the actual starting-state snapshot, combined gates, browser checks and commit. This subtask changes only ImageUpload, MonitorTemplateApplyPickerDialog and their new regression tests. Frozen source PR7320/PR7365 and the previous assessment remain unchanged.

- `before.json` records the command against the original product: 26 tests, 18 desired failures and 8 passing controls. Full Vitest JSON and stderr are retained in `before.stdout.log` / `before.stderr.log`.
- `after.json`: the initial repair plus the pre-existing parent manager suite passed 31/31.
- `final.json`: two additional same-template retry/method-level controls passed, 33/33 including the parent manager suite.
- `final-extended.json`: direct template switching during apply was added alongside close/reopen; 35/35 passed (30 new regressions/controls and 5 existing parent manager tests).
- `settings-control.json`: existing SettingsView suite passed 34 tests, with 2 existing skips, 0 failures. It exercises the unchanged Settings consumer. No payment operation was performed.
- Targeted ESLint command (exit 0): `pnpm --dir frontend exec eslint src/components/common/ImageUpload.vue src/components/common/__tests__/ImageUpload.readOrder.spec.ts src/components/admin/monitor/MonitorTemplateApplyPickerDialog.vue src/components/admin/monitor/__tests__/MonitorTemplateApplyPickerDialog.requests.spec.ts`. Its stdout/stderr are retained in `lint.*.log`.
- `source-hashes.json` hashes the final four source/test files. Root may independently repeat these hashes after combined validation.

Image reads are invalidated before abort. Both load and error handlers check reader identity, so even a queued callback after abort cannot emit or replace the current error. Selection of a rejected oversized/non-image file also cancels the previous read. Removal clears pending read/error state and emits the existing empty-string update. Unmount cancels the read. The default 300KiB limit, image MIME restriction, SVG raw-text trim and inline DOMPurify sanitization are unchanged; tests include exactly-at-limit acceptance, over-limit rejection, normal read error/recovery and removal of SVG script/onload content from the preview.

The picker resets state with each show/template session and has an independent list request identifier for same-template retries. Stale list success/error/finally callbacks are ignored. Apply requires current successfully loaded data and captures the template/selected IDs before awaiting. Session changes suppress old success/error/finally callbacks without affecting a newer apply. Cancel invalidates immediately before emitting close. Normal subset selection, select-all/none, method-level duplicate prevention, retry and current errors remain covered.

Browser hooks: ImageUpload retains `input[type=file]`, removal `button[type=button]` and the SVG inline `span`; no new UI is introduced. The template picker retains `li` monitor rows, checkbox inputs, `button.btn-secondary` cancel and `button.btn-primary` apply. Its real route remains ChannelMonitorView → MonitorTemplateManagerDialog → MonitorTemplateApplyPickerDialog.

Actual Chrome acceptance was subsequently completed at desktop1440 and mobile390 via the real Settings and administrator monitor routes. [Browser evidence](browser/README.md) separates controlled FileReader callbacks from native image/SVG reads, synthetic HTTP responses from business data, and preserved fixture failures from the final accepted run. Both viewport flows pass with empty pageErrors and exact template selection snapshots. The original Vitest checks remain jsdom tests.
