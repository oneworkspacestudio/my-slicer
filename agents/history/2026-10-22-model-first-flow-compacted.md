# Model-first slicer workflow and printer setup

User reports Add printer crashes and asks to redesign the flow: Add model above
the fold, no mandatory project creation, clear screens. Continuing the session's
authorized local update, source push and GitHub release. No Actions, owner-data
fixtures, physical printer contact or automatic print. Read initial app history
and Workspace reinstall64 handoff; owner now Workspace64/Slicer1, 25 apps and
24 healthy services, independent Chat9/Mail17 retained. App source was clean at
e9aac68 on I:/Projects/my-slicer; work branch codex/slicer-model-first-flow.

## Implementation

Complete six-screen .webframe contract: Prepare, Preview, Printers, Printer setup,
Saved projects, Activity. Generated visible wireframe-images, inspected all six,
loaded all screens and exercised connected Preview flow in the existing local
Flowframe reviewer. No feedback comments. AUI owns every new UI tree; generated
TSX regenerated. README links visible screen PNGs and plain docs/USER-FLOW.md.

Prepare puts immediate Add model before the canvas, with optional Drive import,
Save project and New plate. Empty plate renders immediately; no project creation
on visit. On import, backend's existing revisioned Project is created lazily as a
Draft. Additive JSON draft flag defaults false for old named projects. Recent
plates are durable and latest nonempty draft restores after reload. Optional save
names/promotes the same plate without losing models. New plate leaves previous
work available. Native 3MF opens its own plate when mixing would discard native
layout; old active slice and cross-project undo are reset. Immutable revision,
ownership, original source and physical-print protections retained.

Printer setup is a routed Details → Trust → Connection check page, usable before
any model/project. Manual entry survives validation errors; access code clears on
leaving. Saved configuration and diagnostics remain accessible. Does not discover,
probe, connect or print merely on opening. Only actual live paired status enables
reviewing a physical print. Existing dispatch still requires explicit confirmation.

Fresh old Chrome Add printer succeeded. A reproducible compatibility crash was
confirmed by removing native dialog.showModal: old app threw TypeError and blanked
the root. The user's exact browser was not established. New printer page needs no
modal; remaining dialogs have native/open fallback, focus trap, Escape/restore
handling. A rendering boundary offers reopening without deleting server data.

Real A1 mini acceptance found a profiles race: engine-ready enabled Slice before
the matching profile request completed. Fixed readiness binding to the selected
model and compatible selected profiles, with cancellation of stale responses.
No engine/profile upgrade or changed hardware claims. Engine remains pinned
official v02.08.02.61, with source/license links in Settings.

## Verification and delivery resume point

Go packages/vet and frontend TypeScript/build pass. Reusable native Windows browser
regression is web/scripts/workflow.test.mjs, skips unless MYSLICER_TEST_EXE supplied.
It requires MYSLICER_TEST_ROOT under a named C:/Projects/.cache/my-slicer-* root.
Set MYSLICER_TEST_OUTPUT on I:; optional MYSLICER_TEST_ENGINE_ARCHIVE enables actual
verified install and slice roundtrip. Tested: Add model above fold at1366/390,
no project on visit, repeated Add printer with no native modal or project, actual
encrypted dummy printer save/validation recovery, code clearing, automatic draft
import/reload, optional save/fallback modal/Escape, preserved named project/new
plate, real A1 mini 50-layer slice and native3MF reimport/new plate/re-slice.
No browser errors. Literal unpaired fixture printer has no pins and is never
probed/trusted/tested; no device or print request sent. Monitor skips unpaired
records. Invalid candidate runs exposed fixture filename and profile race;
final candidate corrected and passed. Screenshots/receipts outside repo on
I:/Projects/tooling/slicer-flow-20261022/final-native.

Next: package committed app2 source with scripts/build.ps1 -Release, repeat native
test against exact packaged EXE, publish all app2 hashes/source/notices, then pin
Slicer2 in Workspace65 while keeping every other app contract and byte unchanged.
Build manager/full65 from an immutable committed suite export, verify all payloads,
perform isolated managed upgrade/SSO acceptance before authorized normal backed-up
owner update. Publish, verify local/public health and installed bytes, record proof
and delivery follow-up commits without moving source tags. Preserve latest owner
rollback and existing24 other app versions (incl independent Chat9). No functional
test against live MyAccounts/Drive. Artifact root I:/Projects/tooling/slicer-flow-
20261022; NTFS fixtures C:/Projects/.cache/my-slicer-20261022. All completed native
browser fixtures stopped in finally. Stop task Flowframe39124 and verify children;
bounded cleanup exact named C root only after final processes stop. No global temp
sweep. Physical hardware, AMS, clean-host/reboot/security acceptance unperformed.


## Managed visual-review follow-up

App2 source c84a51c and public release2 passed exact packaged native tests and real
Workspace64-to65 managed SSO/Drive upgrade. Source65 is363c380; publication and
normal backed-up owner64-to65 update were underway when detailed managed phone
review found the pre-existing floating Workspace launcher over object buttons and
printer labels. Disclosed it and made app3 recognize Workspace's existing native
trigger contract on the header Apps button. The floating launcher is then hidden;
header Apps opens the suite picker, with app-local fallback if showModal is absent.
The marker's button handler prevents duplicate delegated clicks. No model/printer
storage changes from2 to3. Keep public2 and source65 immutable; ship final3/66 and
preserve65's eventual publication/local update results truthfully. Do not interrupt
an active owner backup/update. Verify current receipts before normal65-to66 update.

Owner65 attempt stopped at backup-capacity preflight before any replacement.
The old64 manager restarted automatically; all24 services healthy and old journal
complete. Preserve that state. Final normal update is64-to66, after removing only
task cache copies and running the exact verified installer/bundle from I: so C:
has sufficient protected recovery capacity. Keep Downloads65; copy66 after update.
