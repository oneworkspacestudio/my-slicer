# Prepare a print

1. Open **Prepare** and choose **Add model**, or **Open from Drive** in Workspace.
   Choose an STL or Bambu 3MF. There is no project form to complete first.
   The model appears on the plate before slicing. While it loads, a message shows
   what is happening. If the preview fails, choose **Retry model preview**.
   Imported 3MF layouts are shown as saved, and the original file is kept.
   If it was made for another printer, choose your A1 or A1 mini and a material
   for each numbered filament. Each colour keeps its own slot. Slice is enabled
   once every material has been selected. Compatible projects can keep their
   original print settings.
2. Choose your printer model, material and quality. Add or arrange the objects on
   the plate. Use the move/rotate/scale toolbar button for transforms. Process settings are
   grouped under Quality, Strength, Speed, Support and Others in the left rail.
   If prompted, install the slicing engine once; you can add models while it runs.
3. Choose **Slice plate**. On **Preview**, inspect the actual layers, time and
   material estimates. The right panel shows native estimates; the vertical slider selects a layer
   and the bottom slider reveals actual path segments. Up to 50 layers are shown
   together, with a translucent model for context. Export the sliced 3MF here.
4. For direct printing, add a printer from **Device**. Enter the details from its
   screen, review and trust its certificates, then test the connection. Select that
   printer on Prepare and slice for it. Review the print, check the physical printer
   and loaded filament, then explicitly choose **Print now**.

Adding a printer, trusting it and testing its connection never start a print.
Connection problems stay on the setup screen so you can correct the details.
Leaving setup clears the access code from the form; the saved code stays encrypted.

## Projects are optional

Your work is kept in **Recent plates** after you add a model. Reopening the slicer
restores the latest unnamed plate. **New plate** starts a blank plate and keeps the
previous work available. Existing saved projects are kept too.

Choose **File → Save project** when you want a name for a reusable plate. Open it later
from **Project**. Importing a native 3MF opens its own plate when necessary
so its layout, modifiers and painting are not mixed with another plate.

**File → Print history** contains print jobs, previous slices and the status of Drive copies.
Return to Prepare and slice again after changing a model, print settings or target
printer. Printing always needs a current slice and your explicit confirmation.


## The screens

Use **Apps** in the header to switch Workspace apps. **Device** contains saved
printers and real reported telemetry. Camera and motion controls clearly show
when the connection does not support them; no print progress is invented.

Use **Calibration** to check the native engine and read printer calibration
guidance. Physical calibration is performed on the printer’s screen.

[Prepare](../wireframe-images/prepare.png) ·
[Preview](../wireframe-images/preview.png) ·
[Device](../wireframe-images/printers.png) ·
[Calibration](../wireframe-images/calibration.png).

Multi-plate overview supports up to 9 imported plates together. Larger projects
use the imported plate selector to keep mesh memory bounded.
