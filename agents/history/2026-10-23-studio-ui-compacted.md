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
real engine, 3MF round trip passed, browser errors0. Publication pending.
Named projects intentionally require reopening from Project after page reload;
automatic restore applies to drafts. Test now respects that existing behavior.

Task artifacts: I:/Projects/tooling/slicer-studio-20261023. Tests only in named
NTFS C:/Projects/.cache/my-slicer-studio-20261023. Never use owner data as fixture,
contact physical test printers, print, or remove the latest owner rollback.

Publish only the app with workspace-update.json; use normal protected app updater
selecting slicer after verified release. Preserve all other app contracts/versions.
Current owner installation C:/tmp/workspace/Private is71.1/Slicer4, with an
unrelated Workspace update/recovery already running and transaction testing.
Another task's recovery must finish before installing Slicer. Maintenance reports
insufficient NTFS backup space (needs8806MB); own failed fixture cleanup is safe
after checking no process references/reparse points. Do not overwrite dirty Suite
navigation changes or interfere with concurrent My3D/Workspace work.

Native screenshots inspected and curated under docs/screenshots/slicer5-*.
Next: commit source, clean release build.
Require release EXE SHA256 to equal the fully tested candidate. Prepare verified
app5 assets/source/notices, push main/tag, publish draft only after every GitHub
asset digest matches. After other recovery completes and sufficient space exists,
apply only Slicer through backed-up updater; verify deployed hash, health and
unchanged other apps/identity/paths; record receipts and stop/clean fixtures.
