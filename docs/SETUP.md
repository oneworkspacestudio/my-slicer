# Set up an A1 or A1 mini

1. Install My Slicer through One Workspace, or run the standalone Windows package.
   Keep private application data on a local NTFS folder. Models and completed
   jobs also copy to `My Slicer/Models` and `My Slicer/Print jobs` in your Drive
   source while the browser is open; the app reports each copy's state.
2. The owner selects **Install verified engine**. Download, checksum verification,
   extraction and real validation are visible. Do not close the server while
   setup is running. Failed setup retains the previous engine. Cancel is available.
3. On each physical printer, inspect its firmware and enable LAN Only and
   Developer Mode where available. LAN Only disables Bambu Handy/cloud access.
   My Slicer does not update firmware or circumvent firmware command restrictions.
4. Select **Add printer**. Discovery is optional; otherwise enter the private IP,
   physical serial, model and LAN access code from that printer. Confirm the
   installed 0.4 mm nozzle and build plate. Save, inspect the two certificates,
   verify the physical IP/serial and explicitly trust that device.
5. Select **Test connections**. Diagnostics distinguish secure status/authentication,
   device identity, file listing and command-mode readiness. Reading status or
   listing storage does not prove a physical print. A new certificate needs a new
   explicit review. Codes are protected with Windows DPAPI for the hosting user.
6. Create a project; add STL or a complete self-contained Bambu project. Select
   the actual saved printer before slicing. Set quality/materials or retain the
   complete imported native settings. Inspect the real layers and native estimates.
7. Check the physical printer and choose **Review & print**. For an external spool,
   confirm the correct single filament is loaded. For AMS lite, choose each actual
   reported tray. Wrong material or missing trays block dispatch; color/preset
   differences require explicit review. Choose **Print now** yourself.

## If something fails

| Stage | What to check |
| --- | --- |
| Discovery is empty | Printer and PC on the same LAN; multicast may be filtered. Enter IP/serial manually. |
| Reachability or TLS | Printer powered on, correct private IPv4, LAN mode, local firewall. Do not publish printer ports. |
| Certificate changed | Recheck the physical device and IP, inspect fingerprints, trust only that device again. |
| Authentication | Re-enter the current LAN access code, not your cloud account password. |
| Identity unconfirmed | Check serial/model on the physical screen. If firmware does not advertise identity, use the explicit screen-confirmed fallback. |
| File listing / transfer | Working SD card, Wi-Fi and data-channel mode. A1 defaults to TLS control with clear passive data; enable protected data only if supported by that firmware. |
| Local commands unavailable | Enable firmware-supported Developer Mode; no cloud impersonation or proprietary plugin fallback is used. |
| Transfer failed | No start is sent. Check storage and connection, then review a new explicit attempt. |
| Start is uncertain | Check the physical screen/bed and refresh status. Never blindly start again. Resolve only after live idle status and your physical check. |
| Drive copy failed | Originals remain in My Slicer. Reconnect the same MyAccounts user and select Retry pending copies. Existing Drive files are numbered, not replaced. |

Changing model/nozzle/plate/material/geometry or the saved printer revision requires
a fresh slice. A server restart does not stop or relaunch a physical job. Restore
an encrypted backup into a new protected folder; printer codes restored under a
different Windows user must be entered again. Workspace owns remote HTTPS/tunnel
and sign-in configuration; My Slicer uses `/slicer/`, with no additional public port.

The inherited standalone encrypted backup format supports 2 GiB total app data,
including engine and generated previews. Use Workspace central backup for larger
libraries, or stop the server and preserve its complete protected data folder.
