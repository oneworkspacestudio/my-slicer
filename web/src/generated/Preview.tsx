// Generated from screens/Preview.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Divider, Field, Heading, Icon, Input, Link, Page, Row, Section, Select, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Preview({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="studio-workbench">
      <Section id="printer-rail" />
      <Stack variant="stage">
        <Row variant="toolstrip">
          <Icon name="Layers" size={22} />
          <Text>{data.plateName}</Text>
          <Spacer />
          <Link href="#">Back to Prepare</Link>
        </Row>
        <Section id="scene" />
        {data.activeSlice ? (
          <>
            <Card variant="slice-result">
              <Row>
                <Heading level={3}>Slicing Result</Heading>
                <Spacer />
                <Text>{data.sliceTitle}</Text>
              </Row>
              <Divider />
              <Field label="Color scheme" variant="inline"><Select name="feature" options={data.featureOptions} /></Field>
              <Row variant="result-columns">
                <Text>Filament</Text>
                <Spacer />
                <Text>Total</Text>
              </Row>
              <Row>
                <Badge tone="success">1</Badge>
                <Text>{data.filamentText}</Text>
              </Row>
              <Divider />
              <Heading level={3}>Time Estimation</Heading>
              <Text>{data.timeText}</Text>
              <Text>{data.layerCount}</Text>
              <Text variant="caption">{data.activeSlice.message}</Text>
              <Text variant="caption">{data.sliceElapsed}</Text>
              {data.slicing && (
                <>
                  <Spinner />
                  <Button variant="secondary" onClick={() => onAction("cancelSlice")}>Cancel slicing</Button>
                </>
              )}
              {data.sliceReady && (
                <>
                  <Button variant="secondary" onClick={() => onAction("downloadSlice")}>Export sliced 3MF</Button>
                  {data.noSliceTarget ? (
                    <Link href="#printers">Add a printer to print directly</Link>
                  ) : (
                    <Button onClick={() => onAction("reviewPrint")} disabled={data.cannotReviewPrint}>Print plate</Button>
                  )}
                  {data.staleSlice && (
                    <Alert tone="warning">{data.staleReason}</Alert>
                  )}
                </>
              )}
              {data.sliceFailed && (
                <Alert tone="warning">Your model is kept. Check the settings on Prepare and slice again.</Alert>
              )}
            </Card>
            {data.sliceReady && (
              <>
                <Stack variant="layer-slider">
                  <Text>{data.maxLayer}</Text>
                  <Input name="layer" type="range" min={1} max={data.maxLayer} step={1} aria-label="Preview layer" />
                  <Text>{data.layerText}</Text>
                </Stack>
                <Row variant="path-slider">
                  <Icon name="ChevronsRight" size={18} />
                  <Input name="pathPercent" type="range" min={1} max={100} step={1} aria-label="Visible toolpath percentage" />
                  <Text>{data.pathText}</Text>
                </Row>
              </>
            )}
          </>
        ) : (
          <Stack variant="stage-empty">
            <Icon name="Layers" size={40} />
            <Heading level={1}>Your slice will appear here</Heading>
            <Text>Add a model on Prepare and select Slice plate.</Text>
            <Link href="#" variant="button">Go to Prepare</Link>
          </Stack>
        )}
        <Row variant="stage-status"><Text variant="caption">{data.sceneNotice}</Text></Row>
      </Stack>
    </Page>
  )
}
