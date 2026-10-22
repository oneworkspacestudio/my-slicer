// Generated from screens/Dialogs.aui; edit the AUI source.
import { Alert, Button, Card, Checkbox, Field, Heading, Input, Page, Row, Select, Text } from "@/components/ui"

export function Dialogs({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      {data.saveProjectDialog && (
        <Card variant="dialog" aria-label="Save project">
          <Heading level={2}>Keep this as a project</Heading>
          <Text>Your plate is already kept in Recent plates. Give it a name to make it a reusable project.</Text>
          <Alert tone="error">{data.error}</Alert>
          <Field label="Project name"><Input name="projectName" autoFocus={true} /></Field>
          <Row>
            <Button onClick={() => onAction("saveNamedProject")}>Save project</Button>
            <Button variant="secondary" onClick={() => onAction("close")} cancel={true}>Cancel</Button>
          </Row>
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
            <Button variant="secondary" onClick={() => onAction("close")} cancel={true}>Back</Button>
          </Row>
        </Card>
      )}
      {data.driveDialog && (
        <Card variant="dialog" aria-label="Open from Drive">
          <Heading level={2}>Open STL or Bambu 3MF from Drive</Heading>
          <Alert tone="error">{data.error}</Alert>
          <Field label="Source"><Select name="driveSource" options={data.driveSources} /></Field>
          <Text>{data.driveFolder}</Text>
          <Button variant="secondary" onClick={() => onAction("driveUp")}>Parent folder</Button>
          {data.driveEntries.map((item, i) => (
            <Button key={i} variant="secondary" item={item.path} onClick={() => onAction("driveEntry")}>{item.label}</Button>
          ))}
          <Button variant="secondary" onClick={() => onAction("close")} cancel={true}>Close</Button>
        </Card>
      )}
    </Page>
  )
}
