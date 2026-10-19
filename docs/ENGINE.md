# Native engine and provenance

Pinned official Bambu Studio **v02.08.02.61**, source commit
`926a7192574bcb9b3a732e1ec59a46d79cb45466`.

- [Official Windows artifact and release](https://github.com/bambulab/BambuStudio/releases/tag/v02.08.02.61)
- [Matching source](https://github.com/bambulab/BambuStudio/tree/926a7192574bcb9b3a732e1ec59a46d79cb45466)
- [Official CLI documentation](https://github.com/bambulab/BambuStudio/wiki/Command-Line-Usage)

Archive: `Bambu_Studio_win-v02.08.02.61-20260820225108.zip`, 473534359 bytes,
SHA256 `3faaa194a90a2e7defa39a6a63f5a654b8cdb762c6a98556cf771fba582229f5`.
The compiled [runtime manifest](../internal/engine/runtime-manifest.json) lists
6753 required files (570212481 bytes). Debug PDBs, development headers and the
separate proprietary networking plugin are excluded. The slicing executable and
its required DLL/resources are unmodified. License files remain in the runtime.

The installer verifies the full official archive before extraction and every
runtime file against the compiled size/hash list. It resolves all inherited and
included settings against that release, prepares the CLI's full preset cache,
and actually slices both model profiles before atomic activation. Failed promotion
restores the previous version; previous engines are retained for recovery.
Only one native slice worker runs, in a kill-on-close Windows job with a 4 GiB
memory limit, process-count bound, private working directory and timeout. This is
resource containment, not an independently audited malicious-native-input sandbox.

Server-only `MYSLICER_ENGINE_ARCHIVE` can point to an offline copy of the exact
official archive. Its checksum and file verification are still mandatory. It is
never accepted as an arbitrary browser-supplied path.

The own LAN adapter uses documented/source-inspected MQTT and FTPS behavior.
`third_party/ftp` contains jlaffaye/ftp v0.2.4 under MIT with a small documented
PROT C option. Paho MQTT is under its original EPL/EDL terms. No proprietary
Bambu networking component or cloud impersonation is included.
