package helpers

import (
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"

	"gitlab.com/pangestu18/janji-online/chat/constant"
)

type Context echo.Context
type Ctx struct {
	C echo.Context
}

func SetContext(c echo.Context) Context {
	return Context(c)
}

func NewCtx(ctx echo.Context) *Ctx {
	return &Ctx{C: SetContext(ctx)}
}

func GetCtx(ctx Context) *Ctx {
	return &Ctx{C: ctx}
}

func SetMetaData(c echo.Context, res map[string]interface{}, isSetCookie bool) {
	metas := []string{
		constant.CtxUserToken,
		constant.CtxDBComp,
		constant.CtxUserCompany,
	}
	for _, key := range metas {
		if value, ok := res[key]; ok {
			c.Set(key, value)
		}
	}
}

func (ctx *Ctx) DB(conn string) *gorm.DB {
	if conn == "" {
		conn = "central"
	}
	return ctx.C.Get("db_" + conn).(*gorm.DB)
}

func (ctx *Ctx) Tx(conn string) *gorm.DB {
	if conn == "" {
		conn = "central"
	}
	return ctx.C.Get("tx_" + conn).(*gorm.DB)
}

func IsRequiredHeader(c echo.Context, key string) bool {
	rules := c.Get(constant.CtxRequiredHeader).(map[string]string)
	if _, ok := rules[key]; ok {
		return true
	} else {
		return false
	}
}

func (ctx *Ctx) AccessToken() string {
	AccessToken := ""
	CtxAccessToken := ctx.C.Get(constant.CtxAccessToken)
	if CtxAccessToken != nil {
		AccessToken = CtxAccessToken.(string)
	}
	return AccessToken
}

func GetAccessToken(c echo.Context) string {
	access_token := ""
	l := len("bearer")
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	if len(auth) > l+1 && strings.ToLower(auth[:l]) == "bearer" {
		access_token = auth[l+1:]
	}
	return access_token
}

func (ctx *Ctx) UserID() string {
	UserID := ""
	CtxUserID := ctx.C.Get(constant.CtxUserID)
	if CtxUserID != nil {
		UserID = CtxUserID.(string)
	}
	return UserID
}

func (ctx *Ctx) UserLang() string {
	lang := "id"
	ctxLang := ctx.C.Get(constant.CtxLang)
	if ctxLang != nil {
		lang = ctxLang.(string)
	}
	return lang
}

func (ctx *Ctx) UserToken() map[string]interface{} {
	var res map[string]interface{}
	usertoken := ctx.C.Get(constant.CtxUserToken)
	if usertoken != nil {
		res = usertoken.(map[string]interface{})
	}
	return res
}

func (ctx *Ctx) ClientID() string {
	ClientID := ""
	CtxClientID := ctx.C.Get(constant.CtxClientID)
	if CtxClientID != nil {
		ClientID = CtxClientID.(string)
	}
	return ClientID
}

func (ctx *Ctx) ErrorMessage() map[string]interface{} {
	ErrorMessage := map[string]interface{}{}
	CtxErrorMessage := ctx.C.Get(constant.CtxErrorMessage)
	if CtxErrorMessage != nil {
		ErrorMessage = CtxErrorMessage.(map[string]interface{})
	}
	return ErrorMessage
}

func (ctx *Ctx) SetErrorMessage(ErrorMessage map[string]interface{}) {
	ctx.C.Set(constant.CtxErrorMessage, ErrorMessage)
}

func CheckOrigin(r *http.Request) bool {
	ret := true
	if strings.Contains(r.Header.Get("Origin"), r.Host) {
		ret = true
	}
	return ret
}
