package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"url-shortner/log"
	"url-shortner/model"
)

var db = make(map[string]string)

func SaveURL(c echo.Context) error {
	body := make(map[string]string)
	err := json.NewDecoder(c.Request().Body).Decode(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "can not decode the body as json")
	}

	URL := body["url"]
	log.Debug("Get long url with this value ", URL)
	if !model.IsURLValid(URL) {
		log.Debug("The long url is not valid ")

		return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
	}

	link := model.NewLink(len(db)+1, URL)
	log.Debug("The short url will be ", link.ShortURL)
	db[link.ShortURL] = link.URL
	log.Debug("Saved in the database with success status")

	return c.JSONPretty(http.StatusOK, link, "	")
}

func Redirect(c echo.Context) error {
	shortURL := c.Param("shortURL")
	log.Debug("Get short url with this value ", shortURL)
	longURL, ok := model.FindShortURL(shortURL, db)
	if !ok {
		log.Debug("the short url is not found ", shortURL)

		return echo.NewHTTPError(http.StatusNotFound, "the short url is not found")
	} else {
		log.Debug("find the long url and redirect ", longURL)

		return c.Redirect(http.StatusFound, longURL)
	}
}
