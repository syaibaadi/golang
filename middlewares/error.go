package middlewares

import (
	"log"

	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	he, _ := err.(*echo.HTTPError)
	code := he.Code
	message := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": err.Error(),
		},
	}
	log.Println(message)
	c.JSON(code, message)
}
