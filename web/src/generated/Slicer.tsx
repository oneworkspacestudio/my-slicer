// Generated from screens/Slicer.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Grid, Heading, Input, Link, Metric, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Slicer({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row gap="md">
        <Stack gap="xs">
          <Heading level={1}>Prepare your next print</Heading>
          <Text>Your models, your printers. Powered by the native Bambu Studio engine.</Text>
        </Stack>
        <Spacer />
        <Button variant="primary" onClick={() => onAction("newProject")}>New project</Button>
      </Row>
      <Row gap="md">
        <Link href={data.sourceURL} external={true}>Application source · AGPL-3.0</Link>
        <Link href={data.engineSource} external={true}>Native engine source</Link>
      </Row>
      {data.engineNeeded && (
        <Card>
          <Heading level={2}>Set up the slicing engine</Heading>
          <Text>{data.engine.message}</Text>
          <Text>{data.engineProgress}</Text>
          <Section id="engine-progress" />
          <Text>One-time download: 474 MB. Installed runtime: about 570 MB. Both printer profiles are sliced and checked before activation.</Text>
          {data.owner ? (
            <>
              <Button variant="primary" onClick={() => onAction("installEngine")} disabled={data.engine.busy}>Install verified engine</Button>
              {data.engine.busy && (
                <Button onClick={() => onAction("cancelEngine")} cancel={true}>Cancel setup</Button>
              )}
            </>
          ) : (
            <Text>Ask the Workspace owner to finish engine setup.</Text>
          )}
        </Card>
      )}
      <Row gap="md">
        <Field label="Project"><Select name="projectID" options={data.projectOptions} /></Field>
        <Field label="Project name"><Input name="projectName" /></Field>
        <Button onClick={() => onAction("renameProject")}>Rename</Button>
      </Row>
      {data.hasProject && (
        <Grid columns={2} variant="slicer">
          <Stack gap="md">
            <Card>
              <Row>
                <Heading level={2}>Build plate</Heading>
                <Spacer />
                <Badge>{data.revisionText}</Badge>
              </Row>
              <Section id="scene" />
              <Text variant="caption">{data.sceneNotice}</Text>
              <Row gap="sm">
                <Field label="Add STL or Bambu 3MF"><Input type="file" name="modelFile" accept=".stl,.3mf" /></Field>
                <Button onClick={() => onAction("upload")} disabled={data.noFile}>Add model</Button>
                {data.workspace && (
                  <Button onClick={() => onAction("openDrive")}>Open from Drive</Button>
                )}
              </Row>
              {data.importNotice && (
                <Alert>{data.importNotice}</Alert>
              )}
              <Field label="Object"><Select name="objectID" options={data.objectOptions} /></Field>
              {data.editable && (
                <>
                  <Grid columns={3}>
                    <Field label="X (mm)"><Input name="x" type="number" step={1} /></Field>
                    <Field label="Y (mm)"><Input name="y" type="number" step={1} /></Field>
                    <Field label="Z (mm)"><Input name="z" type="number" step={1} /></Field>
                    <Field label="Rotate X (°)"><Input name="rx" type="number" step={5} /></Field>
                    <Field label="Rotate Y (°)"><Input name="ry" type="number" step={5} /></Field>
                    <Field label="Rotate Z (°)"><Input name="rz" type="number" step={5} /></Field>
                    <Field label="Scale (%)"><Input name="scale" type="number" min={1} max={10000} /></Field>
                  </Grid>
                  <Row gap="sm">
                    <Button onClick={() => onAction("transform")}>Apply transform</Button>
                    <Button onClick={() => onAction("duplicate")}>Duplicate</Button>
                    <Button onClick={() => onAction("reset")}>Reset</Button>
                    <Button onClick={() => onAction("remove")}>Remove</Button>
                    <Button onClick={() => onAction("arrange")}>Arrange</Button>
                    <Button onClick={() => onAction("undo")} disabled={data.noUndo}>Undo</Button>
                  </Row>
                </>
              )}
            </Card>
            {data.activeSlice && (
              <Card>
                <Heading level={2}>Native slice</Heading>
                <Text>{data.activeSlice.message}</Text>
                <Text>{data.sliceElapsed}</Text>
                {data.slicing && (
                  <Button onClick={() => onAction("cancelSlice")}>Cancel slicing</Button>
                )}
                {data.sliceReady && (
                  <>
                    <Row gap="md">
                      <Metric value={data.timeText} />
                      <Text>{data.filamentText}</Text>
                      <Badge>{data.layerCount}</Badge>
                    </Row>
                    <Field label="Preview layer"><Input name="layer" type="range" min={1} max={data.maxLayer} step={1} /></Field>
                    <Text>{data.layerText}</Text>
                    <Field label="Toolpath feature"><Select name="feature" options={data.featureOptions} /></Field>
                    <Row>
                      <Button onClick={() => onAction("downloadSlice")}>Download sliced 3MF</Button>
                      <Button variant="primary" onClick={() => onAction("reviewPrint")} disabled={data.staleSlice}>Review & print</Button>
                    </Row>
                    {data.staleSlice && (
                      <Alert tone="warning">This slice is from an older revision or has no paired target. Select a printer and slice again.</Alert>
                    )}
                  </>
                )}
              </Card>
            )}
          </Stack>
          <Stack gap="md">
            <Card>
              <Heading level={2}>Print settings</Heading>
              <Field label="Target printer"><Select name="target" options={data.printerOptions} /></Field>
              <Field label="Printer model"><Select name="model" options={data.modelOptions} disabled={data.hasTarget} /></Field>
              <Field label="Nozzle"><Select name="nozzle" options={data.nozzleOptions} disabled={true} /></Field>
              <Text variant="caption">Only the verified 0.4 mm profiles are currently enabled. Confirm the actual nozzle and plate in printer setup.</Text>
              <Field label="Build plate"><Select name="plate" options={data.plateOptions} disabled={data.hasTarget} /></Field>
              {data.nativeProject && (
                <>
                  <Field label="Project plate"><Select name="nativePlate" options={data.nativePlates} /></Field>
                  <Checkbox name="preserveImported" label="Preserve native 3MF settings, modifiers and painting" />
                </>
              )}
              <Field label="Quality profile"><Select name="quality" options={data.processOptions} disabled={data.preserving} /></Field>
              <Field label="Filament profile (STL / single-material override)"><Select name="filament" options={data.filamentOptions} disabled={data.preserving} /></Field>
              <Grid columns={2}>
                <Field label="Layer height (mm)"><Input name="layerHeight" type="number" step={0.01} min={0.08} max={0.28} disabled={data.preserving} /></Field>
                <Field label="Infill (%)"><Input name="infill" type="number" min={0} max={100} disabled={data.preserving} /></Field>
                <Field label="Walls"><Input name="walls" type="number" min={1} max={20} disabled={data.preserving} /></Field>
                <Field label="Brim"><Select name="brim" options={data.brimOptions} disabled={data.preserving} /></Field>
              </Grid>
              <Checkbox name="supports" label="Generate supports" disabled={data.preserving} />
              <Row>
                <Button onClick={() => onAction("saveSettings")}>Save settings</Button>
                <Button variant="primary" onClick={() => onAction("slice")} disabled={data.cannotSlice}>Slice plate</Button>
              </Row>
              <Text variant="caption">Changes create a new revision. Slicing uses immutable models and settings; an old result cannot start a new revision.</Text>
            </Card>
            <Card>
              <Row>
                <Heading level={2}>Your printers</Heading>
                <Spacer />
                <Button onClick={() => onAction("addPrinter")}>Add printer</Button>
              </Row>
              {data.printers.map((item, i) => (
                <Stack key={i} gap="xs">
                  <Row>
                    <Text>{item.config.name}</Text>
                    <Badge>{item.status.status}</Badge>
                    <Button item={item.config.id} onClick={() => onAction("editPrinter")}>Setup & diagnostics</Button>
                  </Row>
                  <Text variant="caption">{item.detail}</Text>
                  {item.diagnostics.map((item, i) => (
                    <Text key={i}>{item.message}</Text>
                  ))}
                </Stack>
              ))}
              {data.noPrinters && (
                <Text>Add your A1 or A1 mini. Upload and start stay local to your network.</Text>
              )}
            </Card>
            {data.prints.map((item, i) => (
              <Card key={i}>
                <Row>
                  <Heading level={3}>{item.state}</Heading>
                  <Badge>{item.printerName}</Badge>
                </Row>
                <Text>{item.message}</Text>
                <Text>{item.progressText}</Text>
                <Row gap="sm">
                  <Button item={item.id} onClick={() => onAction("pause")}>Pause</Button>
                  <Button item={item.id} onClick={() => onAction("resume")}>Resume</Button>
                  <Button item={item.id} onClick={() => onAction("cancelPrint")}>Cancel print…</Button>
                  <Button item={item.id} onClick={() => onAction("reconcile")}>Refresh status</Button>
                  {item.unknown && (
                    <Button item={item.id} onClick={() => onAction("resolve")}>Resolve uncertain start…</Button>
                  )}
                </Row>
              </Card>
            ))}
            {data.workspace && (
              <Card>
                <Heading level={2}>Drive copies</Heading>
                <Text>Originals and completed sliced archives are saved as ordinary files in Drive while this app is open.</Text>
                {data.driveSaves.map((item, i) => (
                  <Row key={i}>
                    <Text>{item.name}</Text>
                    <Badge>{item.state}</Badge>
                    <Text>{item.message}</Text>
                  </Row>
                ))}
                <Button onClick={() => onAction("retryDrive")}>Retry pending copies</Button>
              </Card>
            )}
            <Card>
              <Heading level={2}>Recent slices</Heading>
              {data.projectSlices.map((item, i) => (
                <Row key={i}>
                  <Text>{item.label}</Text>
                  <Button item={item.id} onClick={() => onAction("selectSlice")}>View</Button>
                  <Button item={item.id} onClick={() => onAction("deleteSlice")}>Remove result</Button>
                </Row>
              ))}
            </Card>
          </Stack>
        </Grid>
      )}
      {data.printerDialog && (
        <Card variant="dialog" aria-label="Printer setup">
          <Heading level={2}>Connect your printer</Heading>
          <Alert tone="error">{data.error}</Alert>
          <Alert>On the printer, enable LAN Only and Developer Mode where available. LAN Only disables Bambu Handy/cloud access. No firmware changes are made here.</Alert>
          <Button onClick={() => onAction("discover")}>Find printers on this network</Button>
          {data.discovered.map((item, i) => (
            <Button key={i} item={item.serial} onClick={() => onAction("chooseDiscovered")}>{item.label}</Button>
          ))}
          <Grid columns={2}>
            <Field label="Printer name"><Input name="printerName" /></Field>
            <Field label="Model"><Select name="printerModel" options={data.modelOptions} /></Field>
            <Field label="Local IP or hostname"><Input name="printerHost" /></Field>
            <Field label="Serial from printer screen"><Input name="printerSerial" /></Field>
            <Field label="LAN access code"><Input type="password" name="accessCode" autoComplete="new-password" /></Field>
            <Field label="Physical build plate"><Select name="printerPlate" options={data.plateOptions} /></Field>
          </Grid>
          <Checkbox name="nozzleConfirmed" label="I checked that the installed nozzle is 0.4 mm" />
          <Checkbox name="developerConfirmed" label="I enabled local commands / Developer Mode on this printer" />
          <Checkbox name="manualIdentity" label="I checked the serial and model on the physical screen (for firmware without discovery)" />
          <Checkbox name="protectedData" label="Use encrypted FTPS data channel (only if this firmware supports it)" />
          <Text variant="caption">Access codes are encrypted for this Windows user, never stored in browser storage. A blank code keeps the existing saved code.</Text>
          <Row>
            <Button variant="primary" onClick={() => onAction("savePrinter")}>Save printer</Button>
            <Button onClick={() => onAction("probe")} disabled={data.noPrinterID}>Inspect certificates</Button>
            <Button onClick={() => onAction("testPrinter")} disabled={data.noPrinterID}>Test connections</Button>
          </Row>
          {data.trust && (
            <>
              <Alert tone="warning">Review the two local certificate fingerprints against this device before trusting. A changed certificate requires review again.</Alert>
              <Text variant="code">{data.trustText}</Text>
              <Checkbox name="trustConfirmed" label="I confirm these certificates belong to my selected printer" />
              <Button onClick={() => onAction("trustPrinter")} disabled={data.notTrusted}>Trust this device</Button>
            </>
          )}
          {data.printerDiagnostics.map((item, i) => (
            <Alert key={i}>{item.message}</Alert>
          ))}
          <Button onClick={() => onAction("close")} cancel={true}>Done</Button>
        </Card>
      )}
      {data.printDialog && (
        <Card variant="dialog" aria-label="Confirm print">
          <Heading level={2}>Ready to start this print?</Heading>
          <Alert tone="error">{data.error}</Alert>
          <Text>{data.printSummary}</Text>
          <Text>Check the physical bed, nozzle, plate and loaded material. Starting sends this exact archive to the selected printer.</Text>
          <Checkbox name="useAMS" label="Use the connected AMS lite" />
          {data.useAMS ? (
            <>
              {data.materialMappings.map((item, i) => (
                <Field key={i} label={item.label}><Select name={item.name} options={data.trayOptions} /></Field>
              ))}
              <Checkbox name="acceptTrayDifferences" label="I reviewed and accept any color / preset differences" />
            </>
          ) : (
            <Checkbox name="externalConfirmed" label="The correct single filament is loaded on the external spool" />
          )}
          <Checkbox name="bedLeveling" label="Run bed leveling" />
          <Checkbox name="printConfirmed" label="I checked this physical printer and want to start now" />
          <Row>
            <Button variant="primary" onClick={() => onAction("print")} disabled={data.printNotConfirmed}>Print now</Button>
            <Button onClick={() => onAction("close")} cancel={true}>Back</Button>
          </Row>
        </Card>
      )}
      {data.driveDialog && (
        <Card variant="dialog" aria-label="Open from Drive">
          <Heading level={2}>Open STL or Bambu 3MF from Drive</Heading>
          <Alert tone="error">{data.error}</Alert>
          <Field label="Source"><Select name="driveSource" options={data.driveSources} /></Field>
          <Text>{data.driveFolder}</Text>
          <Button onClick={() => onAction("driveUp")}>Parent folder</Button>
          {data.driveEntries.map((item, i) => (
            <Button key={i} item={item.path} onClick={() => onAction("driveEntry")}>{item.label}</Button>
          ))}
          <Button onClick={() => onAction("close")} cancel={true}>Close</Button>
        </Card>
      )}
    </Page>
  )
}
