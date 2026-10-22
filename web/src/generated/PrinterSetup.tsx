// Generated from screens/PrinterSetup.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Grid, Heading, Input, Link, Page, Row, Section, Select, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function PrinterSetup({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Text variant="eyebrow">PRINTER SETUP</Text>
          <Heading level={1}>{data.printerSetupTitle}</Heading>
          <Text variant="muted">A guided connection to your printer. No print is sent during setup.</Text>
        </Stack>
        <Spacer />
        <Link href="#printers" variant="button-secondary">← Back to printers</Link>
      </Row>
      <Card variant="printer-setup">
        <Row variant="setup-steps">
          <Badge tone={data.detailsTone}>1 · Details</Badge>
          <Badge tone={data.trustTone}>2 · Trust device</Badge>
          <Badge tone={data.checkTone}>3 · Check connection</Badge>
        </Row>
        {data.printerDetailsStep ? (
          <>
            <Heading level={2}>Find these details on your printer</Heading>
            <Text>Enable LAN Only and local commands / Developer Mode where supported. LAN Only turns off Bambu Handy and cloud access. Then find the IP address, serial and LAN access code on the printer screen.</Text>
            <Button variant="secondary" onClick={() => onAction("discover")}>Find printers on my network</Button>
            {data.discoveryEmpty && (
              <Text>No printers found. You can enter the details below.</Text>
            )}
            {data.discovered.map((item, i) => (
              <Button key={i} variant="secondary" item={item.serial} onClick={() => onAction("chooseDiscovered")}>{item.label}</Button>
            ))}
            <Grid columns={2}>
              <Field label="Printer name"><Input name="printerName" placeholder="Desk printer" /></Field>
              <Field label="Model"><Select name="printerModel" options={data.modelOptions} /></Field>
              <Field label="Local IP or hostname"><Input name="printerHost" placeholder="192.168.1.50" /></Field>
              <Field label="Serial number"><Input name="printerSerial" placeholder="From the printer screen" /></Field>
              <Field label="LAN access code"><Input type="password" name="accessCode" autoComplete="new-password" /></Field>
              <Field label="Physical build plate"><Select name="printerPlate" options={data.plateOptions} /></Field>
            </Grid>
            <Text variant="caption">Access codes are encrypted on your server and never saved in browser storage. When editing, leave the code blank to keep the saved code.</Text>
            <Checkbox name="nozzleConfirmed" label="I checked the installed nozzle is 0.4 mm" />
            <Checkbox name="developerConfirmed" label="I enabled local commands / Developer Mode" />
            <Section variant="disclosure" label="Advanced connection options">
              <Checkbox name="manualIdentity" label="I checked the serial and model on the physical screen (firmware without discovery)" />
              <Checkbox name="protectedData" label="Use encrypted FTPS data (only if this firmware supports it)" />
            </Section>
            <Row>
              <Button variant="primary" onClick={() => onAction("savePrinter")}>Save and continue →</Button>
              <Link href="#printers">Cancel</Link>
            </Row>
          </>
        ) : (
          <>
            <Heading level={2}>{data.editingPrinterName}</Heading>
            {data.printerTrustStep ? (
              <>
                <Text>Confirm that you are connecting to your printer before trusting it. Inspect the local certificates and review both fingerprints against this device.</Text>
                <Button onClick={() => onAction("probe")}>Inspect certificates</Button>
                {data.trust && (
                  <>
                    <Text variant="code">{data.trustText}</Text>
                    <Checkbox name="trustConfirmed" label="I confirm these certificates belong to this printer" />
                    <Button variant="primary" onClick={() => onAction("trustPrinter")} disabled={data.notTrusted}>Trust this device and continue →</Button>
                  </>
                )}
              </>
            ) : (
              <>
                <Text>Check that the printer is switched on, reachable from this server and ready for local commands.</Text>
                <Button variant="primary" onClick={() => onAction("testPrinter")} disabled={data.printerTesting}>Test connection</Button>
                {data.printerTesting && (
                  <>
                    <Spinner />
                    <Text>Checking connection…</Text>
                  </>
                )}
                <Text>{data.printerConnectionText}</Text>
                {data.printerDiagnostics.map((item, i) => (
                  <Alert key={i}>{item.message}</Alert>
                ))}
                <Link href="#printers" variant="button">Done</Link>
              </>
            )}
            <Button variant="secondary" onClick={() => onAction("printerDetails")}>← Edit details</Button>
          </>
        )}
      </Card>
      <Link href="#">Back to Prepare</Link>
    </Page>
  )
}
