package helpers

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

func SuccessResponse(c echo.Context, statusCode int, res interface{}) {
	c.JSON(statusCode, res)
}

func ReplaceDetailError(detail map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for i, v := range detail {
		i = strings.Replace(i, "AuthorizationRequired.", "", 1)
		i = strings.Replace(i, "SlugRequired.", "", 1)
		i = strings.Replace(i, "ClientRequired.", "", 1)
		i = strings.Replace(i, "Slug", "slug", 1)
		res[i] = v
	}
	return res
}

func Response(c echo.Context, code int, res map[string]interface{}) error {
	if res["error"] != nil {
		e := res["error"].(map[string]interface{})
		if e["detail"] != nil {
			e["detail"] = ReplaceDetailError(e["detail"].(map[string]interface{}))
		}
		if e["code"] != nil {
			code = e["code"].(int)
			NewCtx(c).SetErrorMessage(res)
			return echo.NewHTTPError(code, e["message"].(string))
		}
	}
	return c.JSON(code, res)
}

func ResponseInternalError(c echo.Context, e error) error {
	return echo.NewHTTPError(500, e.Error())
}

func NotFoundMessage(object, key, id string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    404,
			"message": object + " with " + key + " = '" + id + "' is not found.",
		},
	}
}

func InternalErrorMessage(message string) map[string]interface{} {
	if message == "" {
		message = "Internal Error"
	}
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    500,
			"message": message,
		},
	}
}

func DeletedMessage(object, key, id string) map[string]interface{} {
	return map[string]interface{}{
		"code":    200,
		"message": object + " with " + key + " = '" + id + "' has been deleted.",
	}
}

func GeneralErrorMessage(code int, message string, detail map[string]interface{}) map[string]interface{} {
	err := map[string]interface{}{}
	err["code"] = code
	err["message"] = message
	if len(detail) > 0 {
		err["detail"] = detail
	}
	return map[string]interface{}{"error": err}
}

func ErrorHasAwaitingPaymentInvoice(ctx Context) (map[string]interface{}, error) {
	message := "Anda masih memiliki tagihan yang belum dibayar, silakan selesaikan terlebih dahulu tagihan Anda sebelumnya."
	if GetCtx(ctx).UserLang() == "id" {
		message = "Anda masih memiliki tagihan yang belum dibayar, silakan selesaikan terlebih dahulu tagihan Anda sebelumnya."
	}
	return GeneralErrorMessage(402, message, map[string]interface{}{}), errors.New(message)
}

func ErrorInvoiceAmountIsZero(ctx Context) (map[string]interface{}, error) {
	message := "Invoice amount cannot be zero"
	if GetCtx(ctx).UserLang() == "id" {
		message = "Nilai tagihan tidak boleh nol"
	}
	return GeneralErrorMessage(400, message, map[string]interface{}{}), errors.New(message)
}

func ErrorEmailAlreadyBeenTaken(ctx Context) (map[string]interface{}, error) {
	message := "Email already been taken"
	if GetCtx(ctx).UserLang() == "id" {
		message = "Email sudah digunakan"
	}
	return GeneralErrorMessage(400, message, map[string]interface{}{}), errors.New(message)
}

func ErrorPhoneAlreadyBeenTaken(ctx Context) (map[string]interface{}, error) {
	message := "Phone already been taken"
	if GetCtx(ctx).UserLang() == "id" {
		message = "Nomor telepon sudah digunakan"
	}
	return GeneralErrorMessage(400, message, map[string]interface{}{}), errors.New(message)
}
