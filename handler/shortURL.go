package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"url-shortner/config"
	"url-shortner/log"
	"url-shortner/model"
	"url-shortner/repository"
	"url-shortner/worker"
)

func SaveURL(linkStore *repository.Link, redis *redis.Client,
	workerPool workerpool.WorkerPool) func(c echo.Context) error {
	return func(c echo.Context) error {
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

		workerPool.AddTask(func() error {
			return linkStore.Insert(link)
		})
		err = link.MakeShortURL()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "A database error has occurred")
		}

		err = redis.Set(c.Request().Context(), link.ShortURL, link.URL,
			time.Duration(config.GetRedis().TTL)*time.Hour).Err()
		if err != nil {
			log.Errorf("Can not insert in redis the err is : %s", err)
		} else {
			log.Infof("The value was successfully inserted in redis: %s", link.ShortURL)
		}

		return c.JSON(http.StatusOK, link)
	}

}

func Redirect(linkStore *repository.Link, redisClient *redis.Client) func(c echo.Context) error {
	return func(c echo.Context) error {
		shortURL := c.Param("shortURL")
		link := model.Link{ShortURL: shortURL}
		log.Debug("Get short url with this value ", shortURL)
		if !link.Validate() {
			log.Debug("The short url is not found ", shortURL)

			return echo.NewHTTPError(http.StatusBadRequest, "the short url is not valid")
		}

		url, err := redisClient.Get(c.Request().Context(), shortURL).Result()
		if errors.Is(err, redis.Nil) == true {
			log.Infof("The short url is not in redis : %s", err)
		} else if err != nil {
			log.Errorf("A redis error has occurred: %s", err)
		} else {
			return c.Redirect(http.StatusFound, url)
		}

		id, err := link.ShortURLToID()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "the short url is not found")
		}

		link, err = linkStore.Get(id)
		if err != nil {
			log.Errorf("A database error has occurred: %s", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "A database error has occurred")
		}

		log.Debugf("Find the long url and redirect %s", link.URL)

		return c.Redirect(http.StatusFound, url)
	}
}
