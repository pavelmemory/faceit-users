package main

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap/zapcore"

	"github.com/pavelmemory/faceit-users/internal"
	"github.com/pavelmemory/faceit-users/internal/config"
	"github.com/pavelmemory/faceit-users/internal/logging"
	"github.com/pavelmemory/faceit-users/internal/storage"
	"github.com/pavelmemory/faceit-users/internal/user"
	"github.com/pavelmemory/faceit-users/internal/webhttp"
)

func run([]string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = interrupt(ctx)

	settings, err := config.NewEnvSettings(os.Getenv("ENV_PREFIX"))
	if err != nil {
		logger := logging.NewZapLogger(zapcore.InfoLevel.String())
		logger.WithError(err).Error("settings initialization")
		logger.Sync()
		return err
	}

	logger := logging.NewZapLogger(settings.LogLevel())
	defer logger.Sync()

	logger.WithString("version", internal.Version).
		WithString("commit_sha", internal.CommitSHA).
		WithString("build_timestamp", internal.BuildTimestamp).
		Info("executable build info")

	pgstorage, err := storage.NewPostgres(settings.StorageAddr(), settings.StoragePwd())
	if err != nil {
		logger.WithError(err).Error("postgres connection establishment")
		return err
	}
	defer pgstorage.Close()

	usersService := user.NewService(pgstorage)
	usersHandler := webhttp.NewUsersHandler(usersService)

	router := webhttp.NewRouter(logger)
	usersHandler.Register(router)
	srv := webhttp.NewServer(router)

	return webhttp.Serve(ctx, logger, srv, settings.HTTPPort())
}

// interrupt listens for SIGINT and cancels context.
func interrupt(ctx context.Context) context.Context {
	cctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		cancel()
	}()

	return cctx
}
