package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"url-shortner/log"
	"url-shortner/model"
	"url-shortner/repository"
)

func SaveURL(linkStore *repository.Link, redis *redis.Client) func(c echo.Context) error {
	return func(c echo.Context) error {
		link := &model.Link{}
		err := c.Bind(link)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "can not decode the body as json")
		}

		log.Debug("Get long url with this value ", link.URL)
		if !link.URLValidate() {
			log.Debug("The long url is not valid ")

			return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
		}

		err = linkStore.Insert(link)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "can not insert to the database")
		}

		err = link.MakeShortURL()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "some error in database occur")
		}

		err = redis.Set(c.Request().Context(), link.ShortURL, link.URL, 1*time.Hour).Err()
		if err != nil {
			log.Errorf("can not insert in redis the err is : %s", err)
		}

		return c.JSON(http.StatusOK, link)
	}

}

func Redirect(linkStore *repository.Link, redis *redis.Client) func(c echo.Context) error {
	return func(c echo.Context) error {
		shortURL := c.Param("shortURL")
		link := model.Link{ShortURL: shortURL}
		log.Debug("Get short url with this value ", shortURL)
		if !link.ShortURLValidate() {
			log.Debug("the short url is not found ", shortURL)

			return echo.NewHTTPError(http.StatusBadRequest, "the short url is not valid")
		}

		url, err := redis.Get(c.Request().Context(), shortURL).Result()
		if err != nil {
			log.Errorf("can not retrieve from redis err is : %s", err)
			id, err := link.ShortURLToID()
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "the short url is not found")
			}

			link = linkStore.Get(id)
			log.Debug("find the long url and redirect ", link.URL)

			return c.Redirect(http.StatusFound, link.URL)
		}

		return c.Redirect(http.StatusFound, url)
	}
}
