package main

import (
	"TinyWebServerGo/handler"
	"TinyWebServerGo/mframe"
	"fmt"
)

func main() {
	fmt.Print("server start at port 8888\n")
	engine := mframe.NewEngine()
	engine.GET("/", handler.ShowIndexPage)
	engine.GET("/login", handler.ShowLoginPage)
	engine.GET("/register", handler.ShowRegisterPage)
	engine.GET("/index", handler.ShowIndexPage)
	engine.GET("/welcome", handler.ShowWelcomePage)
	engine.GET("/test1/:mao/end", handler.TestDynamicRouting)
	engine.GET("/test2/middle/*star", handler.TestDynamicRouting)
	engine.POST("/login", handler.CheckLoginReq)
	engine.POST("/register", handler.CheckRegisterReq)
	engine.POST("/logout", handler.Logout)
	engine.Run(":8888")
}
