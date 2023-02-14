package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"url-shortner/config"
	"url-shortner/log"
	"url-shortner/metric"
	"url-shortner/model"
	"url-shortner/repository"
	"url-shortner/tracing"
	"url-shortner/worker"
)

func SaveURL(linkStore *repository.Link, redis *redis.Client,
	workerPool *workerpool.Workerpool) func(c echo.Context) error {
	return func(c echo.Context) error {
		now := time.Now()
		metric.MuxMetric.RequestCounter.With(prometheus.Labels{"url": c.Request().URL.String()}).Inc()
		ctx, span := tracing.DefaultTracer.Start(c.Request().Context(), "handler.create")
		defer span.End()
		span.SetAttributes(attribute.String("url", c.Request().URL.String()))
		span.SetAttributes(attribute.String("method", c.Request().Method))
		link := &model.Link{}
		err := c.Bind(link)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return echo.NewHTTPError(http.StatusBadRequest, "can not decode the body as json")
		}

		log.Debug("Get long url with this value ", link.URL)
		if !link.Validate() {
			log.Debug("The long url is not valid ")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
		}

		workerPool.AddTask(func() error {
			return linkStore.Insert(link)
		})
		err = link.MakeShortURL()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return echo.NewHTTPError(http.StatusInternalServerError, "A database error has occurred")
		}

		err = redis.Set(ctx, link.ShortURL, link.URL,
			time.Duration(config.GetRedis().TTL)*time.Hour).Err()
		if err != nil {
			log.Errorf("Can not insert in redis the err is : %s", err)
		} else {
			log.Infof("The value was successfully inserted in redis: %s", link.ShortURL)
		}

		metric.MuxMetric.RequestDurations.(prometheus.ExemplarObserver).ObserveWithExemplar(
			time.Since(now).Seconds(), prometheus.Labels{"url": c.Request().URL.String()},
		)

		return c.JSON(http.StatusOK, link)
	}

}

func Redirect(linkStore *repository.Link, redisClient *redis.Client) func(c echo.Context) error {
	return func(c echo.Context) error {
		now := time.Now()
		metric.MuxMetric.RequestCounter.With(prometheus.Labels{"url": c.Request().URL.String()}).Inc()
		ctx, span := tracing.DefaultTracer.Start(c.Request().Context(), "handler.redirect")
		defer span.End()
		span.SetAttributes(attribute.String("url", c.Request().URL.String()))
		span.SetAttributes(attribute.String("method", c.Request().Method))
		shortURL := c.Param("shortURL")
		link := model.Link{ShortURL: shortURL}
		log.Debug("Get short url with this value ", shortURL)
		if !link.Validate() {
			log.Debug("The short url is not found ", shortURL)

			return echo.NewHTTPError(http.StatusBadRequest, "the short url is not valid")
		}

		url, err := redisClient.Get(ctx, shortURL).Result()
		if errors.Is(err, redis.Nil) == true {
			log.Infof("The short url is not in redis : %s", err)
		} else if err != nil {
			log.Errorf("A redis error has occurred: %s", err)
		} else {
			return c.Redirect(http.StatusFound, url)
		}

		id, err := link.ShortURLToID()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return echo.NewHTTPError(http.StatusInternalServerError, "the short url is not found")
		}

		link, err = linkStore.Get(id)
		if err != nil {
			log.Errorf("A database error has occurred: %s", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return echo.NewHTTPError(http.StatusInternalServerError, "A database error has occurred")
		}

		log.Debugf("Find the long url and redirect %s", link.URL)
		metric.MuxMetric.RequestDurations.(prometheus.ExemplarObserver).ObserveWithExemplar(
			time.Since(now).Seconds(), prometheus.Labels{"url": c.Request().URL.String()},
		)

		return c.Redirect(http.StatusFound, url)
	}
}
