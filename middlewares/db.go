package middlewares

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"

	"gitlab.com/pangestu18/janji-online/chat/constant"
)

func TransactionHandler(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constant.CtxDBCentral, db)
			c.Set(constant.CtxTxCentral, db)
			n := next(c)
			return n
		}
	}
}
