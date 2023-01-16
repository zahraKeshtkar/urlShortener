package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"url-shortner/log"
	"url-shortner/model"
)

var db = make(map[string]string)

func SaveURL(c echo.Context) error {
	link := &model.Link{}
	err := c.Bind(link)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "can not decode the body as json")
	}

	log.Debug("Get long url with this value ", link.URL)
	if !link.IsURLValid() {
		log.Debug("The long url is not valid ")

		return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
	}

	ok := link.MakeShortURL(db)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "We cannot make short url retry later")
	}

	return c.JSON(http.StatusOK, link)
}

func Redirect(c echo.Context) error {
	shortURL := c.Param("shortURL")
	log.Debug("Get short url with this value ", shortURL)
	longURL, ok := model.FindShortURL(shortURL, db)
	if !ok {
		log.Debug("the short url is not found ", shortURL)

		return echo.NewHTTPError(http.StatusNotFound, "the short url is not found")
	}

	log.Debug("find the long url and redirect ", longURL)

	return c.Redirect(http.StatusFound, longURL)

}
