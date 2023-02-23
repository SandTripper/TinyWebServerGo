package mframe

import "net/http"

// 定义处理函数类
type HandlerFunc func(*Context)

// Engine实现ServeHTTP接口
type Engine struct {
	router *router
}

func NewEngine() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

// 添加method为GET的对应请求路径的处理函数
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.router.addRoute("GET", pattern, handler)
}

// 添加method为POST的对应请求路径的处理函数
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.router.addRoute("POST", pattern, handler)
}

// 框架的启动函数
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	c := newContext(writer, req) //将信息封装在context内
	engine.router.handle(c)      //执行路由匹配
}
