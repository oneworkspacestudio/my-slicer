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
must remain untouched. Report download browser test and release/install receipts
will be reconciled after completion; no delivery claimed yet.
