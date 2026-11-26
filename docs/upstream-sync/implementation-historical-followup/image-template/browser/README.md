# Chrome acceptance: S02 and S03

Command, run from the repository root:

```sh
rtk proxy node docs/upstream-sync/implementation-historical-followup/image-template/browser/acceptance.mjs
```

The final run is `run-4.stdout.log` / `run-4.stderr.log` (exit 0), with structured request snapshots in `result.json` and the four unprefixed desktop/mobile screenshots. Google Chrome (`/usr/bin/google-chrome`) runs through the existing external Playwright installation. Desktop is 1440×1000; mobile is 390×844. The real Vite application at `127.0.0.1:18926` renders SettingsView and the registered `/admin/channels/monitor` route through the real manager and picker. Every `/api/v1` response is synthetic; no business instance, S3 connection, backup, payment operation or real monitor apply occurs.

Both sizes pass the Settings logo double-read/abort/late-error/delete/rejected-next-file controls, unchanged 300KiB rejection, then a real native FileReader PNG read and native SVG read/DOMPurify preview. The race phase replaces FileReader in Chrome with controllable callbacks delivered even after abort; the normal PNG/SVG phase restores the browser implementation. No Vue component internals are exposed or edited by the script.

Both sizes also pass the actual template picker flow: template 1's held association request resolves after reopening template 2; current rows and selection survive. Applying template 2 captures `[20]`; cancelling and reopening template 1 before that apply completes leaves the new picker open and its current selection intact. Normal template 1 apply then sends only `[10]` and closes. `pageErrors` is empty. Final Settings screenshots were inspected and show the sanitized SVG with no fixture-error toast. Mobile template screenshot was inspected and shows current template 1 with monitor 10 selected and monitor 11 unselected.

Preserved fixture history:

- Run 1 timed out waiting for a zero-area SVG path to be visible. The component had already attached the sanitized path; the locator was corrected to test attachment. `run-1.failure.json`, `run-1.result.json` and stderr preserve it.
- Run 2 timed out because the request predicate matched the whole URL with `endsWith`, whereas the API URL contains query parameters. Its recorded requests show the target association read was actually issued. The predicate now compares URL pathname. `run-2.source.mjs`, `run-2.failure.json`, `run-2.result.json` and stderr preserve it. A contemporaneous Vite reload may also appear in the request trace.
- Run 3 passed the requested interactions but visual inspection found caught errors in unrelated Settings children because the synthetic email-template/image-storage responses were incomplete. `run-3.*` preserves its source, result and screenshots. Only the fixture response shapes were completed; run 4 asserts those error toasts are absent and passes again.

The dev server's injected public branding displays v1.1.15 from its existing proxy configuration. These screenshots verify the current source components served by Vite, whose implementation hashes are recorded separately; they are not deployment/version acceptance.
