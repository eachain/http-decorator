package decorator

import (
	"net/http"
)

// Context 简单封装，只是简化写法
type Context struct {
	w http.ResponseWriter
	r *http.Request
}

// Handler 实现Context decorator
type Handler interface {
	ServeCtx(*Context)
}

// HandlerFunc 类似于http.HandlerFunc
// 实现 Handler
type HandlerFunc func(*Context)

// ServeCtx 实现 Handler
func (f HandlerFunc) ServeCtx(c *Context) {
	f(c)
}

// WithContext 将 Handler 转为 http.Handler
func WithContext(handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeCtx(&Context{w: w, r: r})
	})
}

// Parse 只支持解析int, int64, float64, string, []string, time.Time类型字段
func (c *Context) Parse(param interface{}) error {
	return parseParam(param, c.r)
}

func (c *Context) RenderOK(data interface{}) {
	renderOK(c.w, data)
}

func (c *Context) RenderErr(code int, msg string) {
	renderErr(c.w, code, msg)
}
