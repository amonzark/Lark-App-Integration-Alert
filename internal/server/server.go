package server

import (
	"log/slog"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg"
)

type Server struct {
	port              int
	notifier          pkg.Notifier
	alertmanagerHost  string
	verificationToken string

	logger *slog.Logger
}

func New(
	notifier pkg.Notifier,
	port int,
	alertmanagerHost string,
	verificationToken string,

	logger *slog.Logger,
) *Server {
	return &Server{
		port:              port,
		notifier:          notifier,
		alertmanagerHost:  alertmanagerHost,
		verificationToken: verificationToken,

		logger: logger,
	}
}
