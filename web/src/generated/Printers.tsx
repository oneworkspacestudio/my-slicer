// Generated from screens/Printers.aui; edit the AUI source.
import { Badge, Button, Card, Divider, Grid, Heading, Icon, Link, Page, Row, Section, Spacer, Stack, Text } from "@/components/ui"

export function Printers({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="device">
      <Stack variant="device-sidebar">
        <Row variant="rail-heading">
          <Icon name="Printer" size={17} />
          <Text>Device</Text>
          <Spacer />
          <Button variant="tool" onClick={() => onAction("addPrinter")} aria-label="Add printer"><Icon name="Plus" size={19} /></Button>
        </Row>
        {data.printers.map((item, i) => (
          <Button key={i} variant={item.deviceVariant} item={item.config.id} onClick={() => onAction("selectDevice")}>{item.config.name}</Button>
        ))}
        <Text variant="device-selected">Status</Text>
        <Link href="#activity" variant="device-item">Print history</Link>
        <Link href="#projects" variant="device-item">Projects</Link>
        <Link href="#calibration" variant="device-item">Calibration</Link>
        <Button variant="secondary" onClick={() => onAction("addPrinter")}>Add printer</Button>
      </Stack>
      <Stack variant="device-main">
        <Stack variant="camera-section">
          <Row variant="panel-heading">
            <Text>Camera</Text>
            <Spacer />
            <Icon name="VideoOff" size={18} />
          </Row>
          <Stack variant="camera">
            <Icon name="VideoOff" size={42} />
            <Heading level={2}>Camera unavailable</Heading>
            <Text>Live video is not exposed by the supported local printer connection.</Text>
            {data.noPrinters ? (
              <Button onClick={() => onAction("addPrinter")}>Connect your first printer</Button>
            ) : (
              <Text>{data.deviceConnection}</Text>
            )}
          </Stack>
        </Stack>
        <Card variant="printing-progress">
          <Row variant="panel-heading">
            <Text>Printing Progress</Text>
            <Spacer />
            <Link href="#activity">View history</Link>
          </Row>
          {data.devicePrint ? (
            <>
              <Heading level={3}>{data.devicePrint.name}</Heading>
              <Text>{data.devicePrint.state}</Text>
              <Section id="print-progress" />
              <Text>{data.devicePrint.detail}</Text>
            </>
          ) : (
            <>
              <Icon name="Printer" size={33} />
              <Text>No tracked print on this device</Text>
              <Text variant="caption">Start from a reviewed slice. Status appears here when the printer reports it.</Text>
            </>
          )}
        </Card>
      </Stack>
      <Stack variant="device-controls">
        <Row variant="panel-heading">
          <Text>Control</Text>
          <Spacer />
          {data.device && (
            <Button variant="print" item={data.device.config.id} onClick={() => onAction("editPrinter")}>Printer settings</Button>
          )}
        </Row>
        <Card variant="telemetry">
          <Row variant="device-control-top">
            <Stack variant="temperatures">
              <Icon name="Thermometer" size={20} />
              <Text>{data.nozzleText}</Text>
              <Divider />
              <Icon name="PanelBottom" size={20} />
              <Text>{data.bedText}</Text>
              <Divider />
              <Text>{data.deviceState}</Text>
            </Stack>
            <Stack variant="motion">
              <Text>Y</Text>
              <Row>
                <Text>−X</Text>
                <Button variant="home-motion" disabled={true} aria-label="Printer motion unavailable"><Icon name="House" size={28} /></Button>
                <Text>X</Text>
              </Row>
              <Text>−Y</Text>
            </Stack>
            <Stack variant="extruder">
              <Button variant="secondary" disabled={true} aria-label="Extruder motion unavailable"><Icon name="ChevronUp" size={22} /></Button>
              <Icon name="Printer" size={35} />
              <Button variant="secondary" disabled={true} aria-label="Extruder motion unavailable"><Icon name="ChevronDown" size={22} /></Button>
              <Text variant="caption">Extruder</Text>
            </Stack>
          </Row>
          <Text variant="caption">Motion and extrusion controls are unavailable through this connection.</Text>
        </Card>
        <Card variant="ams">
          <Row variant="panel-heading">
            <Text>Filament · AMS</Text>
            <Spacer />
            <Badge>{data.trayCount}</Badge>
          </Row>
          <Grid columns={2} variant="trays">
            {data.deviceTrays.map((item, i) => (
              <Stack key={i} variant="tray">
                <Icon name="Disc3" size={48} />
                <Text>{item.key}</Text>
                <Text>{item.material}</Text>
                <Text variant="caption">{item.color}</Text>
              </Stack>
            ))}
          </Grid>
          {data.noTrays && (
            <Text variant="muted">No live filament trays reported</Text>
          )}
          <Text variant="caption">Filament mapping is confirmed in the print review.</Text>
        </Card>
        {data.device ? (
          <Card variant="device-diagnostics">
            <Heading level={3}>{data.device.config.name}</Heading>
            <Text>{data.deviceDetail}</Text>
            {data.device.diagnostics.map((item, i) => (
              <Text key={i} variant="caption">{item.message}</Text>
            ))}
            <Button variant="secondary" item={data.device.config.id} onClick={() => onAction("editPrinter")}>Setup & diagnostics</Button>
          </Card>
        ) : (
          <Text variant="caption">Pair an A1 or A1 mini on your local network. Adding a device never starts a print.</Text>
        )}
      </Stack>
    </Page>
  )
}
