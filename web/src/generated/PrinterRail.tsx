// Generated from screens/PrinterRail.aui; edit the AUI source.
import { Fragment } from 'react'
import { Badge, Button, Checkbox, Field, Heading, Icon, Input, Link, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function PrinterRail({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page variant="printer-rail">
      <Row variant="rail-heading">
        <Icon name="Printer" size={16} />
        <Text>Printer</Text>
        <Spacer />
        <Link href="#printers" aria-label="Manage printers"><Icon name="Settings2" size={16} /></Link>
      </Row>
      <Row variant="printer-cards">
        <Stack variant="printer-card">
          <Icon name="Printer" size={49} />
          <Field label="Printer model"><Select name="model" options={data.modelOptions} disabled={data.hasTarget} /></Field>
        </Stack>
        <Stack variant="plate-card">
          <Icon name="Grid2X2" size={42} />
          <Field label="Build plate"><Select name="plate" options={data.plateOptions} disabled={data.hasTarget} /></Field>
        </Stack>
        <Button variant="sync" onClick={() => onAction("syncInfo")} aria-label="Sync info">
          <Icon name="RefreshCw" size={30} />
          <Text>Sync info</Text>
        </Button>
      </Row>
      <Stack variant="nozzle">
        <Text variant="caption">Nozzle</Text>
        <Row>
          <Field label="Diameter" variant="inline"><Select name="nozzle" options={data.nozzleOptions} /></Field>
          <Text>Flow</Text>
          <Badge>Standard</Badge>
        </Row>
      </Stack>
      <Row variant="rail-heading">
        <Icon name="Layers" size={16} />
        <Text>Project Filaments</Text>
      </Row>
      <Stack variant="rail-section">
        {data.engineReady ? (
          <Field label="Filament 1" variant="inline"><Select name="filament" options={data.filamentOptions} disabled={data.preserving} /></Field>
        ) : (
          <Text variant="muted">Profiles available after engine setup</Text>
        )}
        <Field label="Target printer" variant="inline"><Select name="target" options={data.printerOptions} /></Field>
        {data.nativeProject && (
          <>
            <Field label="Imported plate" variant="inline"><Select name="nativePlate" options={data.nativePlates} /></Field>
            <Checkbox name="allPlates" label="Show all plates" disabled={data.overviewUnavailable} />
            {data.overviewUnavailable && (
              <Text variant="caption">For projects with more than 9 plates, choose an imported plate above.</Text>
            )}
            <Checkbox name="preserveImported" label="Keep imported settings, modifiers and painting" />
          </>
        )}
      </Stack>
      <Row variant="rail-heading">
        <Icon name="SlidersHorizontal" size={16} />
        <Text>Process</Text>
        <Badge tone="success">Global</Badge>
        <Spacer />
        <Checkbox name="advanced" label="Advanced" />
      </Row>
      <Stack variant="profile">
        {data.engineReady ? (
          <Select name="quality" options={data.processOptions} disabled={data.preserving} aria-label="Quality" />
        ) : (
          <Text>Native process profile</Text>
        )}
      </Stack>
      <Row variant="process-tabs">
        {data.processTabs.map((item, i) => (
          <Button key={i} variant={item.variant} item={item.title} onClick={() => onAction("setProcessTab")} aria-pressed={item.active}>{item.title}</Button>
        ))}
      </Row>
      <Stack variant="process-fields">
        {data.qualityTab && (
          <>
            <Heading level={3}>Layer height</Heading>
            <Field label="Layer height" variant="setting" unit="mm"><Input name="layerHeight" type="number" step={0.01} min={0.08} max={0.28} disabled={data.preserving} /></Field>
          </>
        )}
        {data.strengthTab && (
          <>
            <Heading level={3}>Walls</Heading>
            <Field label="Wall loops" variant="setting"><Input name="walls" type="number" min={1} max={20} disabled={data.preserving} /></Field>
            <Heading level={3}>Sparse infill</Heading>
            <Field label="Sparse infill density" variant="setting" unit="%"><Input name="infill" type="number" min={0} max={100} disabled={data.preserving} /></Field>
          </>
        )}
        {data.supportTab && (
          <>
            <Heading level={3}>Support</Heading>
            <Checkbox name="supports" label="Enable support" disabled={data.preserving} />
          </>
        )}
        {data.othersTab && (
          <>
            <Heading level={3}>Bed adhesion</Heading>
            <Field label="Brim type" variant="setting"><Select name="brim" options={data.brimOptions} disabled={data.preserving} /></Field>
          </>
        )}
        {data.processGroups.map((item, i) => (
          <Fragment key={i}>
            <Heading level={3}>{item.title}</Heading>
            {item.fields.map((item, i) => (
              <Field key={i} label={item.label} variant="setting" unit={item.unit}>
                {item.select ? (
                  <Select name={item.name} options={item.options} disabled={data.preserving} />
                ) : (
                  <Input name={item.name} type="number" min={item.min} max={item.max} step={item.step} disabled={data.preserving} placeholder="Profile" />
                )}
              </Field>
            ))}
          </Fragment>
        ))}
        {data.preserving && (
          <Text variant="caption">Using the imported project’s settings and painting. Clear Keep imported settings to use the controls above.</Text>
        )}
        {data.hasObjects && (
          <Button variant="secondary" onClick={() => onAction("saveSettings")}>Apply settings</Button>
        )}
        {data.engineNeeded && (
          <Stack variant="engine">
            <Heading level={3}>One-time slicing engine setup</Heading>
            <Text>{data.engine.message}</Text>
            <Text variant="caption">{data.engineProgress}</Text>
            <Section id="engine-progress" />
            <Text variant="caption">474 MB download · Models are kept while setup runs.</Text>
            {data.owner ? (
              <>
                <Button onClick={() => onAction("installEngine")} disabled={data.engine.busy}>Install slicing engine</Button>
                {data.engine.busy && (
                  <Button variant="secondary" onClick={() => onAction("cancelEngine")} cancel={true}>Cancel setup</Button>
                )}
              </>
            ) : (
              <Text>Ask the Workspace owner to finish engine setup.</Text>
            )}
          </Stack>
        )}
        <Text variant="caption">{data.sliceHint}</Text>
      </Stack>
    </Page>
  )
}
