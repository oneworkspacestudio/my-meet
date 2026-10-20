// Generated from screens/Meet.aui; edit the AUI source.
import { Badge, Button, Card, Divider, Field, Heading, Input, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Meet({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      {data.home && (
        <Row variant="meet-home">
          <Stack variant="hero">
            <Badge>ONE WORKSPACE</Badge>
            <Heading level={1}>Bring everyone together.</Heading>
            <Text>Video meetings, shared screens and conversations. Hosted by you.</Text>
            <Row>
              <Button onClick={() => onAction("new")}>New meeting</Button>
              <Input name="joinCode" placeholder="Enter a meeting code or link" />
              <Button variant="secondary" onClick={() => onAction("open")}>Join</Button>
            </Row>
            <Card>
              <Heading level={2}>Plan a meeting</Heading>
              <Field label="Meeting title"><Input name="title" placeholder="Team catch-up" /></Field>
              <Field label="Date and time"><Input name="starts" type="datetime-local" /></Field>
              <Button variant="secondary" onClick={() => onAction("schedule")}>Create meeting link</Button>
            </Card>
          </Stack>
          <Stack variant="meeting-list">
            <Heading level={2}>Your meetings</Heading>
            {data.noMeetings && (
              <Text>Your recent and scheduled meetings appear here.</Text>
            )}
            {data.meetings.map((item, i) => (
              <Card key={i} variant="meeting-card">
                <Text>{item.title}</Text>
                <Text variant="caption">{item.when}</Text>
                <Row>
                  <Button variant="secondary" onClick={() => onAction("meeting")} item={item.code}>{item.label}</Button>
                  <Button variant="secondary" onClick={() => onAction("invite")} item={item.code}>Copy link</Button>
                </Row>
              </Card>
            ))}
          </Stack>
        </Row>
      )}
      {data.prejoin && (
        <Row variant="prejoin">
          <Stack variant="preview">
            <Section id="preview" />
            <Row>
              <Button variant="secondary" onClick={() => onAction("preview")}>{data.previewLabel}</Button>
              <Button variant={data.micVariant} onClick={() => onAction("mic")}>{data.micLabel}</Button>
              <Button variant={data.cameraVariant} onClick={() => onAction("camera")}>{data.cameraLabel}</Button>
            </Row>
          </Stack>
          <Stack variant="join-panel">
            <Heading level={1}>{data.meeting.title}</Heading>
            <Text>{data.joinStatus}</Text>
            <Field label="Microphone"><Select name="micDevice" options={data.microphones} /></Field>
            <Field label="Camera"><Select name="cameraDevice" options={data.cameras} /></Field>
            <Button onClick={() => onAction("join")} disabled={data.waiting}>{data.joinLabel}</Button>
            <Button variant="secondary" onClick={() => onAction("home")}>Back to meetings</Button>
            <Button variant="secondary" onClick={() => onAction("copy")}>Copy meeting link</Button>
            <Text variant="caption">MyAccounts identifies participants. The host approves people before they enter.</Text>
          </Stack>
        </Row>
      )}
      {data.inCall && (
        <Stack variant="meeting">
          <Row variant="meeting-heading">
            <Heading level={2}>{data.meeting.title}</Heading>
            <Badge tone={data.connectionTone}>{data.connection}</Badge>
            <Text variant="caption">{data.duration}</Text>
            <Spacer />
            <Button variant="secondary" onClick={() => onAction("copy")}>Meeting link</Button>
          </Row>
          <Row variant="meeting-content">
            <Stack variant="stage">
              <Section id="videos" />
              {data.reaction && (
                <Text variant="reaction">{data.reaction}</Text>
              )}
            </Stack>
            {data.panelOpen && (
              <Card variant="meeting-panel">
                <Row>
                  <Heading level={2}>{data.panelTitle}</Heading>
                  <Spacer />
                  <Button variant="icon" onClick={() => onAction("closePanel")} aria-label="Close panel">×</Button>
                </Row>
                {data.chatOpen ? (
                  <>
                    <Stack variant="messages">
                      {data.messages.map((item, i) => (
                        <Stack key={i} variant="message">
                          <Text variant="caption">{item.name}</Text>
                          <Text>{item.body}</Text>
                        </Stack>
                      ))}
                    </Stack>
                    <Row>
                      <Input name="message" placeholder="Send a message to everyone" onClick={() => onAction("send")} />
                      <Button onClick={() => onAction("send")}>Send</Button>
                    </Row>
                  </>
                ) : (
                  <>
                    {data.participants.map((item, i) => (
                      <Row key={i}>
                        <Text>{item.label}</Text>
                        {data.host && (
                          <>
                            <Button variant="secondary" onClick={() => onAction("mutePerson")} item={item.id}>Mute</Button>
                            <Button variant="secondary" onClick={() => onAction("remove")} item={item.id}>Remove</Button>
                          </>
                        )}
                      </Row>
                    ))}
                    {data.host && (
                      <>
                        <Divider />
                        <Heading level={2}>Waiting to join</Heading>
                        {data.waitingMembers.map((item, i) => (
                          <Row key={i}>
                            <Text>{item.name}</Text>
                            <Button onClick={() => onAction("admit")} item={item.id}>Admit</Button>
                            <Button variant="secondary" onClick={() => onAction("deny")} item={item.id}>Decline</Button>
                          </Row>
                        ))}
                        <Button variant="secondary" onClick={() => onAction("lock")}>{data.lockLabel}</Button>
                        <Button variant="danger" onClick={() => onAction("end")}>End for everyone</Button>
                      </>
                    )}
                  </>
                )}
              </Card>
            )}
          </Row>
          <Row variant="call-controls">
            <Button variant={data.micVariant} onClick={() => onAction("mic")}>{data.micLabel}</Button>
            <Button variant={data.cameraVariant} onClick={() => onAction("camera")}>{data.cameraLabel}</Button>
            <Button variant="secondary" onClick={() => onAction("screen")}>{data.screenLabel}</Button>
            <Button variant={data.handVariant} onClick={() => onAction("hand")}>Raise hand</Button>
            <Button variant="secondary" onClick={() => onAction("reaction")}>👏</Button>
            <Button variant="secondary" onClick={() => onAction("chat")}>Chat</Button>
            <Button variant="secondary" onClick={() => onAction("people")}>{data.peopleLabel}</Button>
            <Button variant="secondary" onClick={() => onAction("record")}>{data.recordLabel}</Button>
            <Button variant="danger" onClick={() => onAction("leave")}>Leave</Button>
          </Row>
        </Stack>
      )}
      {data.showNetwork && (
        <Card variant="network">
          <Heading level={2}>Meeting connection</Heading>
          <Text>{data.networkNotice}</Text>
          <Text>{data.networkPorts}</Text>
          <Text variant="caption">Local meetings work without a public media port. Remote connections depend on your network setup.</Text>
        </Card>
      )}
    </Page>
  )
}
