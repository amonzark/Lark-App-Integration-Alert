package main

import (
	"log/slog"
	"os"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/lark"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/repository"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/server"
	
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// TODO: refactor this somewhere
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		panic("REDIS_ADDRESS is required")
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	
	redisRepository := repository.New(
		redisAddress,
		redisPassword,
		0,

		slog.Default(),
	)

	alertmanagerHost := os.Getenv("ALERTMANAGER_HOST")
	if alertmanagerHost == "" {
		panic("ALERTMANAGER_HOST is required")
	}

	verificationToken := os.Getenv("VERIFICATION_TOKEN")
	if verificationToken == "" {
		panic("VERIFICATION_TOKEN is required")
	}
	// TODO: refactor this somewhere
	larkAppID := os.Getenv("LARK_APP_ID")
	if larkAppID == "" {
		panic("LARK_APP_ID is required")
	}
	larkAppSecret := os.Getenv("LARK_APP_SECRET")
	if larkAppSecret == "" {
		panic("lark_app_secret is required")
	}
	larkNotifier := lark.New(
		larkAppID,
		larkAppSecret,

		redisRepository,

		slog.Default(),
	)
	server := server.New(
		larkNotifier,
		8080,
		alertmanagerHost,
		verificationToken,
		slog.Default(),
	)
	server.Start()
}
