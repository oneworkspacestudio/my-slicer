# Model import geometry and loading

User supplied one Bambu production-extension 3MF (separate part model, saved build
and component transforms) and a 26,660-triangle binary STL. The old viewer skipped
all 3MF objects, despite successful import. The STL rendered correctly locally
(upload ~60 ms), but transfer/parsing had no plate feedback and main-thread decode.

Implementation: read-only bounded 3MF geometry follows package references, nested
transforms, measurement units and native plate instance membership. Source archive,
settings and slicing remain unchanged. Preview is generated lazily per plate and
cached atomically, so retained imports also work without re-upload. Invalid parts,
cycles, missing/empty plates and complexity limits report errors. Preview meshes
are capped at100k triangles; original full source is retained for slicing.

UI: immediate import response metadata starts rendering without waiting for all
refresh endpoints; background worker parses meshes/normals. Bounded per-account
viewer buffer cache, immediate STL bounds outline, loading and retry states in AUI,
automatic camera framing, stale response/worker cleanup, selected plate refresh.

Go tests/vet and frontend build pass. Synthetic tests cover referenced parts,
rotation/translation, plate instance membership, units, invalid faces/cycles and
HTTP cached preview/original-byte preservation. Both exact user files rendered in
isolated native app; phone layout and held/failing preview retry passed. Source
models were not copied into Git or publication assets. Native release/managed
acceptance and publication are still pending; do not claim delivery yet.

Artifacts: I:/Projects/tooling/slicer-import-20261022. Functional data only named
C:/Projects/.cache/my-slicer-import-20261022. Never use owner data as test fixture,
never send physical prints. Follow existing native/local release workflows and
immutable tags. Target app4 / Workspace67, preserve other24 app bytes/contracts.

## Final delivery complete (supersedes pending notes)

Public My Slicer4 source9c9ea4cc5433ec52e66156ee4e7074791aa3af91:
https://github.com/oneworkspacestudio/my-slicer/releases/tag/v0.1.0-preview.4
Public Workspace67 source5feafa8 (full source commit in evidence):
https://github.com/oneworkspacestudio/one-workspace/releases/tag/v0.3.0-preview.67
All published asset sizes/SHA256 match the tested local artifacts. Source archives
contain only the committed task changes; concurrent uncommitted installer work
was preserved and excluded. No Actions. User model source files/visuals were not
published. Immutable release tags remain at the code commits; documentation may
be committed later.

Both exact supplied car inputs rendered before slicing in standalone Windows and
through the real Workspace proxy/worker route. Go tests/vet, AUI/TypeScript, 3MF
component transforms/plate membership/unit/error tests pass. Delayed preview
feedback and retry without upload, phone layout and zero browser errors verified.
Actual verified native engine setup, 50-layer slice and native3MF roundtrip pass.
No physical printer commands or clean-host/reboot certification.

Fresh protected64-to67 managed fixture preserved existing project/model/printer
records, ordinary files, identities and other app versions. Real MyAccounts
sign-in, ordinary Drive copies and header Apps navigation passed. All test
services stopped; bounded named NTFS fixture/build cache removed with process,
containment and no-reparse checks. Original Downloads models unchanged.

Owner now67/Slicer4 through normal verified-backup updater run from I:. Identity,
paths, ports, selections, other24 app contracts/packages, Chat9, Mail17 and encrypted
Mail config preserved. Latest recovery manifest verified, all24 local/public
services healthy. Exact frontend and worker bytes match the installed local app; the real proxy
worker route passed isolated acceptance. Downloads installer67 matches public. Owner was never a functional fixture.

Evidence docs/evidence/slicer-import-20261022.json; durable artifacts on
I:/Projects/tooling/slicer-import-20261022. No delivery work remains. Do not rerun
owner updates or recreate fixtures just to check publication. Physical/hardware
acceptance remains a separate explicit human action.

Verification tooling: Python37 supports HTTPS here; use a browser User-Agent for
public health requests. Anonymous Workspace asset requests correctly require
sign-in; direct installed static bytes plus isolated real-proxy worker acceptance
verify delivery without using an owner session as a functional test.
