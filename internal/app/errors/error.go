package errors

import (
	"rk-api/pkg/logger"

	"github.com/nicksnyder/go-i18n/i18n"
	"go.uber.org/zap"
)

type ExtraFields map[string]interface{}

// Error 带有错误码 与 错误信息的错误类
// type Error interface {
// 	error
// 	ErrCode() int
// 	TMessage(T i18n.TranslateFunc) string
// 	Track(where, detail string) Error
// 	Log(T i18n.TranslateFunc) Error
// }

type Error struct {
	Code    int    `json:"code,omitempty"` // The http status code
	Message string `json:"message"`        // Message to be display to the end user without debugging information
	Detail  string `json:"detail"`         // Internal error string to help the developer
	Where   string `json:"-"`              // The function where it happened in the form of Struct.Func
	Fields  ExtraFields
}

func (e *Error) Error() string {
	if e.Message == "" {
		return GetErrorMsg(e.Code)
	}
	if e.Where != "" {
		e.Message = e.Where + ":" + e.Message
	}
	if e.Detail != "" {
		e.Message = e.Message + "," + e.Detail
	}
	return e.Message
}

func (e *Error) TMessage(t ...i18n.TranslateFunc) string {
	if e.Code != 0 && e.Code != ERROR {
		e.Message = GetErrorMsg(e.Code)
	}
	if len(t) > 0 {
		var T = t[0]
		if T != nil {
			if e.Fields != nil {
				e.Message = T(e.Message, e.Fields)
			}
			e.Message = T(e.Message)
		}
	}
	return e.Message
}

func (e *Error) ErrCode() int {
	return e.Code
}

func (e *Error) Track(where, detail string) *Error {
	e.Where = where
	e.Detail = detail
	return e
}

func (e *Error) Log(t ...i18n.TranslateFunc) *Error {
	if len(t) != 0 { //没有被处理过，并且有
		e.TMessage(t[0]) //根据语言处理
	}

	fields := []zap.Field{}
	for key, value := range e.Fields {
		fields = append(fields, zap.Any(key, value))
	}
	fields = append(fields, zap.Int("code", e.Code))
	fields = append(fields, zap.String("err", e.Error()))
	// logger.GetLogger().Info("Error", fields...)

	logger.ZError("Server Resp Error", fields...)

	return e
}

func (e *Error) WithFields(fields ExtraFields) *Error {
	if len(e.Fields) == 0 {
		e.Fields = make(ExtraFields)
	}
	for key, value := range fields {
		e.Fields[key] = value
	}
	return e
}

//////////////////////////////////////////////////////////////////////////////////////////////

func With(message string) *Error {
	e := &Error{}
	e.Message = message
	e.Code = ERROR
	return e
}

func WithFields(message string, fields ExtraFields) *Error {
	e := &Error{}
	e.Message = message
	e.Fields = fields
	return e
}

func WithError(err error) *Error {
	e := &Error{}
	if err != nil {
		e.Message = err.Error()
	}
	if appError, ok := err.(*Error); ok {
		e.Code = appError.Code
	} else {
		e.Code = ERROR
	}

	return e
}

func WithCode(code int) *Error {
	e := &Error{}
	e.Code = code
	return e
}

func WithCodeFields(code int, fields ExtraFields) *Error {
	e := &Error{}
	e.Code = code
	e.Fields = fields
	return e
}

func TruncateError(err error, maxSize int) string {
	errorMessage := err.Error()
	if len(errorMessage) > maxSize {
		return errorMessage[:maxSize]
	}
	return errorMessage
}
