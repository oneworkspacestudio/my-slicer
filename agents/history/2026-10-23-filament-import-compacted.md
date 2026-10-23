# Expanded filament settings and explicit retargeting

User reported an import failure on a supplied P2S 3MF and authorized fix, push,
release and local install. Base99b31a6 / installed Slicer5, Workspace74, Social4.
Branch codex/import-filament-settings. No GitHub Actions or physical printing.

The file has two filament IDs but four nonempty nozzle-temperature entries.
InspectProject incorrectly treated any vector length other than two as missing.
Accept complete expanded vectors without truncating or changing source bytes;
genuinely short or empty settings still report plain-language recovery guidance.

There was a second blocker: saveProject rejected every multi-material retarget.
Allow explicit slot-for-slot selections, keep original layout/painting/source,
and import unsupported printer projects with empty material selections and a
source-printer notice. Save keeps the slot count; slice requires a compatible
named profile for every slot. Never collapse a multi-material project into one
material. Native CLI uses official target machine/process and selected filament
profiles; physical print review/tray mapping remain required and unchanged.

AUI PrinterRail has per-slot selectors, original profile labels, unavailable
preserve-settings choice for an incompatible printer, and material/profile
readiness gates. Profile responses consult current form values to avoid racing
project reload. Backend tests check expanded/incomplete vectors, original bytes,
empty retarget slots, rejection of one-slot fallback and persisted two-slot choices.
Webframe Prepare/Preview HTML, visible PNGs, design/screens and user docs updated.

Go tests/vet and AUI/TypeScript build passed. Local exact native candidate passes
existing workflow, both official engine validations, real3MF roundtrip, actual
reported model import with four plates visible, byte-identical source retrieval,
two explicit A1 material choices, reload and390px layout, and real75-layer A1
slice with two filament slots. Browser errors0; no hardware contacted/print sent.
Accounts/data are isolated fixtures; one user-provided model is an authorized
test input, not a synthetic model. Never publish its source, geometry or images.

Artifacts: tooling/slicer-filament-20261023 on the task tools drive; named NTFS
my-slicer-filament-20261023 cache only. Source/data must not use owner installation
as a functional fixture. Existing Bambu Studio windows belong to the user; do not
stop them or confuse them with the headless engine under the test cache.

Delivery complete: source a60092534881918113d1d91b410a0717204831b4 pushed main,
branch and immutable v0.1.0-preview.6. All12 public asset sizes/SHA256 verified,
including source ZIP and complete notices. Exact tested/published/installed EXE:
13d61f97212a959928d33bb9f86b11a90ce5433ba64b902fd3e4228b8453b0b2.
Native Workspace74 publisher generated the four-owned-package manifest; uncached
validation against actual public asset metadata and SelectUpdates scope passed.

Normal protected Slicer-only offline update used the checksum-verified public74
bundle/setup with only Slicer payloads changed. Full consistent backup verified,
native updater actually exited0, latest verified recovery copy retained. Kept
Workspace74, Social4 actual bytes, all non-Slicer app contracts/package hashes,
identity, paths, ports, selection and autostart. All25 local/public managed
services healthy. Installed frontend/mesh-worker hashes match the tested build.
No owner-installation functional fixture or physical print. Stopped isolated test
installations/named NTFS cache removed after path/reparse/process checks; user's
separate Bambu Studio windows were left alone. Public evidence is sanitized and
contains no supplied model source or screenshots. No delivery work remains.
User should refresh My Slicer, reopen the file and choose a matching material for
each numbered filament when retargeting to A1/A1 mini.
