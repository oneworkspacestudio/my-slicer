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
