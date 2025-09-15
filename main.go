//nolint:gocritic
package main

import (
	"context"
	"frank/app/client/bothub"
	"frank/app/client/yandex"
	"frank/app/service/act"
	"frank/app/service/knowledge"
	"frank/app/service/prompt_manager"
	"frank/app/service/reason"
	"frank/app/service/scheduler"
	"frank/app/service/secret"
	"frank/app/service/telegram_bot"
	"frank/app/service/telegram_reply"
	"frank/pkg/config"
	"frank/pkg/database"
	"frank/pkg/migration"
	"frank/pkg/tlog"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
	_ "go.uber.org/automaxprocs"
)

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exitChan := make(chan struct{})

	di := do.New()
	do.ProvideValue(di, appCtx)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	do.ProvideValue(di, cfg)

	if err = tlog.Init(cfg); err != nil {
		log.Fatalf("logging init failed: %v", err)
	}
	slog.ErrorContext(appCtx, "Service restarted")

	dbConnStr := "postgres://" + cfg.DB.User + ":" + cfg.DB.Pass + "@" + cfg.DB.Host + "/" + cfg.DB.Database + "?sslmode=disable&pool_max_conns=30&pool_min_conns=5&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m&pool_health_check_period=1m&connect_timeout=10"

	dbConf, err := pgxpool.ParseConfig(dbConnStr)
	if err != nil {
		log.Fatalf("pgxpool.ParseConfig() failed: %v", err)
	}

	dbConf.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "30000",
		"idle_in_transaction_session_timeout": "60000",
	}

	dbConn, err := pgxpool.NewWithConfig(appCtx, dbConf)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	if err = database.InitSchema(appCtx, dbConn); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	do.ProvideValue(di, dbConn)

	queries := database.New(dbConn)
	do.ProvideValue(di, queries)

	if err = migration.Migrate(appCtx, di); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	telegramBot, err := bot.New(cfg.Telegram.Token)
	if err != nil {
		log.Fatalf("failed to create telegram bot: %v", err)
	}
	do.ProvideValue(di, telegramBot)

	do.Provide(di, bothub.NewClient)
	do.Provide(di, yandex.NewClient)
	do.Provide(di, secret.New)
	do.Provide(di, knowledge.New)
	do.Provide(di, prompt_manager.New)
	do.Provide(di, telegram_bot.New)
	do.Provide(di, telegram_reply.New)
	do.Provide(di, reason.New)
	do.Provide(di, act.New)
	do.Provide(di, scheduler.New)

	go telegramBot.Start(appCtx)
	defer telegramBot.Close(appCtx)

	go do.MustInvoke[*telegram_bot.Service](di).Run(appCtx)

	do.MustInvoke[*reason.Service](di).SetActor(do.MustInvoke[*act.Service](di))
	do.MustInvoke[*scheduler.Service](di).SetActor(do.MustInvoke[*act.Service](di))

	if err = do.MustInvoke[*scheduler.Service](di).Start(); err != nil {
		log.Fatalf("failed to start scheduler: %v", err)
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Info("Shutting down server...")

		close(exitChan)
	}()

	log.Info("Server started")

	<-exitChan
	cancel()

	log.Info("Waiting for services to finish...")
	_ = di.Shutdown()
}
