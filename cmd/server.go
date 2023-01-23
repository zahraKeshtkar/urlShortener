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
	"url-shortner/router"
	"url-shortner/store"
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
		log.Errorf("Starting server failed: %s", err)

		return err
	}

	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(cfg.Log.Level)
	db, err := database.NewConnection(cfg.Database.Host,
		cfg.Database.Retry,
		10*time.Second,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DB,
		cfg.Database.Port)
	if err != nil {
		return err
	}

	serverRouter := router.New()
	linkStore := store.NewLinkStore(db)
	h := handler.NewHandler(linkStore)
	serverRouter.POST("/new", h.SaveURL)
	serverRouter.GET("/:shortURL", h.Redirect)
	serverRouter.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
	go func() {
		if err = serverRouter.Start(":" + strconv.Itoa(port)); err != nil {
			log.Errorf("Starting server failed: %s", err)
		}
	}()
	if err != nil {
		return err
	}

	<-shutdownCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = serverRouter.Shutdown(ctx); err != nil {
		log.Errorf("Shutting down has error: %s", err)

		return err
	}

	return err
}
