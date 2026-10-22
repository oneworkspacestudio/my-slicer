// Generated from screens/SceneStatus.aui; edit the AUI source.
import { Alert, Button, Page, Row, Spinner, Stack, Text } from "@/components/ui"

export function SceneStatus({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      {data.loading && (
        <Alert>
          <Stack gap="xs">
            <Row>
              <Spinner />
              <Text>{data.loading}</Text>
            </Row>
            <Text variant="caption">Your model will appear here. You can keep moving the view while it loads.</Text>
          </Stack>
        </Alert>
      )}
      {data.error && (
        <>
          <Alert tone="error">{data.error}</Alert>
          <Button variant="secondary" onClick={() => onAction("retry")}>Retry model preview</Button>
        </>
      )}
      {data.ready && (
        <Text variant="caption">Model ready</Text>
      )}
    </Page>
  )
}
