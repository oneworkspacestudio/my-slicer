<div align="center">
<img src="docs/assets/logo.svg" alt="My Slicer" width="100" height="100">
<h1>My Slicer</h1>
<p>Prepare a model, inspect real layers and print on your own network.</p>
<a href="https://github.com/oneworkspacestudio/my-slicer/releases">Download for Windows</a> · <a href="docs/SETUP.md">Printer setup</a> · <a href="docs/ACCEPTANCE.md">Validation and limits</a>
</div>

Part of [One Workspace](https://github.com/oneworkspacestudio/one-workspace).
Native Windows app, no Docker. Open it from the Workspace app launcher and use
your existing MyAccounts session, or run `MySlicer.exe` independently.

- Import STL and self-contained Bambu 3MF projects, including projects with
  several colours. If a project was made for another printer, choose a matching
  material for each filament before slicing for your A1 or A1 mini.
- Move, rotate, scale, duplicate, arrange, reset and undo STL instances.
- Slice A1 and A1 mini plates using the actual Bambu Studio engine and complete
  release-matched inherited profiles, with a verified 0.4 mm nozzle configuration.
- Inspect real layer extrusion paths and native time/material estimates.
- Pair local printers with device-specific certificate trust and encrypted access codes.
- Explicitly upload and start a selected print; follow actual status and control
  its matching job. Uncertain starts are reconciled, never automatically retried.
- Use an external spool or map actual native filaments to reported AMS lite trays.
- Open from Drive; save originals and sliced archives as ordinary Drive files.

## Get started

Use Windows 10/11 x64 and an NTFS data folder. With One Workspace preview42 or
newer, My Slicer is installed by default unless you uncheck it in advanced setup.
Open **My Slicer** and choose **Add model**. No project setup is required.
Prepare the plate, choose **Slice plate**, then check the layers on **Preview**.
Install the slicing engine once when prompted; you can add models while it runs.
Use **Save project** only when you want a named project to reuse. Recent plates
are kept automatically. Add your printer from **Device** when you are ready.
The one-time official download is 474 MB; the installed runtime is about 570 MB,
plus your models and results. Progress shows bytes, files and real validation stages.

Standalone: extract the Windows ZIP and double-click `MySlicer.exe`, or choose
your own protected data path:

```powershell
.\MySlicer.exe --data D:\WorkspaceData\slicer
```

Follow the local first-owner link shown by the app. No default password is shipped.
See [setup](docs/SETUP.md) before connecting your printer. Printing requires you
to check the physical bed, nozzle, plate and filament and explicitly press Print.

## Studio workspace

Prepare and Preview use the familiar compact Studio layout: charcoal menus,
green navigation/actions, a printer/filament/process rail, full-height build plate,
model toolbar, dark slicing summary and layer/toolpath sliders. Native 3MF can
show every imported plate together. STL placement remains editable; native 3MF
keeps its original structure. Quality, Strength, Speed, Support and Others expose
validated native process settings. Advanced hides the less common controls.

Device shows saved printers, actual reported temperatures, filament trays and
tracked print progress. Camera, jogging, extrusion and remote hardware calibration
are explicitly unavailable through the supported connection. These panels never
substitute sample telemetry for a live printer. Calibration reports native engine
setup and explains the checks performed on the printer itself.

The layout follows Bambu Studio references while retaining My Slicer identity;
this app does not implement every Bambu Studio command or hardware integration.

## Project support

STL placement is editable in the browser. Native 3MF is passed intact to the engine
so plate membership, material assignments, modifiers and painting stay in the
source project. Browser editing of those native internals is not provided; choose
one plate and inspect its newly sliced toolpath. Incomplete settings, external
resources, scripts, multiple extruders and oversized projects are rejected with
guidance. Originals are never overwritten. See [exact limits](docs/ACCEPTANCE.md).

This is an **unsigned preview**. Real software slicing checks cover both printer
models; physical printing, AMS and clean-machine acceptance are reported separately.
It is not an official Bambu Lab product. No proprietary networking plugin is used.

## Build and source

Go 1.27.2, Node 22+ and npm are required for development. Run
`scripts/build.ps1 -Go <go.exe> -Npm <npm.cmd> -Release` from a clean source commit.
The script compiles AUI screens, checks TypeScript and Go, embeds the committed
Windows icon and creates the EXE, ZIP, dependency notices and checksum manifest.
See [engine provenance](docs/ENGINE.md) for the separately downloaded native engine.

AGPL-3.0; see [LICENSE](LICENSE), [NOTICE](NOTICE), and packaged dependency license
texts. The app provides links to corresponding application and engine source.

## See the screens

[Prepare](wireframe-images/prepare.png) · [Preview](wireframe-images/preview.png) ·
[Device](wireframe-images/printers.png) · [Printer setup](wireframe-images/printer-setup.png) ·
[Project](wireframe-images/projects.png) · [Activity](wireframe-images/activity.png) · [Calibration](wireframe-images/calibration.png).

[How to prepare a print](docs/USER-FLOW.md).

Multi-plate overview supports up to 9 imported plates together. Larger projects
use the imported plate selector to keep mesh memory bounded.
