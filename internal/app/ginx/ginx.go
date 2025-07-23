package ginx

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/odeke-em/go-uuid"
)

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"error,omitempty"` // 当Msg为空时，"error"键将不会出现在JSON中
	Data interface{} `json:"data"`
}

func Mine(c *gin.Context) uint {
	user, exists := c.Get("userID")
	if exists && user != nil {
		userID, _ := strconv.ParseUint(user.(string), 10, 64)
		return uint(userID)
	}
	// xuser := c.GetHeader("X-User")
	// if len(xuser) <= 0 {
	// 	xuser = "0" // test的用户ID
	// }
	// 对字符串的用户ID进行转换
	// userID, _ := strconv.ParseUint(xuser, 10, 64)
	// return uint(userID)
	return 0
}

func MustMine(c *gin.Context) uint {
	user, exists := c.Get("userID")
	if exists && user != nil {
		userID, _ := strconv.ParseUint(user.(string), 10, 64)
		return uint(userID)
	}
	return 0
}

// Param returns the value of the URL param
func ParseParamID(c *gin.Context, key string) uint64 {
	val := c.Param(key)
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// Parse body json data to struct
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.WithError(err).Track("ginx.ParseJSON", fmt.Sprintf("Parse request json failed: %s", err.Error()))
	}
	return nil
}

// Parse query parameter to struct
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.WithError(err).Track("ginx.ParseQuery", fmt.Sprintf("Parse request json failed: %s", err.Error()))
	}
	return nil
}

// Parse body form data to struct
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.WithError(err).Track("ginx.ParseForm", fmt.Sprintf("Parse request json failed: %s", err.Error()))
	}
	return nil
}

// Response data object
func RespSucc(c *gin.Context, v interface{}) {
	ResJSON(c, http.StatusOK, &Resp{
		Code: errors.SUCCESS,
		Data: v,
	})
}

// MarkError 将错误 存入日志
func MarkError(c *gin.Context, err error) {
	if appError, ok := err.(*errors.Error); ok {
		appError.WithFields(errors.ExtraFields{"user": c.GetString("user"), "requestID": c.GetString("requestID")}).Log()
	} else {
		errors.WithError(err).WithFields(errors.ExtraFields{"user": c.GetString("user"), "requestID": c.GetString("requestID")}).Log()
	}

}

func GetPreferredLanguage(c *gin.Context) string {
	acceptLang := c.GetHeader("Accept-Language")
	// if acceptLang == "" {
	// 	acceptLang = c.GetHeader("Accept-Language")
	// }
	if len(acceptLang) > 1 {
		// 通常 Accept-Language 可能包含多个语言，例如 "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7"。
		// 这里简化处理，仅提取首选语言
		return acceptLang
	}
	return config.Get().ServiceSettings.Language // 默认返回英语前缀
}

// 翻译参数错误
func transValidationError(locale string, validationErrors validator.ValidationErrors) string {
	ut := utils.GetTrans(locale) //
	errorMessages := make([]string, 0, len(validationErrors))
	for _, valErr := range validationErrors {
		translated := valErr.Translate(ut)
		errorMessages = append(errorMessages, translated)
	}
	return strings.Join(errorMessages, ",")
}

// 翻译参数错误
func transAppError(locale string, appError *errors.Error) string {
	return appError.TMessage(utils.GetUserTranslations(locale)) //根据语言处理
}

// 翻译第一个参数错误
func transFirstValidationError(locale string, validationErrors validator.ValidationErrors) string {
	ut := utils.GetTrans(locale) //
	if len(validationErrors) > 0 {
		return validationErrors[0].Translate(ut)
	}
	return ""
}

// ResponseError 返回失败
func RespErr(c *gin.Context, err error) {

	if validationErrors, ok := err.(validator.ValidationErrors); ok { //参数校验
		locale := GetPreferredLanguage(c)
		ResJSON(c, http.StatusOK, &Resp{
			Code: errors.ValidationParamError,
			Msg:  transFirstValidationError(locale, validationErrors), //只返回第一个错误参数提示
		})
	} else if appError, ok := err.(*errors.Error); ok { //调用错误
		// if config.Get().ServiceSettings.IsDevelopment() {

		// }

		locale := GetPreferredLanguage(c)
		ResJSON(c, http.StatusOK, &Resp{
			Code: appError.ErrCode(),
			Msg:  transAppError(locale, appError),
		})
	} else { ////其他错误
		ResJSON(c, http.StatusOK, &Resp{
			Code: errors.ERROR,
			Msg:  err.Error(),
		})
	}
	MarkError(c, err)
}

// Response json data with status code
func ResJSON(c *gin.Context, status int, v interface{}) {
	c.JSON(status, v)
}

/*****************************************************************************/

// SetAuthTokens 设置认证 Token 到 HTTP Header 和 Cookie 中
func SetAuthTokens(ctx *gin.Context, accessToken, refreshToken string, tokenExpireTime int) {
	// 设置 HTTP Header
	ctx.Header(constant.ACCESS_TOKEN, accessToken)
	ctx.Header(constant.REFRESH_TOKEN, refreshToken)

	// // 判断是否为 HTTPS
	secure := utils.IsHttps(ctx)
	// 计算 Cookie 的 MaxAge
	maxAge := tokenExpireTime * 3600
	// 设置 Access Token 到 Cookie
	ctx.SetCookie(constant.ACCESS_TOKEN, accessToken, maxAge, "/", "", secure, true)
	// 设置 Refresh Token 到 Cookie
	ctx.SetCookie(constant.REFRESH_TOKEN, refreshToken, maxAge, "/", "", secure, true)

	SetDeviceID(ctx)
}

func SetDeviceID(ctx *gin.Context) string {
	newDeviceId := uuid.New()
	ctx.SetCookie("deviceId", newDeviceId, 3600*24*30, "/", "", false, true) // 有效期30天
	return newDeviceId
}
