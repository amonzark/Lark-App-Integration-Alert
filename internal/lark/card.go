package lark

import (
	"encoding/json"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg/model"
)

// TODO: Add logging here
type cardBuilder struct{}

func newCardBuilder() *cardBuilder {
	return &cardBuilder{}
}

func (l *cardBuilder) Build(alert *model.WebhookAlert) *model.LarkCard {
	return &model.LarkCard{
		Header:   l.buildCardHeader(alert),
		Elements: l.buildCardElements(alert),
	}
}

func (l *cardBuilder) BuildJSON(alert *model.WebhookAlert) (string, error) {
	card := l.Build(alert)
	jsonBytes, err := json.Marshal(card)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (l *cardBuilder) buildCardHeader(alert *model.WebhookAlert) *model.LarkCardHeader {
	return &model.LarkCardHeader{
		Title: &model.LarkCardText{
			Content: alert.Title,
			Tag:     "plain_text",
		},
		Color: alert.Color,
	}
}

func (l *cardBuilder) buildCardElements(alert *model.WebhookAlert) []*model.LarkCardElement {
	return []*model.LarkCardElement{
		{
			Tag: "div",
			Text: &model.LarkCardText{
				Content: alert.Text,
				Tag:     "lark_md",
			},
		},
		{
			Tag: "hr",
		},
		l.buildCardActions(alert),
	}
}

func (l *cardBuilder) buildCardActions(alert *model.WebhookAlert) *model.LarkCardElement {
	cardActions := make([]*model.LarkCardElementAction, 0)
	for _, action := range alert.Actions {
		if action.URL != "" {
			cardActions = append(cardActions, &model.LarkCardElementAction{
				Tag:  "button",
				Type: "default",
				Text: &model.LarkCardText{
					Content: action.Text,
					Tag:     "plain_text",
				},
				MultiURL: &model.LarkCardMultiURL{
					URL: action.URL,
				},
			})
		}
	}
	// Dropdown for silence duration (triggers silence immediately on selection)
	cardActions = append(cardActions, &model.LarkCardElementAction{
		Tag: "select_static",
		Placeholder: &model.LarkCardText{
			Tag:     "plain_text",
			Content: "Select duration",
		},
		Options: []*model.LarkCardSelectOption{
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "30 Minutes"}, Value: "30m"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "1 hour"}, Value: "1h"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "3 hours"}, Value: "3h"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "6 hours"}, Value: "6h"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "12 hours"}, Value: "12h"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "1 day"}, Value: "1d"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "3 days"}, Value: "3d"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "1 week"}, Value: "1w"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "3 weeks"}, Value: "3w"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "1 month"}, Value: "1M"},
			{Text: &model.LarkCardText{Tag: "plain_text", Content: "1 year"}, Value: "1Y"},
		},
		Value: map[string]string{
			"alert_id": alert.CallbackID,
		},
	})
	// Silence button (always uses default duration)
	if alert.CallbackID != "" {
		cardActions = append(cardActions, &model.LarkCardElementAction{
			Tag:  "button",
			Type: "danger",
			Text: &model.LarkCardText{
				Tag:     "plain_text",
				Content: "Silence",
			},
			Value: map[string]string{
				"alert_id": alert.CallbackID,
			},
		})
	}

	return &model.LarkCardElement{
		Tag:     "action",
		Actions: cardActions,
	}
}
