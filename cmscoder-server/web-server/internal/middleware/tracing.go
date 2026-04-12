package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// Tracing injects trace ID into request context and response header.
func Tracing(r *ghttp.Request) {
	traceId := r.GetHeader("X-Trace-Id")
	if traceId == "" {
		traceId = guid.S()
	}
	r.Response.Header().Set("X-Trace-Id", traceId)
	r.SetCtxVar("traceId", traceId)
	r.Middleware.Next()
}
