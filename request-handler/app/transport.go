package app

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Router() *echo.Echo {
	app := NewApplication()

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/*", app.HandleProjectRequest)

	return e
}

func (app *Application) HandleProjectRequest(c echo.Context) error {
	var (
		contentType string
		objKey      string
	)

	host := c.Request().Host
	projectID := strings.Split(host, ".")[0]
	requestedPath := c.Param("*")

	if requestedPath == "" {
		objKey = projectID + "/dist/" + "index.html"
		contentType = "text/html"
	} else {
		segments := strings.Split(requestedPath, "/")
		objKey = projectID + "/dist/" + requestedPath
		fileType := strings.Split(segments[len(segments)-1], ".")[1]
		if fileType == "html" {
			contentType = "text/html"
		} else if fileType == "css" {
			contentType = "text/css"
		} else if fileType == "js" {
			contentType = "application/javascript"
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
	}

	content, err := app.getFileContent(objKey)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	c.Response().Header().Add("Content-type", contentType)
	c.Response().Writer.Write(content)
	return nil
}
