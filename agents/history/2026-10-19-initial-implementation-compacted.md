# My Slicer initial implementation

User requested implementing the attached A1/A1 mini brief, ecosystem design/SSO/
Drive/installer integration, local tunnel update, public org repo and local-built
releases. No GitHub Actions. Both physical printers are available on LAN; pairing
credentials must be entered privately into wizard. Actual printing needs user
confirmation, not automated test dispatch. Software and hardware acceptance must
be distinguished. All stages, including direct print adapter, are initial scope.

Source I:/Projects/my-slicer, initialized codex/initial. Existing Workspace root
I:/Projects/one-workspace main c147009; live owner preview41, Tube2/PDF3. Preserve
owner C:/tmp/workspace roots and newest recovery copy. Synthetic tests only on
named C:/Projects/.cache/my-slicer-20261019, with large artifacts on I:.

Pinned candidate official stable engine v02.08.02.61 (source926a719). Native ZIP
473534359bytes, SHA2563faaa194a90a2e7defa39a6a63f5a654b8cdb762c6a98556cf771fba582229f5.
Download and upstream sparse source/reference checkouts under
I:/Projects/tooling/my-slicer. Resolve all inherits/includes against release-matched
resources and actually slice both machines before claiming support. Download
official package automatically in app with visible progress; no plugin reuse.
Shared auth/UI foundation is apphost from public Workspace41 and existing AUI
Shell/Login/Settings + semantic components, with original authorization preserved.

## Real native engine milestone completed

Official archive verified before extraction. Runtime minus PDB/debug symbols and
development headers has6753 files/570212481bytes; download remains473534359bytes.
Compiled JSON checksums in internal/engine/runtime-manifest.json. Main EXE/ZIP
will not bundle the engine; Manager installs from official source with visible
byte/file stages, validates hashes, runs both native slice checks before promotion,
retains previous version, supports cancellation and restart verification. Uses a
bounded4GB Windows job with suspended-start assignment and process-tree teardown.
`MYSLICER_ENGINE_ARCHIVE` is a server-only offline/reviewed-archive override.

Profile resolution indexes BBL.json's declared lists, not unrelated metadata
maps found alongside presets. Resolve inherits, includes, then child; detect missing
parents/cycles and preserve all five machine templates. Native Windows CLI expects
generated machine_full/process_full/filament_full directories; PrepareResources
generates supported0.4 profiles and machine_model definitions from the same release.
This is necessary for actual printer_model_id=N2S(A1)/N1(mini), which was initially
empty without those files. Trusted profile labels such as PA/PET need nested full
cache paths; guard traversal and create their parents. CLI prepends --outputdir to
--export-3mf unconditionally: pass only result.gcode.3mf as the export basename;
an absolute export filename initially produced native error-13 after slicing.

Real native-validation-9 and optional native installer test passed. Both output
archives contain selected plate G-code, matching native model IDs, actual material/
time metadata and non-empty printable layer extrusions. ParseGCode respects G90/91,
M82/83/G92, actual layer/feature markers and arcs; writes real per-layer preview JSON.
Model package sanitizes binary/ASCII STL, validates transforms against precise
profile polygon/exclusions/height, writes transformed standard3MF and guards archive
paths/entities/size. Tests cover an A1-only model rejected on mini, numeric safety,
actual extrusion modes, missing templates/cycles and checksum failures.

## Printer adapter source underway

Own Go implementation using paho.mqtt.golang1.5.1 and jlaffaye/ftp0.2.4. Third-party
MIT FTP source retained under third_party/ftp with a narrow PROT C option patch;
A1-family control is implicit TLS990, model-specific data defaults clear; protected
data remains selectable and must be hardware-checked. Pins per-device certificates,
never weakens global HTTP TLS. Private IPv4 DNS/IP checks constrain local connections;
PASV stays on the paired IP. DPAPI protects access codes; no raw commands exposed.
MQTT clean sessions/no auto replay, serial-specific reports, freshness/idle guards,
actual AMS lite tray data, identity checks/manual physical-screen fallback and
developer-mode evidence. Hexadecimal fun flag must be parsed as hex even if all
digits. A command ack/echo never fabricates physical state. BuildStart binds native
model/plate/checksum/task metadata, dynamically sizes actual filament slots and
uses no guessed external-spool sentinel (use_ams=false; mapping omitted). Mismatched
tray material blocks; color/preset differences require explicit review. Controls
require a live matching job. Protocol/dedup/persistence/HTTP integration still needed.

Go model/engine regressions and native installer passed (16.8s); printer pure
tests passed: busy/identity/model/tray guards, actual/sparse mappings, external
mode, ack-vs-status, hex command gating, LAN restrictions, device TLS pins and DPAPI.
No actual printer credentials, connection or physical print yet. Both devices
available, but user must enter credentials privately once wizard is ready and
explicitly press Print. Do not claim hardware or AMS tests passed.

## Repositories, environment and remaining work

Public org repo https://github.com/oneworkspacestudio/my-slicer created EMPTY;
Actions disabled immediately and confirmedfalse. Origin SSH registered. No source
commit/push or release yet. Branch codex/initial. Main app/backend HTTP and Product
UI are not implemented; cmd/my-slicer currently cannot compile for that reason.
Copied common AUI Shell/Login/Settings, ui adapter, main session/account guards and
CSS from my-tasks. npm install finished (39 packages), web uses Three.js0.180,
React19.3.1 and parser0.2.1. main.go fixed to webui.Files() function contract.
Go apphost pinned public Workspace v0.3.0-preview.41; go.mod uses local FTP replace.

Tools/cache environment: dot-source I:/Projects/tooling/four-apps/env.ps1, then
GIT_CONFIG_VALUE_0=I:/Projects/my-slicer; TEMP/TMP/GOTMPDIR named C task root. Go
I:/Projects/tooling/go1.27.2/go/bin/go.exe. Node/NPM C:/Program Files/nodejs.
Python C:/Users/fahim/anaconda3/python.exe; explicit UTF8 reads/writes. Upstream
source/reference sparse checkouts require exact git safe.directory overrides.
Native logs/receipts under I:/Projects/tooling/my-slicer; complete reviewed engine
at its engine/ folder. Debug PDB1.18GB is only external tooling, not app runtime.
Native installer tests clean their own temporary runtime. Named C build root still
exists; track all future preview PIDs and clean only exact task-owned paths.

Next: finish immutable/user-owned SQLite+ordinary-file project/slice/print API;
3MF preserve/re-slice (no discarded painting/modifiers/scripts); slice queue and
durable idempotent printer dispatch with ambiguous-start recovery; pairing/trust
wizard, real WebGL plate/layer preview/transforms/undo and Drive import/export with
shared design/icon/SSO. Then isolated native/browser and protocol acceptance,
matching AGPL/source/notices docs, release locally, Workspace42 default-app/catalog/
SSO/branding/README/build/update integration and release, authorized local install/
tunnel update and real hardware checks. Preserve owner runtime until authorized
maintenance, newest recovery and unrelated untracked histories. Do not run GitHub
Actions or mark frontend-only delivery complete.

## Superseding application milestone

The earlier "backend/Product not implemented" note is now superseded. Native
apphost API and browser Product are implemented and compile. SQLite stores only
metadata; ordinary source/project revision/native result files are retained. The
AUI screen and Three WebGL scene support STL placement/rotation/scale/duplicate/
arrange/reset/undo, native project selected-plate preservation, actual per-layer
paths/features/estimates, engine stages, printer discovery/trust/diagnostics,
external/actual AMS tray mapping, explicit print controls, unknown reconciliation,
and durable ordinary Drive copies with numbered conflicts/account checks.

Real HTTP/native acceptance passed for both models including exact native model,
plate, nonempty Gcode and full native project round trip. Found/fixed missing plate
membership in generated standard3MF: add explicit Metadata/model_settings.config
model_instance entries so native CLI export can be selected-plate imported later.
Device FTPS test passed real TLS control/passive LIST/STOR/SIZE/rename in PROTC and
PROTP modes. Found/fixed protected data handshake timing: lazy data TLS handshake
after transfer command, eager control handshake. Wrong code/incomplete size block
finalization. Concurrent print writes retain cancellation/progress/terminal states;
start intent is persisted before one publish; no replay on restart/unknown outcome.

Source tests/vet passed; govulncheck reported zero reachable/imported package
vulnerabilities (eight findings in required modules whose symbols are not used).
Native default release tests run with MYSLICER_TEST_ENGINE_ARCHIVE; test functions
skip real engine when unset. Browser fixture on localhost19020 used only synthetic
credentials/data, uploaded cube, scaled/duplicated/arranged and really sliced,
checked native layers and model-change invalidation. Small/desktop layout reviewed.
No physical printer paired or started: available devices still need private wizard
entry and explicit user Print. Hardware/AMS/clean-host/reboot/SCM/desktop modifier
comparison are explicitly unperformed in docs/ACCEPTANCE.md.

Source/build remains on I. Synthetic protected data C:/Projects/.cache/my-slicer-
20261019/preview, tracked preview PID in external tooling/my-slicer/preview.pid.
Stop/verify/clean only task fixtures after acceptance. Owner runtime untouched41.
Frontend uses immutable session guards and MyAccounts->Drive current-user check.
Standalone encrypted app backup inherits2GiB total cap; central Workspace backup
is needed for larger libraries; docs/UI disclose it. Only0.4nozzle and supportedPEI
plate enabled. Native3MF browser transforms unsupported and clearly disclosed;
retargeting multi-material native projects without per-slot profiles is blocked.

Workspace42 integration source now adds foundation session/env and seamless sign-in
registry, matching vector/ICO/PNG mark, README table/logo inventory, default20th
app build contract at9012 /slicer, checked org installer dependency on accounts+drive.
Still finish app source commit/build/release and public org component pin, then
verified full Workspace42 package, isolated managed SSO/Drive/upgrade tests, source
push/release, authorized owner upgrade and tunnel check. Do not certify hardware.

Release repetition exposed native Windows post-export cleanup retaining a process/
output handle despite a final success receipt and complete archive. Run now bounds
pipe cleanup and, after the actual return_code0 receipt plus archive and a3s grace,
ends only its own Windows job, then still validates all output metadata/toolpaths.
Do not infer completion from Gcode existence alone. Repeated real native HTTP
workflow regression is required before packaging. Added per-dispatch locks and
explicit pre-start upload cancellation; durable cancellation is checked under the
same lock as one start publish, so cancelling an upload cannot later start it.
Standalone main now reports only actual errors, avoiding inherited Report(nil).
