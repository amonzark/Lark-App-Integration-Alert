package pkg

import (
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg/model"
)

type Notifier interface {
	NotifyAlerts(alert model.Webhook) error
	SendResponseCreatedSilence(message_id string, chat_id string, text string) error
	GetUserInfo(openID string) (*larkcontact.User, error)
}
