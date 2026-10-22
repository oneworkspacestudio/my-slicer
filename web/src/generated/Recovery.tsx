// Generated from screens/Recovery.aui; edit the AUI source.
import { Button, Card, Heading, Page, Text } from "@/components/ui"

export function Recovery({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Card variant="empty">
        <Heading level={1}>Let’s reopen your slicer</Heading>
        <Text>This screen could not be displayed. Your saved models and printer details are still on your server.</Text>
        <Button onClick={() => onAction("reload")}>Reload slicer</Button>
        <Button variant="secondary" onClick={() => onAction("prepare")}>Return to Prepare</Button>
      </Card>
    </Page>
  )
}
