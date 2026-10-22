// Generated from screens/Printers.aui; edit the AUI source.
import { Badge, Button, Card, Heading, Icon, Link, Page, Row, Spacer, Stack, Text } from "@/components/ui"

export function Printers({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Text variant="eyebrow">CONNECT · 3 OF 3</Text>
          <Heading level={1}>Your printers</Heading>
          <Text variant="muted">Connect locally when you are ready to print.</Text>
        </Stack>
        <Spacer />
        <Button variant="primary" onClick={() => onAction("addPrinter")}>Add printer</Button>
      </Row>
      {data.noPrinters && (
        <Card variant="empty">
          <Icon name="Printer" size={44} />
          <Heading level={2}>Connect your first printer</Heading>
          <Text>Add an A1 or A1 mini to print on your local network. You can prepare and slice models before pairing.</Text>
          <Button variant="primary" onClick={() => onAction("addPrinter")}>Add printer</Button>
          <Link href="#">Keep preparing a model</Link>
        </Card>
      )}
      {data.printers.map((item, i) => (
        <Card key={i}>
          <Row>
            <Icon name="Printer" />
            <Heading level={2}>{item.config.name}</Heading>
            <Spacer />
            <Badge>{item.status.status}</Badge>
          </Row>
          <Text>{item.detail}</Text>
          {item.diagnostics.map((item, i) => (
            <Text key={i} variant="caption">{item.message}</Text>
          ))}
          <Button variant="secondary" item={item.config.id} onClick={() => onAction("editPrinter")}>Setup & diagnostics</Button>
        </Card>
      ))}
      <Text variant="caption">Adding or testing a printer never starts a print. Select a saved printer on Prepare before slicing for it.</Text>
    </Page>
  )
}
