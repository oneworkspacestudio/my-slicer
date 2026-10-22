// Generated from screens/Preview.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Field, Grid, Heading, Icon, Input, Link, Metric, Page, Row, Section, Select, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Preview({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Text variant="eyebrow">PREVIEW · 2 OF 3</Text>
          <Heading level={1}>Check before you print</Heading>
          <Text variant="muted">Inspect the actual sliced layers and material estimates.</Text>
        </Stack>
        <Spacer />
        <Link href="#" variant="button-secondary">← Back to Prepare</Link>
      </Row>
      {data.activeSlice ? (
        <Grid columns={2} variant="slicer">
          <Card variant="plate">
            <Heading level={2}>{data.plateName}</Heading>
            <Section id="scene" />
            <Text variant="caption">{data.sceneNotice}</Text>
            {data.sliceReady && (
              <>
                <Field label="Preview layer"><Input name="layer" type="range" min={1} max={data.maxLayer} step={1} /></Field>
                <Text>{data.layerText}</Text>
                <Field label="Toolpath feature"><Select name="feature" options={data.featureOptions} /></Field>
              </>
            )}
          </Card>
          <Card variant="print-settings">
            <Heading level={2}>{data.sliceTitle}</Heading>
            <Text>{data.activeSlice.message}</Text>
            <Text variant="caption">{data.sliceElapsed}</Text>
            {data.slicing && (
              <>
                <Spinner />
                <Button variant="secondary" onClick={() => onAction("cancelSlice")}>Cancel slicing</Button>
              </>
            )}
            {data.sliceReady && (
              <>
                <Row gap="md">
                  <Metric value={data.timeText} />
                  <Text>{data.filamentText}</Text>
                  <Badge>{data.layerCount}</Badge>
                </Row>
                <Button variant="secondary" onClick={() => onAction("downloadSlice")}>Download sliced 3MF</Button>
                {data.noSliceTarget ? (
                  <>
                    <Link href="#printers" variant="button-secondary">Add a printer to print directly</Link>
                    <Text variant="caption">Select the printer on Prepare and slice again for that device.</Text>
                  </>
                ) : (
                  <>
                    <Button variant="primary" onClick={() => onAction("reviewPrint")} disabled={data.cannotReviewPrint}>Review print →</Button>
                    {data.printerNeedsConnection && (
                      <Link href="#printers">Connect and test your printer first</Link>
                    )}
                  </>
                )}
                {data.staleSlice && (
                  <Alert tone="warning">{data.staleReason}</Alert>
                )}
              </>
            )}
            {data.sliceFailed && (
              <>
                <Alert tone="warning">Your model is kept. Return to Prepare, check the settings and try slicing again.</Alert>
                <Link href="#" variant="button">Back to Prepare</Link>
              </>
            )}
          </Card>
        </Grid>
      ) : (
        <Card variant="empty">
          <Icon name="Layers" size={40} />
          <Heading level={2}>Your slice will appear here</Heading>
          <Text>Add a model on Prepare, choose your settings and select Slice plate.</Text>
          <Link href="#" variant="button">Go to Prepare</Link>
        </Card>
      )}
    </Page>
  )
}
