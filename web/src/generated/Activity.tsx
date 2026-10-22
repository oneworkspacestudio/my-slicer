// Generated from screens/Activity.aui; edit the AUI source.
import { Badge, Button, Card, Heading, Link, Page, Row, Section, Spacer, Stack, Text } from "@/components/ui"

export function Activity({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Heading level={1}>Activity</Heading>
          <Text variant="muted">Print jobs, recent slices and your Drive copies.</Text>
        </Stack>
        <Spacer />
        <Link href="#" variant="button-secondary">Back to Prepare</Link>
      </Row>
      <Card>
        <Heading level={2}>Print jobs</Heading>
        {data.noPrints && (
          <Text>No print jobs yet. Review a ready slice and choose Print now when you are ready.</Text>
        )}
        {data.prints.map((item, i) => (
          <Section key={i} variant="activity-item">
            <Row>
              <Heading level={3}>{item.state}</Heading>
              <Badge>{item.printerName}</Badge>
            </Row>
            <Text>{item.message}</Text>
            <Text>{item.progressText}</Text>
            <Row gap="sm">
              <Button variant="secondary" item={item.id} onClick={() => onAction("pause")}>Pause</Button>
              <Button variant="secondary" item={item.id} onClick={() => onAction("resume")}>Resume</Button>
              <Button variant="secondary" item={item.id} onClick={() => onAction("cancelPrint")}>Cancel print…</Button>
              <Button variant="secondary" item={item.id} onClick={() => onAction("reconcile")}>Refresh status</Button>
              {item.unknown && (
                <Button variant="secondary" item={item.id} onClick={() => onAction("resolve")}>Resolve uncertain start…</Button>
              )}
            </Row>
          </Section>
        ))}
      </Card>
      <Card>
        <Heading level={2}>Slices for this plate</Heading>
        {data.noSlices && (
          <Text>Slice results appear here after you choose Slice plate.</Text>
        )}
        {data.projectSlices.map((item, i) => (
          <Row key={i} variant="library-item">
            <Text>{item.label}</Text>
            <Spacer />
            <Button variant="secondary" item={item.id} onClick={() => onAction("selectSlice")}>View</Button>
            <Button variant="secondary" item={item.id} onClick={() => onAction("deleteSlice")}>Remove result</Button>
          </Row>
        ))}
      </Card>
      {data.workspace && (
        <Card>
          <Heading level={2}>Drive copies</Heading>
          <Text>Originals and completed slices are copied as ordinary files while this app is open.</Text>
          {data.driveSaves.map((item, i) => (
            <Row key={i} variant="library-item">
              <Text>{item.name}</Text>
              <Badge>{item.state}</Badge>
              <Text>{item.message}</Text>
            </Row>
          ))}
          <Button variant="secondary" onClick={() => onAction("retryDrive")}>Retry pending copies</Button>
        </Card>
      )}
    </Page>
  )
}
