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

Next: commit source, package6 and compare exact EXE with native candidate. Publish
immutable6 with all asset hashes/source/licences and four-package update manifest
validated by the current native Workspace distribution code. Then normal protected
Slicer-only update, preserving current manager, other apps, identities and data.
Wait actual updater exit/cleanup; verify installed hashes, health and retained
rollback, clean only stopped named task caches and record final delivery.
