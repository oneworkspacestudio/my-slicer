# Acceptance scope — initial Windows preview

## Completed software checks

- Actual checksum-verified official engine install, both A1 and A1 mini profile
  validations, runtime restart verification and no debug-PDB redistribution.
- Actual native HTTP workflow for both models: STL import, immutable settings,
  precise target geometry, real G-code 3MF, nonempty extrusion layers, native
  model IDs, plate/nozzle settings and time/material estimates.
- Native generated project round trip through original 3MF import and re-slice,
  retaining complete native settings and plate membership. The test caught and
  fixed missing plate membership in generated projects.
- Account ownership/CSRF guards, stale revision rejection, invalid upload cleanup,
  resource and archive bounds, profile include/inheritance/cycle/missing-template
  cases, A1-only geometry rejection on mini, actual G-code modes and AMS mapping.
- Real local TLS/FTPS fixture: authentication, LIST, passive transfer, size check,
  rename and both clear/protected data modes; wrong code and incomplete size fail.
  This is a software protocol fixture, not a physical printer result.
- Browser workflow on the native app: sign-in, file chooser, transforms,
  duplication, arrange, actual slice, layer/features and target revision changes.

## Explicit limits

Initial physical configuration support is A1/A1 mini with a confirmed 0.4 mm
nozzle and Textured PEI or Smooth PEI/High Temp plate. Native projects are kept
intact and passed to the engine; browser transforms/duplicate/arrange apply to
STL only. One selected native plate is sliced. Post-processing scripts, external
resources, multiple extruders and incomplete material settings are rejected.
No broad claim is made for every desktop painting/modifier feature or third-party
3MF dialect. Check the resulting toolpath before printing.

100 MiB uploads; 1.5 million STL triangles; 100 instances/plate; 100 saved projects
and recent plates combined;
2 GiB source uploads/account; five queued/active slices/account and 50 retained
results. Browser meshes are decimated above 100,000 triangles, with full geometry
used by the engine. Layer preview has explicit segment/range limits; complexity
overflow is reported rather than replaced by simulated toolpaths.

## Required physical and deployment checks

These are **unperformed unless a later release receipt explicitly records them**:

- Pair each actual A1 and A1 mini; verify serial, firmware, secure status and FTPS.
- User explicitly starts a small print on each; verify upload, acknowledgement,
  preparation, first layer and matching progress, then pause/resume/cancel.
- Actual AMS lite trays and material changes where the hardware is installed.
- Independent desktop Bambu Studio comparison using matching complete settings,
  representative painted/modifier/multi-plate desktop project fixtures.
- Clean second Windows machine, actual reboot/SCM service recovery and independent
  security review. This unsigned preview is not a certification of those checks.

No automated test starts a physical printer. Unknown start outcomes never cause
an automatic resend. See the Workspace release receipt for managed SSO/Drive,
installer/update and remote route validation, reported separately from hardware.


## Studio UI revision — preview.5

The browser shell and Prepare/Preview/Device panels follow the supplied Studio
references. This is a browser implementation with My Slicer identity, not a claim
of full Bambu Studio feature parity. Unsupported camera, jogging, extrusion and
physical calibration commands stay unavailable. Current A1/A1 mini 0.4 mm profile
limits, pairing/trust checks and explicit physical print confirmation remain.

New process overrides use an explicit server whitelist. Unknown native keys,
G-code/script fields, invalid enum values, non-finite values and out-of-range or
fractional layer counts are rejected before saving. A failed save retains the
previous usable project. Existing projects without the additive process map load
unchanged. Preserved native 3MF settings take precedence over the process rail.

Preview requests at most 50 actual layers and falls back to a single layer when
the existing 250,000-segment response limit is reached. The translucent mesh is
context, not a fabricated toolpath. Native multi-plate overview fetches real meshes
per imported plate; exported/sliced source archives remain intact.

Multi-plate overview supports up to 9 imported plates together. Larger projects
use the imported plate selector to keep mesh memory bounded.


Final Windows candidate acceptance passed on2026-10-23: Go tests/vet, generated
AUI/TypeScript build, native plate-origin regression and the full opt-in browser
workflow. A verified pinned engine performed its A1/A1 mini setup validation,
then real slicing, a persisted outer-wall speed override slice and native3MF
round trip. Desktop/390px phone layout, real sliders, Device/Calibration routes,
printer validation/trust page, draft recovery and delayed/failed mesh retry passed.
Browser errors:0. No owner data fixture, hardware connection or physical print.

[Native Prepare](screenshots/slicer5-prepare-desktop.png) ·
[Native Preview](screenshots/slicer5-preview-desktop.png) ·
[Native Device](screenshots/slicer5-device-desktop.png) ·
[Native Device on phone](screenshots/slicer5-device-phone.png).
