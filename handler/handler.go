package handler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"url-shortner/db"
	"url-shortner/log"
	"url-shortner/model"
)

type API struct {
	DB database.DataBase
}

func (api *API) Run() {
	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	e := echo.New()
	e.POST("/new", api.SaveURL)
	e.GET("/:shortURL", api.Redirect)
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

func (api *API) SaveURL(c echo.Context) error {
	link := &model.Link{}
	err := c.Bind(link)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "can not decode the body as json")
	}

	log.Debug("Get long url with this value ", link.URL)
	if !link.Validate() {
		log.Debug("The long url is not valid ")

		return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
	}

	api.DB.InsertLink(link)
	link.MakeShortURL()

	return c.JSON(http.StatusOK, link)
}

func (api *API) Redirect(c echo.Context) error {
	shortURL := c.Param("shortURL")
	link := model.Link{ShortURL: shortURL}
	log.Debug("Get short url with this value ", shortURL)
	if !link.Validate() {
		log.Debug("the short url is not found ", shortURL)

		return echo.NewHTTPError(http.StatusNotFound, "the short url is not found")
	}

	id := link.ShortURLToID()
	link = api.DB.GetLink(id)
	log.Debug("find the long url and redirect ", link.URL)

	return c.Redirect(http.StatusFound, link.URL)

}
