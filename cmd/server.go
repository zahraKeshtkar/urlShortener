package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"url-shortner/config"
	database "url-shortner/db"
	"url-shortner/handler"
	"url-shortner/log"
	"url-shortner/repository"
	workerpool "url-shortner/worker"
)

func RegisterServer(root *cobra.Command, cfg config.Config) {
	var port int
	command := &cobra.Command{
		Use:   "server",
		Short: "Run server",
		RunE:  runServer}
	command.Flags().IntVar(&port, "port", cfg.HttpHandler.Port, "port for server")
	root.AddCommand(command)
}

func runServer(cmd *cobra.Command, args []string) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}

	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(cfg.Log.Level)
	redis, err := database.NewRedisConnection(
		cfg.Redis.Host,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.Port,
		time.Duration(cfg.Redis.RetryTimeout)*time.Second,
		cfg.Redis.Retry,
	)
	defer database.Disconnect(redis)
	if err != nil {
		return err
	}

	db, err := database.NewMySQLConnection(cfg.Database.Host,
		cfg.Database.Retry,
		time.Duration(cfg.Database.RetryTimeout)*time.Second,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port)
	if err != nil {
		return err
	}

	defer database.Close(db)
	linkStore := &repository.Link{
		DB: db,
	}
	worker, err := workerpool.NewWorkerpool(cfg.HttpHandler.Workers)
	if err != nil {
		return err
	}

	worker.Run()
	defer worker.Close()
	e := echo.New()
	e.POST("/new", handler.SaveURL(linkStore, redis, worker))
	e.GET("/:shortURL", handler.Redirect(linkStore, redis))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			log.WithFields(logrus.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")

			return nil
		},
	}))
	shutdownCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	serverChannel := make(chan error)
	go func() {
		if err = e.Start(":" + strconv.Itoa(port)); err != nil {
			log.Errorf("Starting server failed: %s", err)
			serverChannel <- err
		}
	}()
	if err != nil {
		return err
	}

	select {
	case <-shutdownCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = e.Shutdown(ctx); err != nil {
			log.Errorf("Shutting down has error: %s", err)

			return err
		}
	case err = <-serverChannel:
		close(serverChannel)

		return err
	}

	return err
}
