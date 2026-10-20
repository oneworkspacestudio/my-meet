// Generated from screens/Shell.aui; edit the AUI source.
import { Alert, Button, Card, Heading, Icon, Image, Link, Page, Row, Section, Spacer, Spinner, Text } from "@/components/ui"

export function Shell({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="topbar">
        <Link href="#"><Image src={data.logo} alt={data.name} width={38} height={38} /></Link>
        <Heading level={2}>{data.name}</Heading>
        <Spacer />
        <Text variant="caption">{data.user.name}</Text>
        <Link href="#settings" aria-label="Settings"><Icon name="Settings2" /></Link>
        <Button variant="icon" onClick={() => onAction("theme")} aria-label="Change appearance"><Icon name="SunMoon" /></Button>
        {data.workspace && (
          <Button variant="secondary" onClick={() => onAction("apps")} data-workspace-apps-trigger={true} aria-label="Open Workspace apps">
            <Icon name="Layers" />
            <Text variant="desktop-label">Workspace</Text>
          </Button>
        )}
        <Button variant="secondary" onClick={() => onAction("logout")} aria-label="Sign out">
          <Icon name="LogOut" />
          <Text variant="desktop-label">Sign out</Text>
        </Button>
      </Row>
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
      {data.appsOpen && (
        <Card variant="dialog" aria-label="Workspace apps">
          <Row>
            <Heading level={2}>Your Workspace apps</Heading>
            <Spacer />
            <Button variant="secondary" onClick={() => onAction("closeApps")}>Close</Button>
          </Row>
          {data.appLinks.map((item, i) => (
            <Link key={i} href={item.href}>{item.title}</Link>
          ))}
        </Card>
      )}
    </Page>
  )
}
