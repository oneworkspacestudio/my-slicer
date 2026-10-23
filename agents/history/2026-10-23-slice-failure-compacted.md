# Slice failure recovery and cross-PC diagnosis

Follow-up to filament-import history: user reported generic native -100 failure
and undefined layers. Failure happened on another PC, so this owner installation
has no matching log. Printer/plate question remains unanswered. Do not claim the
other PC's exact fault is resolved or remotely installed without its evidence.
Prior user authorization includes push, release and local install. No Actions.

Changes: failed/pending Preview hides estimates, material badges, feature/layer
controls; failed/cancelled result has Return to Prepare and downloadable JSON
troubleshooting report. Native generic failure uses selected plate warning to
suggest empty-layer, bounds, conflict or memory recovery; never bypass native
checks, silently enable supports, invent output or start a physical print.
Report requires authenticated slice ownership, excludes model/project/G-code,
raw logs, device IDs and authentication details. Includes exact saved print
settings, app/engine versions, native code and plate warnings. User downloads
and chooses whether to share; nothing automatically sent. AUI sources/generated
TSX and Webframe/design/screens/Preview PNG/user guide updated. Version7.

Verification so far: Go tests/vet and frontend build pass. Exact candidate native
workflow passes desktop/390px failure UI and return action, both official engine
validations, normal slice/process/native roundtrip, supplied multi-material file
with original bytes and explicit persisted two-slot A1 choices, real75-layer
slice; no browser errors or printer hardware calls. Synthetic real native
empty-layer regression reproduces -100 and explicit supports recovers29layers.
Separate cross-target investigation on supplied file: A1 all4plates succeed;
A1mini plate1 succeeds75layers, plates2–4 return native empty/outside error.
That error is different from the user's screenshot. No automatic geometry change.

All functional tests use named NTFS cache, synthetic accounts and authorized
supplied model only, never owner data. Don't publish supplied model/geometry or
its screenshots. Native empty-layer test is opt-in with named fixture root.
Tool artifacts: slicer-failure-20261024 on task tools drive (name differs from
actual23October date), NTFS my-slicer-failure-20261024. User Bambu Studio windows
must remain untouched. Report download browser test passed with an actual JSON download and correct
filename. Ownership/privacy and real native failure/recovery tests passed.

Delivered immutable v0.1.0-preview.7, source 2b2040df80efe1d8120eb0b7abc77f4c9ccefc62,
main and branch pushed. All12 published asset hashes/sizes verified; exact tested
native candidate equals published and installed executable. Native current
Workspace74 publisher/SelectUpdates scope test passed against staged and actual
public metadata; only four owned Slicer package contracts selected.

Normal protected Slicer-only installer actually exited0 after consistent backup,
startup checks and rollback retention/cleanup. Workspace74, Social4 and all other
app contracts/bytes, owner identity, paths, ports, selection and autostart retained.
All25 local/public services healthy; installed frontend and worker hashes match.
Latest verified rollback retained, stopped named NTFS fixture cache cleaned with
resolved path/reparse/process checks. Sanitized delivery proof is committed under
docs/evidence/slice-failure-20261023.json; task tool receipts retain operational
details. No supplied model source/images published, no printer hardware contacted.

Remaining user issue: the failure occurred on another PC. It has not been remotely
updated or its exact failed configuration reproduced. User should update My Slicer
there through Workspace Check for updates, retry, then attach Download
troubleshooting report if still failing. Read the actual selected settings/native
warning before making another slicing change. Do not equate the verified synthetic
empty-layer recovery or local supplied-file successes with resolving that PC.
No local publication/installation work remains for version7.
