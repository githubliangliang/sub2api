# Browser acceptance

Final command and environment: `command.json`; executable script: `run.mjs`; raw output: `seventh.stdout.log` / `seventh.stderr.log`; final assertions: `results.json`; actual intercepted request JSON: `requests.json`.

Google Chrome headless runs the real Vite `/admin/users` route, UsersView, nested definition editor, UserEditModal/UserAttributeForm and UserAllowedGroupsModal. Both 1440×1000 and 390×844 pass all six scenario groups (12 total), with zero page exceptions:

1. Existing user editor submits zero/negative custom integers as JSON strings, preserves text and min/max inputs.
2. Existing nested definition editor sends explicit empty description and placeholder strings.
3. A delayed attribute read from a closed user editor cannot replace another user's current edited values; normal current-user save succeeds.
4. A failed group load leaves save disabled and emits no user PUT. Existing refresh retries successfully; exclusive deselection and explicit rate deletion retain public default semantics.
5. A delayed group read from a closed session cannot replace another user's selection or rate. Only the current user is saved.
6. A pending originating user save retains its original user and attribute snapshot, while its completion cannot close or overwrite a newly opened user's draft; the new session also saves normally.

`desktop-*-group-load.png`, `mobile-*-group-load.png` and `*-new-user-session.png` are actual browser screenshots. The mobile new-user screenshot was visually checked after the run.

All API responses and writes are intercepted synthetic fixtures. The script does not connect a business instance or claim browser-to-SQLite integration. Root records independent real HTTP/SQLite checks.

## Preserved script fixture failures

Every preceding round's raw stdout/stderr remains unchanged. No product code was changed to eliminate these browser fixture errors:

- `first` / `second`: response URL regular expression anchored at `/attributes$` omitted the real API client's automatically appended `timezone` query. Registering the response wait earlier did not resolve it; inspecting `frontend/src/api/client.ts` confirmed the query, and the matcher now uses URL pathname.
- `third`: Playwright's default `getByTestId` expects `data-testid`; this project button uses `data-test`. The fixture now selects the real attribute.
- `fourth`: direct pointer clicking the screen-reader-only checkbox collided with its existing styled indicator. The fixture now clicks the visible label, as an operator does.
- `fifth`: mobile scrolling to another user's More button fired UsersView's existing scroll-to-close handler, closing the action menu. The fixture pre-scrolls, lets that scroll finish and opens the menu afterwards.
- `sixth`: after explicitly clearing the note placeholder, a broad “last text input” selector could initially identify Username before asynchronous attribute definitions rendered. The fixture now finds the exact Note label and its own input, waiting for the actual attribute field.

Final `seventh` completed successfully in both viewports. These failures describe the test harness and are not counted as product defects or passing acceptance.

## Final built-product step-up followup

After the additional real-controller regression found the intermediate step-up lifecycle gap, root completed the final frontend production build and served it through preview on the same local port. `run-stepup.mjs`, `stepup/command.json`, `stepup.stdout.log`, `stepup.stderr.log`, `stepup/results.json` and `stepup/requests.json` record the built-product acceptance, separate from the prior Vite-source twelve-scenario record.

Both 1440×1000 and 390×844 pass three step-up groups each (six total) with zero page exceptions: a closed user's late STEP_UP_REQUIRED cannot prompt/retry into another user's editor; local Escape cancels the current prompt and a later current verification retries exactly once; an old verification success arriving after a different prompt opens cannot verify/close/retry that newer session, whose own normal code verification still succeeds. All request bodies use synthetic intercepted APIs, including TOTP grants. Current prompt and completion screenshots live in `stepup/`.

The stale verification-failure toast is separately reproduced before the fix and accepted after it using the real dialog/component unit regression. No remote TOTP, real authorization grant, business instance or database is used by these browser fixtures.
