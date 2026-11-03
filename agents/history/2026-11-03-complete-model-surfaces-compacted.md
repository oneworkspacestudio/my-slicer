# Complete model surface previews

User supplied a detailed binary STL and screenshot of a speckled/incomplete model
on the build plate. Actual app Git root is the Slicer repo on the projects drive;
outer OneWorkspace folder is an aggregate, not a publication target. Base11dd546,
last release7. Current request is to fix previews; no new publication or owner
installation was performed. Keep that distinction explicit.

Cause: Mesh.Binary(previewLimit) kept every nth triangle above100k. This is face
deletion, not topology-preserving simplification, so closed surfaces developed
holes. Supplied file has342782 unique triangles; old stride4 showed85696. Older
history's description of preview "decimation" was misleading. No duplicate faces
were found in the supplied input; do not invent geometry repairs or smooth away
its real surface detail.

Fix: Binary() always preserves every facet. STL /mesh serves existing full,
centred mesh.stl and new imports no longer write a second sampled preview. Saved
STL preview count metadata is corrected in responses, with no DB migration or
source/print-data changes. 3MF previews regenerate lazily under full-v2 cache names
so prior partial caches are bypassed. Viewer requests a versioned full-v2 URL to
avoid old browser cache responses; private revalidation retained. Raw viewer
buffers now use a byte-bounded64MiB/eight-entry LRU, with transferable copies;
oversized meshes display without being retained. Existing1.5m source limits and
background worker decoding remain. No face-dropping fallback.

Verification: all Go tests and frontend build pass. Dense196608-face closed-cube
regression checks identical vertices and every edge still shared twice. App tests
cover a100800-face upload, retained legacy preview and metadata, original bytes,
and invalid old3MF caches. Native isolated Windows app/browser workflow passes:
actual supplied STL342782 facets reach GPU buffers, every returned vertex equals
the positioned source, original bytes unchanged, reload, top view and390px layout,
existing printer/draft/retry/failure-report workflows, zero browser errors and no
physical printer calls. Browser cache tests cover byte eviction, oversized entries,
LRU replacement and transferable-buffer isolation. Docs ACCEPTANCE updated.

Artifacts: tooling/slicer-surfaces-20261103 on projects drive; named NTFS cache
my-slicer-surfaces-20261103 only. MySlicer-preview-fix.exe is an unreleased
0.1.0-preview.8-dev candidate, not a managed installer/release. User-provided model
and screenshots stay in private task artifacts, never Git or release assets.
Functional tests use synthetic isolated accounts, not owner installation/data.
No real-engine or hardware slice/print performed: this changes only preview
serialization and caching; full slicing mesh and original file remain unchanged.

Final close-up reviewed: solid walls and all five compartments are visible; the
outer wall texture is actual input detail, not missing facets. Final four browser
tests pass, Go vet passes, test processes stopped and the named NTFS cache was
removed after resolved-path/process/no-reparse checks. Private receipts and fixed
desktop/top/mobile/close-up PNGs remain in task tools artifacts. Worktree changes
are intentionally uncommitted; publication/local owner installation still pending
current-task authorization under the updated aggregate AGENTS.md.

Resume: after authorization, follow current APP-LIFECYCLE/RELEASING and
immutable app/managed-channel paths, not the old Workspace74 owner bundle recipe.
Keep current owner apps/identity/recovery copies; don't claim an installed fix on
this or another PC based solely on the development candidate.

User subsequently authorized push, release and install. Preparing release8, with
version8 shipping candidate passing the same actual-model native/browser checks.
Current owner is Workspace113/Slicer7 with37 app contracts; Finance-only update
is active in another task. Do not overlap owner updates. Public113/source and
current public update channel replace the old74 distribution recipe. Work from
isolated current-main Workspace clone under tooling/slicer-release8-20261103 for
channel promotion and active managed-seed Slicer pin; keep113 immutable. App-only
publication uses its own workflow, so no new Workspace/Android release is needed.

## Completed managed delivery (2026-11-03)

Slicer8 is immutable and public from source5189d2a5fe39075b2504dfe146846b5f50971a60:
https://github.com/oneworkspacestudio/my-slicer/releases/tag/v0.1.0-preview.8
The exact shipping EXE passed the supplied342782-facet model test and every
positioned preview vertex matched. Source models/screenshots remain private.

Public Workspace114 source6593d1295a9291f5a396aaea5561e94a6f885b4c repairs
optional-guide handling for historical four-package Slicer registrations.
https://github.com/oneworkspacestudio/one-workspace/releases/tag/v0.3.0-preview.114
Ordinary clean build includes37 apps/451 packages and published Slicer8 pin;
mandatory publish-release.ps1 built permanent-key signed Android13 and verified
all immutable public assets. Actions remain disabled. Fresh114 full-eight-package
and legacy113 four-package controller114 then anonymous Slicer4→8 upgrades pass
real MyAccounts/session, saved projects, original bytes, full surface and unrelated
contracts. Historical catalog replay is synthetic only; no owner catalog edited.

The primary PC completed the normal anonymous selected workspace+slicer update
with native updater exit0. Public Windows manager and Slicer bytes, served frontend
and worker match tested release assets. Credentials/identity, paths/ports/selection,
autostart, unrelated application contracts and every registered unselected payload
were preserved; all36 configured services are healthy locally and publicly. Latest
verified recovery retained by supported cleanup. See docs/evidence/my-slicer for
sanitized release, fresh/legacy and installed receipts. No user file is published.

Other-PC installation remains pending; no remote access or physical printer/phone
acceptance is claimed. This fixes preview surface loss; the older other-PC slicing
failure is not certified resolved. Reopen/refresh saved models after updating.
Public release assets/tags stay immutable; later evidence commits do not retag.

The opt-in native public-controller check also selects Workspace114+Slicer8
against a synthetic four-file catalog and downloads/verifies selected payloads
anonymously. Shared channel advancement was reconciled by regenerating the
standard index and rebasing its promotion; no public tag/asset was changed.
