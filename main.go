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
	engine.GET("/test1/:mao/end", handler.TestDynamicRouting)
	engine.GET("/test2/middle/*star", handler.TestDynamicRouting)
	engine.GET("/index", handler.ShowIndexPage)

	groupAuth := engine.Group("/auth")
	{
		groupAuth.GET("/login", handler.ShowLoginPage)
		groupAuth.GET("/register", handler.ShowRegisterPage)
		groupAuth.POST("/login", handler.CheckLoginReq)
		groupAuth.POST("/register", handler.CheckRegisterReq)
	}

	groupUser := engine.Group("/user")
	{
		groupUser.GET("/welcome", handler.ShowWelcomePage)
		groupUser.POST("/logout", handler.Logout)
	}

	engine.Run(":8888")
}
