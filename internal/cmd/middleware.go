package cmd

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"alpaca-app/internal/consts"
)

// responseMiddleware is ghttp.MiddlewareHandlerResponse, except only user-safe messages
// (validation and consts.CodeUser) reach the client; everything else is logged and replaced
// with a generic Hebrew message.
func responseMiddleware(r *ghttp.Request) {
	r.Middleware.Next()
	err := r.GetError()
	if r.Response.BytesWritten() > 0 || err == nil && r.Response.BufferLength() > 0 {
		return
	}
	// On panic GoFrame writes the panic text into the buffer; never send it.
	r.Response.ClearBuffer()

	code, msg := gerror.Code(err), "OK"
	switch {
	case err != nil && code == consts.CodeUser:
		msg = gerror.Cause(err).Error() // the Hebrew message, without wrap context
	case err != nil && code == gcode.CodeValidationFailed:
		msg = err.Error()
	case err != nil:
		g.Log().Errorf(r.Context(), "%+v", err)
		code, msg = gcode.CodeInternalError, consts.MsgInternal
	default:
		code = gcode.CodeOK
	}
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{Code: code.Code(), Message: msg, Data: r.GetHandlerResponse()})
}
