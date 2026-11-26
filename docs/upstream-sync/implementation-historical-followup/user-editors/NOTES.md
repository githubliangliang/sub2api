# F03 / S01 implementation evidence

Actual start: `89060f532a9000af28a47036f70cc44821dafd55`; root froze the worktree and protected historical inputs in `../initial-state.json`. This directory contains only the user attribute/group editor subset; root records backend HTTP/SQLite and combined gates separately.

## Product changes

- `UserAttributeForm.vue`: numeric input reads the original HTML input string; existing integer input, required/min/max and other field controls remain. Values only apply to the current user/request; user changes clear old values, failed/pending user reads keep attribute inputs disabled, stale errors and unmount callbacks are ignored. New-user model updates may explicitly clear the map.
- `UserAttributesConfigModal.vue`: explicitly sends `description` and `placeholder` strings, including `''`, rather than dropping empty fields from JSON. Backend omitted-field semantics are unchanged.
- `UserAllowedGroupsModal.vue`: each visible user/load has its own version, start-time grant/rate snapshot, loaded identity and readiness. Pending/failed loads block button and method saves. The existing `common.refresh` label retries a failed load. Saves snapshot user/config/rates, preserve public default access and selected exclusive IDs, retain null rate deletion, prevent duplicates and ignore stale success/error/finally. Local Cancel/dialog-close invalidates immediately, before parent props update.
- `UserEditModal.vue`: user/show session owns form and attribute child lifetime. Submit captures the user and attribute map before awaiting the existing user/step-up update. A submitted originating save can complete with its original snapshot; its result/error/finally cannot change another session. Cancel/dialog-close invalidates immediately, before parent props update. Existing concurrency validation and step-up behavior remain.

Attribute value reads have no new standalone retry UI. Closing/reopening the user editor creates a fresh attribute child and retries through the existing dialog flow; a regression covers an old pending read arriving after reopening. If a read fails, saving other user fields retains the existing omitted-attribute-write behavior and leaves stored attributes alone.

## Test evidence

- `before-targets.*`: 31 tests, 23 target failures and 8 passing controls on unchanged product. Numeric string/zero/negative failures, explicit clears, stale reads, load readiness and originating-save snapshots reached their intended assertions.
- `before.*` and `before-fixture-corrected.*`: retained first fixture stages. The first session-test module mocked only `useI18n`; a real imported dependency required `createI18n`, so the fixture now partially mocks i18n. Group tests now open through the same false→true transition as the current parent, preserving a meaningful passing normal control.
- `after.*`: first combined implementation passed 31/31.
- `final.*`: expanded close/reopen/retry/unmount suite plus existing UserEditModal and UsersView controls passed 37/37.
- `close-ownership.*`: local-close coverage initially had two test fixture errors (unnamed BaseDialog stub and generic secondary-button selector selecting password generation). Original errors remain here.
- `close-ownership-corrected.*`: final seven-file suite passed 40/40: five new regression files plus the two existing UserEditModal/UsersView files. Covers old success/error, pending-current state, close/reopen, retries, current loaded snapshots, public/exclusive/rate-clear semantics, duplicate submits, unmount and local close before parent props change.
- `eslint-final.*`: all four product files and five new test files pass targeted ESLint. Full frontend typecheck/lint/build and combined suites are recorded by root.
- `final-hashes.json`: SHA256 of these four product files and five new regression files.

All fixtures use synthetic users/groups/attributes. No frontend dependency, generated code, schema, version or resource default was changed.

## Step-up session followup from independent review

The initial session suite replaced `useStepUp.run` with a direct action call, so it did not cover the composable's intermediate STEP_UP_REQUIRED prompt/retry. Independent review traced a late old-user 403 through the real controller: the outer user-editor catch ran too late to stop a prompt or prevent it sharing a newer pending prompt.

`UserEditModal.stepup.spec.ts` uses the real `useStepUp` composable and real `TotpStepUpDialog` component. `stepup-before-full.*` preserves five desired failures and one passing normal control: old required response opens a new prompt; editor close leaves a prompt active; an old action shares and retries after a new prompt (actual IDs `[1,2,2,1]`); old TOTP success closes a newer prompt; old TOTP failure toasts into it. The earlier five-case `stepup-before.*` is also retained.

The fix remains local to this existing editor flow. `UserEditModal` checks the originating session before action execution/retry and again before returning a request error to `useStepUp`, translating obsolete actions to the existing `StepUpCancelledError`. Session reset cancels a pending prompt immediately. Its TOTP child is mounted only for the active editor and keyed by session.

One necessary additional existing consumer boundary is `components/auth/TotpStepUpDialog.vue`: its old asynchronous verification could still call the shared controller or toast after unmount. A five-line active/onBeforeUnmount guard drops both late success and failure. The composable API and other consumers' mounted behavior are unchanged. The new session regression file no longer replaces the composable.

`stepup-after.*` passes all 58 cases across the final nine-file user/editor/composable suite. Six real-controller/dialog cases include successful current verification with exactly one retry. `eslint-stepup.*` passes the newly changed component/test subset. `final-hashes.json` now records five product files and six new test files; `pre-stepup-hashes.json` preserves the previously accepted four-source/five-test hashes, and the earlier 40-case/browser records keep their original snapshot meaning. Root reruns the combined frontend gates on the new final product.
