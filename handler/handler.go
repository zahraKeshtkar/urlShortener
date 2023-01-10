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
	hash := model.MakeShortUrl(url)
	longUrl, ok := db[hash]
	if longUrl == url && ok {
		return c.JSON(http.StatusOK, hash)
	}

	url = strings.Replace(url, "www.", "", 1)
	link := model.NewLink(url)

	db[link.Url] = link.Hash
	return c.JSON(http.StatusCreated, hash)

}

func Redirect(c echo.Context) error {
	hash := c.Param("hash")

	if hash != "" {
		longUrl, ok := db[hash]

		if ok {
			echo.NewHTTPError(http.StatusBadRequest, "short url is not valid")
		} else {
			c.Redirect(http.StatusFound, longUrl)
		}
	}
	return echo.NewHTTPError(http.StatusBadRequest, "short url is not valid")

}
