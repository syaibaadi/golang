package middlewares

import (
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/thoas/bokchoy"

	"gitlab.com/pangestu18/janji-online/chat/constant"
)

func CacheHandler(ch *redis.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constant.CtxChc, ch)
			n := next(c)
			return n
		}
	}
}

func QueueHandler(q *bokchoy.Bokchoy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constant.CtxQueue, q)
			n := next(c)
			return n
		}
	}
}
