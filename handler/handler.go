package handler

import (
	"TinyWebServerGo/mframe"
	"TinyWebServerGo/sessionmanager"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Context = mframe.Context

var globalSessions *sessionmanager.SessionManager

var globalDb *sql.DB

func init() {
	//实例化会话管理器
	provider := sessionmanager.NewMemoryProvider()
	globalSessions = sessionmanager.NewSessionManager("sessionId", 3600, 3600, 60, provider)
	var err error
	globalDb, err = sql.Open("mysql", "root:@/tiny_web_server_go?charset=utf8")
	checkError(err)
}

// 显示首页
func ShowIndexPage(c *Context) {
	err := c.HTMLF(http.StatusOK, "root/index.html")
	checkServerUnavailableErr(c, err)
}

// 显示登录界面
func ShowLoginPage(c *Context) {
	if _, err := globalSessions.Get(c.Req, "username"); err == nil { //已经登录，重定向至欢迎界面
		c.SetHeader("Location", "/welcome")
		c.Status(302)
		return
	}
	err := c.HTMLFT(http.StatusOK, "root/login.html", "")
	checkServerUnavailableErr(c, err)
}

// 验证登录
func CheckLoginReq(c *Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	fmt.Printf("username: %v, password: %v\n", username, password)
	isAccept, err := checkLogin(username, password)
	if checkServerUnavailableErr(c, err) { //读取数据时发生错误
		return
	}
	if isAccept {
		data := make(map[string]string)
		data["username"] = username
		globalSessions.Create(&c.Writer, c.Req, data)
		c.SetHeader("Location", "/welcome")
		c.Status(302)
	} else {
		err := c.HTMLFT(http.StatusOK, "root/login.html", "用户名或密码错误")
		checkServerUnavailableErr(c, err)
	}
}

// 显示注册界面
func ShowRegisterPage(c *Context) {
	err := c.HTMLFT(http.StatusOK, "root/register.html", "")
	checkServerUnavailableErr(c, err)
}

// 处理注册逻辑
func CheckRegisterReq(c *Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	fmt.Printf("username: %v, password: %v\n", username, password)
	isAccept, err := doRegister(username, password)
	if checkServerUnavailableErr(c, err) { //读写数据库时发生错误
		return
	}
	if isAccept { //注册成功，重定向到登录界面
		c.SetHeader("Location", "/login")
		c.Status(302)
	} else {
		err := c.HTMLFT(http.StatusOK, "root/register.html", "该用户名已被注册")
		checkServerUnavailableErr(c, err)
	}
}

// 显示欢迎界面
func ShowWelcomePage(c *Context) {
	if username, err := globalSessions.Get(c.Req, "username"); err == nil { //已经登陆，显示用户名字
		err := c.HTMLFT(http.StatusOK, "root/welcome.html", username)
		checkServerUnavailableErr(c, err)
	} else { //未登录，重定向至登录界面
		c.SetHeader("Location", "/login")
		c.Status(302)
	}
}

// 登出逻辑
func Logout(c *Context) {
	globalSessions.Destroy(&c.Writer, c.Req)
	c.SetHeader("Location", "/index")
	c.Status(302)
}

// 从数据库中比对用户名和密码
func checkLogin(username string, password string) (bool, error) {
	err := globalDb.QueryRow("SELECT * FROM user_tb WHERE username = ? AND password = ?", username, password).Scan(&username, &password)
	switch {
	case err == sql.ErrNoRows: //不存在结果，即用户名或密码错误
		return false, nil
	case err != nil: //查询出现错误
		return false, err
	default:
		return true, nil
	}
}

// 检查并尝试注册，返回注册结果
func doRegister(username string, password string) (bool, error) {
	err := globalDb.QueryRow("SELECT username FROM user_tb WHERE username = ?", username).Scan(&username)
	switch {
	case err == sql.ErrNoRows: //不存在结果，即用户名或密码错误
		break
	case err != nil: //查询出现错误
		return false, err
	default:
		return false, nil
	}
	stmt, err := globalDb.Prepare("INSERT INTO user_tb VALUES (?,?)")
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(username, password) //执行插入
	if err != nil {
		return false, err
	}
	return true, nil
}

// 检查是否出现服务器错误，若出错，向浏览器返回503，并返回true
func checkServerUnavailableErr(c *Context, err error) bool {
	if err != nil {
		log.Fatal(err)
		serviceUnavailable(c) //向客户端返回服务器错误
		return true
	}
	return false
}

func serviceUnavailable(c *Context) {
	c.String(http.StatusServiceUnavailable, "Service Unavailable")
}

// 如果出现错误，panic
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
