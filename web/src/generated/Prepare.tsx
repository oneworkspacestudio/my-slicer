// Generated from screens/Prepare.aui; edit the AUI source.
import { Badge, Button, Card, Divider, Field, Grid, Heading, Icon, Input, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Prepare({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="studio-workbench">
      <Section id="printer-rail" />
      <Stack variant="stage">
        <Row variant="toolstrip">
          <Input type="file" name="modelFile" variant="import" label="Add model" accept=".stl,.3mf" />
          {data.workspace && (
            <Button variant="tool" onClick={() => onAction("openDrive")} aria-label="Open from Drive"><Icon name="FolderOpen" size={22} /></Button>
          )}
          <Button variant="tool" onClick={() => onAction("newPlate")} aria-label="New plate"><Icon name="Grid2X2" size={24} /></Button>
          <Divider />
          <Button variant="tool" onClick={() => onAction("arrange")} disabled={data.notEditable} aria-label="Arrange objects"><Icon name="LayoutGrid" size={23} /></Button>
          <Button variant="tool" onClick={() => onAction("toggleTransform")} disabled={data.notEditable} aria-label="Move, rotate and scale"><Icon name="Move" size={23} /></Button>
          <Button variant="tool" onClick={() => onAction("duplicate")} disabled={data.notEditable} aria-label="Duplicate object"><Icon name="Copy" size={23} /></Button>
          <Button variant="tool" onClick={() => onAction("remove")} disabled={data.notEditable} aria-label="Remove object"><Icon name="Trash2" size={23} /></Button>
          <Button variant="tool" onClick={() => onAction("fitView")} aria-label="Fit plate to view"><Icon name="Scan" size={23} /></Button>
          <Button variant="tool" onClick={() => onAction("topView")} aria-label="Top view"><Icon name="Box" size={23} /></Button>
          <Spacer />
          <Badge>{data.plateLabel}</Badge>
        </Row>
        <Section id="scene" />
        {data.noObjects && (
          <Stack variant="stage-empty">
            <Icon name="Box" size={40} />
            <Heading level={1}>Prepare your plate</Heading>
            <Text>Add an STL or Bambu 3MF model to begin.</Text>
          </Stack>
        )}
        {data.showTransform && (
          <Card variant="transform">
            <Heading level={3}>Move, rotate & scale</Heading>
            <Field label="Object"><Select name="objectID" options={data.objectOptions} /></Field>
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
              <Button variant="secondary" onClick={() => onAction("toggleTransform")} cancel={true}>Close</Button>
            </Row>
          </Card>
        )}
        <Row variant="stage-status">
          <Text variant="caption">{data.sceneNotice}</Text>
          <Spacer />
          {data.importNotice && (
            <Text variant="caption">{data.importNotice}</Text>
          )}
        </Row>
      </Stack>
    </Page>
  )
}
