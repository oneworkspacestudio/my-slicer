# Flow and state ownership

```mermaid
flowchart LR
  prepare["Prepare: add model"] -->|"Slice plate"| preview["Preview: inspect actual layers"]
  preview -->|"Back to Prepare / re-slice"| prepare
  preview -->|"Pair a printer"| printers["Printers"]
  printers -->|"Add / edit printer"| setup["Printer setup: details → trust → check"]
  setup -->|"Done"| printers
  printers -->|"Back to Prepare"| prepare
  prepare -->|"Save / open project"| projects["Saved projects and recent plates"]
  projects -->|"Open plate"| prepare
  prepare -->|"Activity"| activity["Print jobs, slices and Drive copies"]
  activity -->|"View slice"| preview
  preview -->|"Explicit review + Print now"| activity
```

```mermaid
flowchart TB
  aui["AUI screen trees + semantic adapter"] --> product["Product controller: routing, selection, fields"]
  product --> api["Authenticated APIs: revision / ownership checks"]
  product --> scene["Three.js plate and real extrusion preview"]
  api --> store["SQLite metadata + original ordinary files"]
  api --> engine["Bounded verified native slicing worker"]
  api --> adapter["Private LAN printer adapter + encrypted access codes"]
  product --> drive["Existing account-checked Drive integration"]
```

Routing belongs to the URL hash and preserves the current plate across screens.
The server owns projects/drafts, models, immutable slices, printer records and
dispatch state. No user action must explicitly create a project to import a model.
Draft status is an additive JSON field; absent means an existing saved project.
The browser owns only current controls, unsaved form fields, selection and dialog
state. Access codes never enter local/session storage. Leaving printer setup clears
them. Missing/unknown connection state stays distinct from connected/idle.


```mermaid
flowchart LR
 Prepare["Prepare: models + process rail"] -->|Slice plate| Preview["Preview: real layers + result"]
 Preview -->|Print plate: human review| Confirm["Confirm physical print"]
 Prepare --> Device["Device: saved printers + real telemetry"]
 Device --> Setup["Printer setup: details, trust, test"]
 Prepare --> Project["Project: saved projects and activity"]
 Prepare --> Calibration["Calibration: engine checks and printer guidance"]
```

Top navigation owns only route changes; project/revision and profile selection stay
in Product. AUI owns UI trees. Scene owns disposable WebGL resources and bounded mesh
cache. Server validates process overrides against a fixed native-key schema; no
scripts, commands or arbitrary slicer keys. Device uses existing pairing/session
authorization and real report fields. Printing requires existing explicit review.
