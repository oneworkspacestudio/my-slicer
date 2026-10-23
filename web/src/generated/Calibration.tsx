// Generated from screens/Calibration.aui; edit the AUI source.
import { Badge, Button, Card, Grid, Heading, Icon, Link, Page, Row, Section, Text } from "@/components/ui"

export function Calibration({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="studio-page">
      <Row variant="page-header">
        <Heading level={1}>Calibration</Heading>
        <Link href="#printers" variant="button-secondary">Open Device</Link>
      </Row>
      <Grid columns={2}>
        <Card>
          <Icon name="ShieldCheck" size={32} />
          <Heading level={2}>Native slicing engine</Heading>
          <Badge>{data.engine.phase}</Badge>
          <Text>{data.engine.message}</Text>
          <Text>Engine setup validates a real A1 and A1 mini slice before profiles become available.</Text>
          {data.engineNeeded ? (
            <>
              <Button onClick={() => onAction("installEngine")} disabled={data.engine.busy}>Install slicing engine</Button>
              <Section id="engine-progress" />
            </>
          ) : (
            <Link href="#" variant="button">Prepare a calibration model</Link>
          )}
        </Card>
        <Card>
          <Icon name="SlidersHorizontal" size={32} />
          <Heading level={2}>Printer calibration</Heading>
          <Text>Run bed leveling, vibration compensation and flow calibration on the printer’s screen. Remote calibration commands are unavailable through this connection.</Text>
          <Text>Check the physical nozzle and build plate match the selected profile before printing a calibration model.</Text>
          <Link href="#printer-new" variant="button-secondary">Add or check a printer</Link>
        </Card>
      </Grid>
      <Link href="#settings">Engine source, licenses & backup</Link>
    </Page>
  )
}
