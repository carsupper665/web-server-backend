package gin

import (
	"encoding/json"
	"net/http"
)

type H map[string]interface{}

type HandlerFunc func(*Context)

type Engine struct {
	mux *http.ServeMux
}

func Default() *Engine {
	return &Engine{mux: http.NewServeMux()}
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e.mux)
}

func (e *Engine) Use(_ ...HandlerFunc) {}

func (e *Engine) handle(method, path string, h HandlerFunc) {
	e.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		c := &Context{w: w, r: r}
		h(c)
	})
}

func (e *Engine) POST(path string, h HandlerFunc)   { e.handle(http.MethodPost, path, h) }
func (e *Engine) GET(path string, h HandlerFunc)    { e.handle(http.MethodGet, path, h) }
func (e *Engine) DELETE(path string, h HandlerFunc) { e.handle(http.MethodDelete, path, h) }

func (e *Engine) Group(prefix string) *RouterGroup {
	return &RouterGroup{engine: e, prefix: prefix}
}

// RouterGroup represents a group of routes with a common prefix.
type RouterGroup struct {
	engine *Engine
	prefix string
}

func (g *RouterGroup) Use(_ ...HandlerFunc)              {}
func (g *RouterGroup) POST(path string, h HandlerFunc)   { g.engine.POST(g.prefix+path, h) }
func (g *RouterGroup) GET(path string, h HandlerFunc)    { g.engine.GET(g.prefix+path, h) }
func (g *RouterGroup) DELETE(path string, h HandlerFunc) { g.engine.DELETE(g.prefix+path, h) }

// Context carries request-scoped values across the middleware chain.
type Context struct {
	w http.ResponseWriter
	r *http.Request
}

func (c *Context) PostForm(key string) string {
	if err := c.r.ParseForm(); err != nil {
		return ""
	}
	return c.r.FormValue(key)
}

func (c *Context) Param(key string) string {
	// No real routing params in this stub
	return ""
}

func (c *Context) JSON(code int, obj interface{}) {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	json.NewEncoder(c.w).Encode(obj)
}

func (c *Context) String(code int, s string) {
	c.w.WriteHeader(code)
	c.w.Write([]byte(s))
}

func (c *Context) Status(code int) {
	c.w.WriteHeader(code)
}

func (c *Context) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

func (c *Context) AbortWithStatusJSON(code int, obj interface{}) {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	json.NewEncoder(c.w).Encode(obj)
}

func (c *Context) Next() {}
