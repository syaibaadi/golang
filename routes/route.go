package routes

import (
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	c "gitlab.com/pangestu18/janji-online/chat/controllers"
	"gitlab.com/pangestu18/janji-online/chat/helpers"
	"gitlab.com/pangestu18/janji-online/chat/middlewares"
)

func Init(db *gorm.DB, ch *redis.Client) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = middlewares.ErrorHandler
	e.Use(middlewares.TransactionHandler(db))
	e.Use(middlewares.CacheHandler(ch))

	helpers.CreateRoute()
	helpers.AddRoute("connect", helpers.RouteMap{
		IsPublished: true,
		ACL:         "",
		Type:        "",
		Route:       "/connect/:room",
		Method:      "GET",
		Callback: func() {
			e.GET("/connect/:room", c.ServeWs)
		},
		Headers: map[string]string{},
	},
	)
	helpers.AddRoute("socket_connect", helpers.RouteMap{
		IsPublished: true,
		ACL:         "",
		Type:        "",
		Route:       "/socket/connect/:room",
		Method:      "GET",
		Callback: func() {
			e.GET("/socket/connect/:room", c.ServeWs)
		},
		Headers: map[string]string{},
	},
	)

	helpers.AddRoute("sandbox", helpers.RouteMap{
		IsPublished: true,
		ACL:         "",
		Type:        "",
		Route:       "/sandbox/:room",
		Method:      "GET",
		Callback: func() {
			e.GET("/sandbox/:room", c.Test)
		},
		Headers: map[string]string{},
	},
	)

	helpers.AddRoute("/", helpers.RouteMap{
		IsPublished: true,
		ACL:         "",
		Type:        "",
		Route:       "/",
		Method:      "GET",
		Callback: func() {
			e.GET("/", c.Ping)
		},
		Headers: map[string]string{},
	},
	)

	for _, m := range helpers.GetRouteMap() {
		m.Callback()
	}
	return e
}
