package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"url-shortner/model"
)

var db = make(map[string]string)

func SaveUrl(c echo.Context) error {
	url := c.FormValue("url")

	if !model.IsUrlValid(url) {
		return echo.NewHTTPError(http.StatusBadRequest, "This is not a url at all")
	}
	if !model.IsLinkExits(url) {
		return echo.NewHTTPError(http.StatusNotFound, "This link is not found")
	}

	ShortUrl := model.MakeShortUrl(url)
	longUrl, ok := db[ShortUrl]

	if longUrl == url && ok {
		return c.JSON(http.StatusOK, ShortUrl)
	}

	url = strings.Replace(url, "www.", "", 1)
	link := model.NewLink(url)

	db[link.ShortUrl] = link.Url

	return c.JSONPretty(http.StatusCreated, link, "	")

}

func Redirect(c echo.Context) error {
	ShortUrl := c.Param("hash")
	if ShortUrl != "" {
		longUrl, ok := db[ShortUrl]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "short url is not valid")
		} else {
			return c.Redirect(http.StatusFound, longUrl)
		}
	}
	return echo.NewHTTPError(http.StatusBadRequest, "short url is not valid")

}
