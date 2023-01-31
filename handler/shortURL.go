package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"url-shortner/log"
	"url-shortner/model"
	"url-shortner/repository"
)

func SaveURL(linkStore *repository.Link) func(c echo.Context) error {
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

		err = linkStore.Insert(link)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "can not insert to the database")
		}

		err = link.MakeShortURL()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "some error in database occur")
		}

		return c.JSON(http.StatusOK, link)
	}

}

func Redirect(linkStore *repository.Link) func(c echo.Context) error {
	return func(c echo.Context) error {

		shortURL := c.Param("shortURL")
		link := model.Link{ShortURL: shortURL}
		log.Debug("Get short url with this value ", shortURL)
		if !link.Validate() {
			log.Debug("the short url is not found ", shortURL)

			return echo.NewHTTPError(http.StatusBadRequest, "the short url is not valid")
		}

		id, err := link.ShortURLToID()
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "the short url is not found")
		}

		link = linkStore.Get(id)
		log.Debug("find the long url and redirect ", link.URL)

		return c.Redirect(http.StatusFound, link.URL)
	}
}
