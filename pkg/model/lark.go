package model

type LarkCard struct {
	Header   *LarkCardHeader    `json:"header"`
	Elements []*LarkCardElement `json:"elements"`
}

type LarkCardText struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type LarkCardMultiURL struct {
	URL        string `json:"url"`
	PcURL      string `json:"pc_url"`
	AndroidURL string `json:"android_url"`
	IosURL     string `json:"ios_url"`
}

type LarkCardHeader struct {
	Title *LarkCardText `json:"title"`
	Color string        `json:"template"`
}

type LarkCardElement struct {
	Tag     string                   `json:"tag"`
	Text    *LarkCardText            `json:"text,omitempty"`
	Actions []*LarkCardElementAction `json:"actions,omitempty"`
}

type LarkCardElementAction struct {
	Tag         string                  `json:"tag"`
	Type        string                  `json:"type"`
	Text        *LarkCardText           `json:"text"`
	MultiURL    *LarkCardMultiURL       `json:"multi_url"`
	Options     []*LarkCardSelectOption `json:"options"`
	Value       interface{}             `json:"value"`
	Placeholder *LarkCardText           `json:"placeholder"`
}

type LarkCardSilenceButtonValue struct {
	AlertID         string `json:"alert_id"`
	UserID          string `json:"user_id"`
	SilenceDuration string `json:"silence_duration"`
}

type LarkCardSelectOption struct {
	Text  *LarkCardText `json:"text"`
	Value string        `json:"value"`
}
