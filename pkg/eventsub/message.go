package eventsub

type Status string

const (
	StatusConnected      Status = "connected"
	StatusEnabled        Status = "enabled"
	StatusReconnecting   Status = "reconnecting"
	StatusRevoked        Status = "authorization_revoked"
	StatusUserRemoved    Status = "user_removed"
	StatusVersionRemoved Status = "version_removed"
)

type MessageType string

const (
	TypeSessionWelcome   MessageType = "session_welcome"
	TypeSessionKeepalive MessageType = "session_keepalive"
	TypeNotification     MessageType = "notification"
	TypeSessionReconnect MessageType = "session_reconnect"
	TypeRevocation       MessageType = "revocation"
)

type Metadata struct {
	ID        string `json:"message_id"`
	Type      string `json:"message_type"`
	Timestamp string `json:"message_timestamp"`
}

type Session struct {
	ID                      string `json:"id"`
	Status                  string `json:"status"`
	KeepaliveTimeoutSeconds int    `json:"keepalive_timeout_seconds"`
	ReconnectURL            string `json:"reconnect_url"`
	ConnectedAt             string `json:"connected_at"`
}

type Subscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Cost      int               `json:"cost"`
	Condition map[string]string `json:"condition"`
	Transport struct {
		Method    string `json:"method"`
		SessionID string `json:"session_id"`
	} `json:"transport"`
	CreatedAt string `json:"created_at"`
}

type Payload struct {
	Session      *Session      `json:"session"`
	Subscription *Subscription `json:"subscription"`
	Event        *Event        `json:"event"`
}

type Event map[string]interface{}

type WebsocketMessage struct {
	Metadata Metadata `json:"metadata"`
	Payload  Payload  `json:"payload"`
}

type EventHandler func(Subscription, Event)
