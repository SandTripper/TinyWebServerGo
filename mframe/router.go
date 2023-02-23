package mframe

import (
	"log"
	"net/http"
)

type router struct {
	handlers map[string]HandlerFunc
}

// 创建新路由
func newRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc)}
}

// 添加路由规则
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// 执行路由规则对应的处理函数
func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		err := c.HTMLF(http.StatusOK, "root/404.html")
		if err != nil { //读取页面时出错
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path) //返回默认404界面
		}
	}
}
