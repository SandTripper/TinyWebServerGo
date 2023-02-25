package main

import (
	"TinyWebServerGo/handler"
	"TinyWebServerGo/mframe"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// 配置日志文件切割
func configLocalFileSystemLogger(logPath string, logFileName string, rotationTime time.Duration, leastFile uint) {
	baseLogPath := path.Join(logPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPath+"-%Y-%m-%d.log",
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
		rotatelogs.WithRotationCount(leastFile),   // 保留的日志数量
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	log.SetOutput(writer)
}

// 配置日志系统
func configLogger() {
	configLocalFileSystemLogger("./logs", "TinyWebServerLog", time.Hour*24, 3)

	//配置格式
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funcInfo := runtime.FuncForPC(frame.PC)
			if funcInfo == nil {
				return "error during runtime.FuncForPC"
			}
			fullPath, line := funcInfo.FileLine(frame.PC)
			return fmt.Sprintf(" [%v:%v]", filepath.Base(fullPath), line)
		},
		NoColors: true,
	})
	//日志中显示文件名和行数
	log.SetReportCaller(true)
}

func main() {
	//配置日志系统
	configLogger()

	engine := mframe.NewEngine()

	engine.GET("/", handler.ShowIndexPage)
	engine.GET("/test1/:mao/end", handler.TestDynamicRouting)
	engine.GET("/test2/middle/*star", handler.TestDynamicRouting)
	engine.GET("/index", handler.ShowIndexPage)

	engine.Use(handler.RecordAccessLog)

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

	fmt.Print("server start at port 8888\n")
	engine.Run(":8888")
}
