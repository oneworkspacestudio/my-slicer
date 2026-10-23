# Studio UI revision

User requested the supplied Bambu Studio Prepare, Preview and Device layout,
then GitHub push/release and local installation. Work is on
`codex/bambu-studio-ui`, based on `eafce99`; version target is preview.5.

## Implementation

AUI shell now has charcoal File/menu and Studio tabs, green active/actions,
shared compact Printer/Filament/Process rail, full-height gray Three.js stage,
working model toolbar/transforms, dark slicing summary and layer/path sliders.
Device uses existing saved printers, actual telemetry/trays/tracked progress.
Unsupported camera/motion/extrusion/hardware-calibration controls are unavailable;
there is no simulated live data or full Bambu Studio feature-parity claim.
Project/Calibration tabs retain old printer/project/activity routes. Apps marker,
account isolation, optional projects, draft recovery and explicit print review stay.

Additive settings.process map uses one reviewed native schema in
internal/engine/process.go for both save validation and native CLI arguments.
No arbitrary keys/scripts/commands. Imported native settings take precedence.
Preview shows up to50 real layers with a single-layer fallback at existing segment
limit and translucent clipped model context. 3MF overview supports up to9 plates;
larger projects select one. Source origins follow pinned Studio PartPlateList
ceil(sqrt(count))/20% gap convention; view transforms do not alter source archives.

Seven Webframe screens/PNGs and documentation are updated; feedback IDs preserved.
Flowframe all-screens/Prepare→Preview→Device→setup passed. Vinext SSR did not
hydrate on this host, so the unchanged Flowframe client component was bundled in
task tooling for review; product repository was not edited.

## Validation and delivery state

Go tests/vet passed after the shared native-schema adjustment.
AUI/TypeScript builds and native plate-coordinate regression passed. Isolated
native acceptance proved engine A1/A1 mini validation, real slicing/preview,
desktop/phone Device, import/reload/printer setup/retry. It caught a dynamic AUI
action bug (fixed using supported item selection) and native's second override
whitelist (now uses shared validation). Full final native acceptance passed: persisted speed override was accepted by the
real engine, 3MF round trip passed, browser errors0.
Named projects intentionally require reopening from Project after page reload;
automatic restore applies to drafts. Test now respects that existing behavior.

Task artifacts: I:/Projects/tooling/slicer-studio-20261023. Tests only in named
NTFS C:/Projects/.cache/my-slicer-studio-20261023. Never use owner data as fixture,
contact physical test printers, print, or remove the latest owner rollback.

Native screenshots inspected and curated under docs/screenshots/slicer5-*.
Deployment must use the normal protected app-only updater selecting slicer.
Preserve other app contracts, versions, identities, paths and ports. Do not
overwrite unrelated dirty Suite navigation work or interfere with other updates.


## Publication

Final public [preview.5.2 packaging](https://github.com/oneworkspacestudio/my-slicer/releases/tag/v0.1.0-preview.5.2)
source commit: 24bb9b4f02e9517ec69680425122b43a2a711054. Runtime version remains
0.1.0-preview.5. Main and all release tags were pushed. All13 published asset sizes and
GitHub SHA256 digests matched local files before publication. Release EXE exactly
matches the full native/browser-tested candidate:
`5e62a9ece380ea84f74a4c69e26d99d924661841e5aaced2f6fea2907aaa33e4`.
Includes native ZIP/EXE, managed app manifest, source ZIP, engine manifest,
dependency/license inventory and acceptance. Actions remained disabled.

The immutable preview.5 and preview.5.1 releases were superseded because their
managed manifests failed the current catalog/schema checks. No installed bytes
were changed by those refusals. Preview.5.2 uses Workspace's native publisher:
schema2, integration3, minimum manager72 and exactly the four existing Slicer
package IDs. All three releases contain the same fully tested executable.

Anonymous GitHub checks hit HTTP403 during concurrent app updates. The supported
offline update path uses checksum-verified public Workspace73 bundle/installer
and final published Slicer5.2 payloads. ZIP members are copied unchanged except
the four Slicer payloads and catalog app record; the native updater re-verifies
the archive and retains normal backup, rollback and restart checks. No tokens
were retrieved or passed to app processes. An earlier Workspace72 bundle was
correctly refused after a concurrent manager upgrade to73.

Stopped synthetic installations were removed after process-reference/reparse
checks; both review servers stopped. Other tasks' My3D/Workspace/Social updates
were allowed to finish. The human authorized chat coordination for a protected
Slicer installation window; the Social task agreed to hold further owner updates.
Never kill another updater, remove locks, or use owner auth/data as a fixture.

The normal Slicer-only update reached a complete journal. Independent
verify-installed.py checked against that transaction's verified backup: all four
installed Slicer payloads match final5.2, Slicer5 health passes, all other app
contracts/package hashes and identity/location/port/selection values unchanged.
Sanitized owner-install-receipt.json records those checks. The native updater
exited0 after cleanup freed4,409,519,447 bytes, removed one older automatic
backup and retained the verified recovery copy. An older backup failed
verification and was safely retained for separate attention. The strict wrapper
then rejected an all-apps-healthy assertion because Social failed; the update
itself committed successfully, so it must not be rerun or rolled back.

Restart exposed a separate Social3 migration bug: its Migrate(1) call rejects
already-migrated schema2 on the next startup. Social's installed EXE hash still
matches the prior published/tested Social3 package; an initial tentative bundle
overwrite theory was disproved by the exact hashes and catalog comparison.
The authorized coordinating Social task is preparing a regression-tested release
and protected repair; it has been explicitly told that Slicer's native updater
exited and the window is available. Do not edit or downgrade owner Social data.
Final suite-health verification remains pending that repair.
