package main

import (
	"github.com/labstack/echo/v4"

	"url-shortner/handler"
)

func main() {
	e := echo.New()
	e.POST("/new", handler.SaveUrl)
	e.GET("/:hash", handler.Redirect)
	e.Start(":8080")
}
