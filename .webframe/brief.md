# My Slicer: start with the model

Audience: A1 and A1 mini owners, including people unfamiliar with slicing tools.
The core job is add a model, set up its plate, slice, inspect and explicitly print.

## Requirements

- Add model is a primary action at the top of Prepare, visible at 390 × 844 and
  1366 × 768. Open from Drive is adjacent when Workspace is connected.
- A blank plate works immediately. An internal recoverable plate record is created
  only on import; naming/saving a project is an optional later action. Resume the
  latest unsaved plate after reload. Existing named projects stay available.
- Prepare, Preview and Printers have distinct screens. Saved projects and Activity
  are secondary navigation. Avoid a long dashboard of unrelated forms.
- Printer setup is a normal page with Details → Trust → Connection check. It works
  before importing a model or creating a project. No native modal requirement.
- Show straightforward recovery for unavailable engine, invalid model, offline
  printer and failed connection checks. Keep entered printer details after errors;
  clear access codes when leaving setup. Never display or persist codes in browser
  storage. Never start a print while adding, trusting or testing a printer.
- Keep originals, immutable slice revisions, ownership, device certificate review,
  explicit physical print confirmation, Drive ordinary files and engine provenance.

## Assumptions and limits

Existing backend project records provide durable plates; a backwards-compatible
draft flag distinguishes unsaved plates from named projects. No models or printer
records are deleted by this redesign. Software fixtures are not hardware acceptance.
The shared blue branding, light/dark mode and accessible keyboard flow stay intact.

The requested crash has not reproduced on a fresh supported Chrome installation.
Test repeated setup, saved printer validation, embedded launch and missing native
dialog support, and report that distinction honestly.
