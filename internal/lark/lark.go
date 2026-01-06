package lark

import (
	"log/slog"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg"
)

var client *lark.Client

type Lark struct {
	client *lark.Client

	cardBuilder *cardBuilder
	repository  pkg.Repository

	logger *slog.Logger
}

var _ pkg.Notifier = (*Lark)(nil)

func New(
	appId string,
	appSecret string,

	repository pkg.Repository,

	logger *slog.Logger,
) *Lark {
	if client == nil {
		client = lark.NewClient(
			appId,
			appSecret,
			lark.WithEnableTokenCache(true),
		)
	}

	return &Lark{
		client,

		newCardBuilder(),
		repository,

		logger,
	}
}
