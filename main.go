package main

import (
	"TinyWebServerGo/handler"
	"TinyWebServerGo/mframe"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Print("server start at port 8888\n")
	engine := mframe.NewEngine()
	engine.GET("/", handler.ShowIndexPage)
	engine.GET("/login", handler.ShowLoginPage)
	engine.GET("/register", handler.ShowRegisterPage)
	engine.GET("/index", handler.ShowIndexPage)
	engine.GET("/welcome", handler.ShowWelcomePage)
	engine.POST("/login", handler.CheckLoginReq)
	engine.POST("/register", handler.CheckRegisterReq)
	engine.POST("/logout", handler.Logout)
	engine.Run(":8888")

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
