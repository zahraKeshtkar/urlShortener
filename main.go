package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"url-shortner/handler"
	"url-shortner/log"
)

func main() {
	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.TraceLevel)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	e := echo.New()
	e.POST("/new", handler.SaveURL)
	e.GET("/:shortURL", handler.Redirect)
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
	go func() {
		if err := e.Start(":8080"); err != nil {
			log.Fatalf("Starting server failed: %s", err)
		}
	}()
	<-shutdownCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Shutting down has error: %s", err)
	}
}
