package repository

import (
	"log/slog"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg"
)

func New(
	address, password string, db int,
	logger *slog.Logger,
) pkg.Repository {
	return newRedis(address, password, db, logger)
}
