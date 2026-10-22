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
