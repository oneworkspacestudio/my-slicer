// Generated from screens/Shell.aui; edit the AUI source.
import { Alert, Button, Card, Divider, Heading, Icon, Link, Page, Row, Section, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Shell({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="studio">
      <Row variant="menubar">
        <Section variant="disclosure" label="☰ File⌄">
          <Text variant="caption">{data.user.name}</Text>
          <Button variant="secondary" onClick={() => onAction("newPlate")}>New plate</Button>
          <Link href="#projects">Open project…</Link>
          <Button variant="secondary" onClick={() => onAction("showSaveProject")} disabled={data.noObjects}>Save project…</Button>
          <Link href="#activity">Print history</Link>
          <Link href="#settings">Settings & backup</Link>
          <Button variant="secondary" onClick={() => onAction("logout")}>Sign out</Button>
        </Section>
        <Divider />
        <Button variant="tool" onClick={() => onAction("showSaveProject")} disabled={data.noObjects} aria-label="Save project"><Icon name="Save" size={17} /></Button>
        <Button variant="tool" onClick={() => onAction("undo")} disabled={data.noUndo} aria-label="Undo"><Icon name="Undo2" size={17} /></Button>
        <Spacer />
        <Text>{data.plateName}</Text>
        <Spacer />
        <Text variant="caption">My Slicer</Text>
        <Button variant="tool" onClick={() => onAction("theme")} aria-label="Toggle color theme"><Icon name="SunMoon" size={16} /></Button>
        {data.workspace && (
          <Button variant="tool" workspaceTrigger={true} onClick={() => onAction("apps")}>Apps</Button>
        )}
        <Link href="#settings" variant="menu" aria-label="Settings"><Icon name="Settings2" size={16} /></Link>
      </Row>
      <Row variant="studio-nav">
        <Link href="#" variant="home" aria-label="Home"><Icon name="House" size={21} /></Link>
        {data.nav.map((item, i) => (
          <Link key={i} href={item.href} variant={item.variant}>
            <Icon name={item.icon} size={18} />
            <Text>{item.title}</Text>
          </Link>
        ))}
        <Spacer />
        <Button variant="slice" onClick={() => onAction("slice")} disabled={data.cannotSlice}>Slice plate</Button>
        <Button variant="print" onClick={() => onAction("reviewPrint")} disabled={data.cannotReviewPrint}>Print plate</Button>
      </Row>
      <Stack variant="studio-content">
        {data.error && (
          <Alert tone="error">{data.error}</Alert>
        )}
        {data.notice && (
          <Alert tone="success">{data.notice}</Alert>
        )}
        {data.busy && (
          <Row variant="busy">
            <Spinner />
            <Text>{data.busy}</Text>
          </Row>
        )}
        <Section id="screen-outlet" />
      </Stack>
      {data.appsOpen && (
        <Card variant="dialog" aria-label="Workspace apps">
          <Heading level={2}>Your Workspace apps</Heading>
          {data.appLinks.map((item, i) => (
            <Link key={i} href={item.href}>{item.title}</Link>
          ))}
          <Button variant="secondary" onClick={() => onAction("closeApps")}>Close</Button>
        </Card>
      )}
    </Page>
  )
}
