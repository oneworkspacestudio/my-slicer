// Generated from screens/Projects.aui; edit the AUI source.
import { Button, Card, Field, Grid, Heading, Input, Link, Page, Row, Spacer, Stack, Text } from "@/components/ui"

export function Projects({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="page-header">
        <Stack gap="xs">
          <Heading level={1}>Saved projects</Heading>
          <Text variant="muted">Keep a named project when you want to reuse a plate.</Text>
        </Stack>
        <Spacer />
        <Link href="#" variant="button">Add a model</Link>
      </Row>
      {data.hasProject && (
        <Card variant="save-project">
          <Heading level={2}>Save the current plate</Heading>
          <Text>Naming a project is optional. Your recent plate is already kept for you.</Text>
          <Field label="Project name"><Input name="projectName" placeholder="Desk organiser" /></Field>
          <Button onClick={() => onAction("saveNamedProject")}>Save project</Button>
        </Card>
      )}
      <Grid columns={2}>
        <Card>
          <Heading level={2}>Named projects</Heading>
          {data.noSavedProjects && (
            <Text>No named projects yet. You can still add models, slice and print.</Text>
          )}
          {data.savedProjects.map((item, i) => (
            <Row key={i} variant="library-item">
              <Stack gap="xs">
                <Text>{item.name}</Text>
                <Text variant="caption">{item.summary}</Text>
              </Stack>
              <Spacer />
              <Button variant="secondary" item={item.id} onClick={() => onAction("openProject")}>Open</Button>
            </Row>
          ))}
        </Card>
        <Card>
          <Heading level={2}>Recent plates</Heading>
          <Text variant="caption">Unsaved plates are kept automatically so you can return to your work.</Text>
          {data.noRecentPlates && (
            <Link href="#">Start with a model</Link>
          )}
          {data.recentPlates.map((item, i) => (
            <Row key={i} variant="library-item">
              <Stack gap="xs">
                <Text>{item.name}</Text>
                <Text variant="caption">{item.summary}</Text>
              </Stack>
              <Spacer />
              <Button variant="secondary" item={item.id} onClick={() => onAction("openProject")}>Open plate</Button>
            </Row>
          ))}
        </Card>
      </Grid>
    </Page>
  )
}
