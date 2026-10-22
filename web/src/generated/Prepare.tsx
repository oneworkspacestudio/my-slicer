// Generated from screens/Prepare.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Grid, Heading, Input, Link, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Prepare({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Text variant="eyebrow">PREPARE · 1 OF 3</Text>
          <Heading level={1}>Start with your model</Heading>
          <Text variant="muted">Add a model, set up the plate, then preview your slice.</Text>
        </Stack>
      </Row>
      <Row variant="importbar">
        <Input type="file" name="modelFile" variant="import" label="Add model" accept=".stl,.3mf" />
        {data.workspace && (
          <Button variant="secondary" onClick={() => onAction("openDrive")}>Open from Drive</Button>
        )}
        <Spacer />
        <Button variant="secondary" onClick={() => onAction("showSaveProject")} disabled={data.noObjects}>Save project…</Button>
        <Button variant="secondary" onClick={() => onAction("newPlate")}>New plate</Button>
      </Row>
      <Grid columns={2} variant="slicer">
        <Stack gap="md">
          <Card variant="plate">
            <Row>
              <Heading level={2}>Build plate</Heading>
              <Spacer />
              <Badge>{data.plateLabel}</Badge>
            </Row>
            {data.noObjects && (
              <Text variant="muted">STL or Bambu 3MF · No project setup needed. Choose Add model above to begin.</Text>
            )}
            <Section id="scene" />
            <Text variant="caption">{data.sceneNotice}</Text>
            {data.importNotice && (
              <Alert>{data.importNotice}</Alert>
            )}
            {data.hasObjects && (
              <Row>
                <Field label="Object"><Select name="objectID" options={data.objectOptions} /></Field>
                <Button variant="secondary" onClick={() => onAction("duplicate")} disabled={data.notEditable}>Duplicate</Button>
                <Button variant="secondary" onClick={() => onAction("remove")} disabled={data.notEditable}>Remove</Button>
                <Button variant="secondary" onClick={() => onAction("arrange")} disabled={data.notEditable}>Arrange</Button>
                <Button variant="secondary" onClick={() => onAction("undo")} disabled={data.noUndo}>Undo</Button>
              </Row>
            )}
            {data.editable && (
              <Section variant="disclosure" label="Position, rotation & scale">
                <Grid columns={3}>
                  <Field label="X (mm)"><Input name="x" type="number" step={1} /></Field>
                  <Field label="Y (mm)"><Input name="y" type="number" step={1} /></Field>
                  <Field label="Z (mm)"><Input name="z" type="number" step={1} /></Field>
                  <Field label="Rotate X (°)"><Input name="rx" type="number" step={5} /></Field>
                  <Field label="Rotate Y (°)"><Input name="ry" type="number" step={5} /></Field>
                  <Field label="Rotate Z (°)"><Input name="rz" type="number" step={5} /></Field>
                  <Field label="Scale (%)"><Input name="scale" type="number" min={1} max={10000} /></Field>
                </Grid>
                <Row>
                  <Button onClick={() => onAction("transform")}>Apply changes</Button>
                  <Button variant="secondary" onClick={() => onAction("reset")}>Reset position</Button>
                </Row>
              </Section>
            )}
          </Card>
        </Stack>
        <Card variant="print-settings">
          <Heading level={2}>Print settings</Heading>
          {data.engineNeeded && (
            <Section variant="engine">
              <Heading level={3}>One-time slicing engine setup</Heading>
              <Text>{data.engine.message}</Text>
              {data.engine.busy && (
                <Text variant="caption">{data.engineProgress}</Text>
              )}
              <Section id="engine-progress" />
              <Text variant="caption">474 MB download · You can add models while setup runs.</Text>
              {data.owner ? (
                <>
                  <Button onClick={() => onAction("installEngine")} disabled={data.engine.busy}>Install slicing engine</Button>
                  {data.engine.busy && (
                    <Button variant="secondary" onClick={() => onAction("cancelEngine")} cancel={true}>Cancel setup</Button>
                  )}
                </>
              ) : (
                <Text>Ask the Workspace owner to finish engine setup.</Text>
              )}
            </Section>
          )}
          <Field label="Printer"><Select name="target" options={data.printerOptions} /></Field>
          <Link href="#printers">Add or manage printers</Link>
          <Grid columns={2}>
            <Field label="Printer model"><Select name="model" options={data.modelOptions} disabled={data.hasTarget} /></Field>
            <Field label="Build plate"><Select name="plate" options={data.plateOptions} disabled={data.hasTarget} /></Field>
          </Grid>
          <Text variant="caption">Verified 0.4 mm nozzle profiles. Pairing a printer is optional for slicing.</Text>
          {data.nativeProject && (
            <>
              <Field label="Imported plate"><Select name="nativePlate" options={data.nativePlates} /></Field>
              <Checkbox name="preserveImported" label="Keep imported settings, modifiers and painting" />
            </>
          )}
          {data.engineReady && (
            <>
              <Field label="Material"><Select name="filament" options={data.filamentOptions} disabled={data.preserving} /></Field>
              <Field label="Quality"><Select name="quality" options={data.processOptions} disabled={data.preserving} /></Field>
              <Section variant="disclosure" label="Advanced print settings">
                <Grid columns={2}>
                  <Field label="Layer height (mm)"><Input name="layerHeight" type="number" step={0.01} min={0.08} max={0.28} disabled={data.preserving} /></Field>
                  <Field label="Infill (%)"><Input name="infill" type="number" min={0} max={100} disabled={data.preserving} /></Field>
                  <Field label="Walls"><Input name="walls" type="number" min={1} max={20} disabled={data.preserving} /></Field>
                  <Field label="Brim"><Select name="brim" options={data.brimOptions} disabled={data.preserving} /></Field>
                </Grid>
                <Checkbox name="supports" label="Generate supports" disabled={data.preserving} />
                <Button variant="secondary" onClick={() => onAction("saveSettings")} disabled={data.noObjects}>Apply settings</Button>
              </Section>
            </>
          )}
          <Button variant="primary" onClick={() => onAction("slice")} disabled={data.cannotSlice}>Slice plate →</Button>
          <Text variant="caption">{data.sliceHint}</Text>
        </Card>
      </Grid>
    </Page>
  )
}
