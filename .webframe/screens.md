# Screens and states

## prepare — Prepare / #

Entry: signed in, or return from preview/projects. Show Add model and Open from
Drive first, optional Save project and New plate secondary. Empty state explains
STL/3MF, with a visible plate rather than a project creation gate. Imported state
shows the actual STL or read-only 3MF geometry, automatically framed around the
model. Upload and preview loading messages stay beside the plate; a wire outline
marks STL bounds during transfer. Failure offers Retry model preview without a
new upload. Mesh decoding runs off the UI thread. Existing 3MF imports receive
read-only previews on demand. Selected native plate filters saved build instances.
Imported state also shows object list and compact arrange/duplicate/remove/undo actions. Numeric
transforms and advanced slice settings expand on demand. The settings rail shows
model/printer, material and quality, with Slice plate as its primary action.
Show one named material selector for every imported filament slot. Projects from
another printer open with all source colours and layout retained, a clear source
printer notice and empty material selections. Slice stays disabled until matching
profiles are loaded and every slot is chosen. Retargeting never collapses slots.
Engine missing/downloading/failed states appear in the rail without pushing import
below the plate. Import failure keeps the current plate. Slicing navigates to
Preview. Save project opens the optional naming state; New plate preserves previous
work in Recent plates rather than deleting it.

## preview — Preview / #preview

Entry: Slice plate or Preview navigation. Queued/processing shows actual status,
elapsed time and Cancel. Ready shows actual paths on the plate, layer/feature
controls, time/material estimates, Download and Review print. Changed settings or
stale revisions explain return to Prepare and re-slice; missing printer links to
Printers. Empty/error/cancelled shows a clear Prepare action. Print review opens a
confirmation state with physical checks, actual tray mapping and explicit Print
now; cancelling preserves the slice. Never fabricate estimates.

## printers — Printers / #printers

Entry: navigation or preview's missing target. Add printer is always visible.
Empty state explains pairing is optional for slicing. Saved devices show actual
connection/unknown status, model, firmware and Setup & diagnostics. Choose a saved
printer on Prepare, not by silently changing the active plate here. Offline state
offers setup/connection checks. Add/edit goes to Printer setup, Back goes to Prepare.

## printer-setup — Add printer / #printer-new or #printer/<id>

Details: plain guidance for obtaining local connection details, optional discovery,
name/model/IP/serial/access code/plate and physical checks. Discovery error stays
on this page and manual entry remains available. Save and continue validates and
persists encrypted details; failure retains entered fields with inline message.
Trust: explain fingerprints, Inspect certificates, explicit confirmation and Trust
this device. Never silently accept a device certificate. Failed inspection offers
retry and Edit details. Check: Test connection shows actual diagnostics and status;
offline stays actionable. Done returns to Printers; Back to Prepare always exists.
Escape/browser Back should not crash the app. No discovery or connection request
occurs merely on opening this screen; automated tests do not contact hardware.

## projects — Saved projects / #projects

Entry: secondary navigation. Separate named projects from Recent plates (unsaved,
automatically recoverable work). Empty state links straight to Add a model on
Prepare. Open restores a plate and resets selection/slice state. Save current plate
is optional and requires a name; cancel preserves the draft. Existing legacy
projects are named projects, with no migration loss or forced renaming.

## activity — Activity / #activity

Entry: secondary navigation. Show actual print jobs, matching-job controls,
uncertain-start reconciliation, recent slices for the current plate and Drive
copy results/retry. Empty states explain where results appear. Viewing a slice
opens Preview, deleting a slice removes its result only. Cancel print requires
human confirmation. No automatic retries of an uncertain physical start.

## Shared states

Loading and busy state provide feedback without shifting the primary import bar.
Dialogs for optional naming, Drive selection and explicit print confirmation have
close/escape/focus handling and a fallback when native modal support is missing.
A rendering failure offers Reload slicer and Return to Prepare, preserving server
data rather than showing a blank page. Existing Settings & backup remain reachable.


## Studio revision of shared layout
Prepare/Preview share a left Printer, Nozzle, Project Filaments and Process rail.
Process has Quality, Strength, Speed, Support and Others sections; Advanced reveals
additional validated settings. Global/Objects switches settings/object selection.
Prepare toolbar offers Add model, New plate, arrange, move/rotate/scale, duplicate,
delete and undo. Unsupported modeling tools remain disabled with explanations.
Native multiple plates have a real numbered plate picker and overview; selecting
a plate preserves the original archive and only changes the active native plate.
Preview retains the rail, shows actual toolpaths and slicing result with vertical
layer range and horizontal move visibility control. Empty/failed/queued results
stay actionable. Print plate always opens the existing explicit review.
Device replaces the visible Printers tab label but retains #printers compatibility.
Its left saved-device/status menu, camera area, printing progress, telemetry and
filament trays reflect real API state. No paired device shows dashes and setup;
camera/motion are unavailable without implemented capability. No automatic probe.
Project links existing projects and Activity. Calibration explains native validation
and manual device calibration; unavailable hardware actions cannot be executed.
