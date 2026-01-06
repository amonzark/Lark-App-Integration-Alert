package model

type WebhookAlert struct {
	Color      string               `json:"color"`
	CallbackID string               `json:"callback_id"`
	Footer     string               `json:"footer"`
	MrkdwnIn   []string             `json:"mrkdwn_in"`
	TitleLink  string               `json:"title_link"`
	Text       string               `json:"text"`
	Title      string               `json:"title"`
	Fallback   string               `json:"fallback"`
	Actions    []WebhookAlertAction `json:"actions"`
}

type WebhookAlertAction struct {
	Text string `json:"text"`
	Type string `json:"type"`
	URL  string `json:"url,omitempty"`
}

type Webhook struct {
	Alerts   []WebhookAlert `json:"attachments"`
	Channel  string         `json:"channel"`
	Username string         `json:"username"`
}

type URLVerificationRequest struct {
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
	Token     string `json:"token"`
}

type CardActionPayload struct {
	Schema string `json:"schema"`
	Header struct {
		EventType  string `json:"event_type"`
		TenantKey  string `json:"tenant_key"`
		EventID    string `json:"event_id"`
		CreateTime string `json:"create_time"`
		AppID      string `json:"app_id"`
		Token      string `json:"token"`
	} `json:"header"`
	Event struct {
		Host    string `json:"host"`
		Context struct {
			OpenMessageID string `json:"open_message_id"`
			OpenChatID    string `json:"open_chat_id"`
		} `json:"context"`
		Action struct {
			Tag   string `json:"tag"`
			Value struct {
				AlertID string `json:"alert_id"`
			} `json:"value"`
			Option string `json:"option"` // add this field for select option
		} `json:"action"`
		Operator struct {
			TenantKey string `json:"tenant_key"`
			UserID    string `json:"user_id"`
			OpenID    string `json:"open_id"`
			UnionID   string `json:"union_id"`
		} `json:"operator"`
		Token string `json:"token"` // extra field from event
	} `json:"event"`
}

// toast message for lark
type Toast struct {
	Type    string `json:"type,omitempty"`    // info, success, error, warning
	Content string `json:"content,omitempty"` // fallback content
}

type CallbackResponse struct {
	Toast *Toast `json:"toast,omitempty"`
}
