package handler

import (
	"TinyWebServerGo/mframe"
	"TinyWebServerGo/sessionmanager"
	"database/sql"
	"fmt"
	"net/http"
	"plugin"
	"runtime"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	log "github.com/sirupsen/logrus"
)

type Context = mframe.Context

var GlobalSessions *sessionmanager.SessionManager

var GlobalDb *sql.DB
var GlobalDbLock sync.Mutex

var lastDouYinReqTime = time.Now().UnixMilli() //上次服务器向抖音发出请求的时间
var intervalBetweenDouYinReq int64 = 10000     //请求的间隔时间

func init() {
	//实例化会话管理器
	provider := sessionmanager.NewMemoryProvider()
	GlobalSessions = sessionmanager.NewSessionManager("sessionId", 3600, 3600, 60, provider)
	var err error
	GlobalDb, err = sql.Open("mysql", "root:@/tiny_web_server_go?charset=utf8")
	checkError(err)
}

// 显示首页
func ShowIndexPage(c *Context) {
	err := c.HTMLF(http.StatusOK, "root/index.html")
	checkServerUnavailableErr(c, err)
}

// 显示登录界面
func ShowLoginPage(c *Context) {
	if c.PermissionLevel >= 1 { //已经登录，重定向至欢迎界面
		c.SetHeader("Location", "/user/welcome")
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

	isAccept, err := checkLogin(username, password)
	if checkServerUnavailableErr(c, err) { //读取数据时发生错误
		return
	}
	if isAccept {
		data := make(map[string]interface{})

		permissionLevel, err := getUserPerFromSql(username)
		if checkServerUnavailableErr(c, err) {
			return
		}

		data["username"] = username
		data["permission_level"] = permissionLevel

		GlobalSessions.Create(&c.Writer, c.Req, data)
		c.SetHeader("Location", "/user/welcome")
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
		c.SetHeader("Location", "/auth/login")
		c.Status(302)
	} else {
		err := c.HTMLFT(http.StatusOK, "root/register.html", "该用户名已被注册")
		checkServerUnavailableErr(c, err)
	}
}

// 显示欢迎界面
func ShowWelcomePage(c *Context) {
	if c.PermissionLevel >= 1 { //已经登陆，显示用户名字
		err := c.HTMLFT(http.StatusOK, "root/welcome.html", c.Username)
		checkServerUnavailableErr(c, err)
	} else { //未登录，重定向至登录界面
		c.SetHeader("Location", "/auth/login")
		c.Status(302)
	}
}

// 登出逻辑
func Logout(c *Context) {
	GlobalSessions.Destroy(&c.Writer, c.Req) //销毁会话
	c.SetHeader("Location", "/index")
	c.Status(302)
}

func ShowRobotsTxt(c *Context) {
	c.String(200, "User-agent: *\nDisallow: /")
}

func ShowFavicon(c *Context) {
	c.File(200, "root/favicon.ico")
}

// 抖音直链提取api
func DouYinUrlHandler(c *Context) {

	p, err := plugin.Open("./lib/ParseDouYinUrl.so")
	checkServerUnavailableErr(c, err)
	symbol, err := p.Lookup("GetDouYinRealUrl")
	checkServerUnavailableErr(c, err)
	GetDouYinRealUrl := symbol.(func(string) (string, error))

	for time.Now().UnixMilli()-lastDouYinReqTime < intervalBetweenDouYinReq { //冷却时间未达到，让出时间片
		runtime.Gosched()
	}
	lastDouYinReqTime = time.Now().UnixMilli()

	ans, err := GetDouYinRealUrl(c.PostForm("url"))
	data := make(map[string]interface{})
	if err == nil {
		data["status"] = "ok"
		data["url"] = ans
		c.JSON(200, data)
	} else {
		data["status"] = "invalid url"
		data["url"] = ""
		c.JSON(200, data)
	}

}

// 测试动态路由
func TestDynamicRouting(c *Context) {
	var text string
	for k, v := range c.Params {
		text += k + ":" + v + "\n"
	}
	c.String(http.StatusOK, text)
}

// 从数据库中比对用户名和密码
func checkLogin(username string, password string) (bool, error) {
	GlobalDbLock.Lock()
	defer GlobalDbLock.Unlock()
	err := GlobalDb.QueryRow("SELECT * FROM user_tb WHERE username = ? AND password = ?", username, password).Scan(&username, &password)
	switch {
	case err == sql.ErrNoRows: //不存在结果，即用户名或密码错误
		return false, nil
	case err != nil: //查询出现错误
		return false, err
	default:
		return true, nil
	}
}

func getUserPerFromSql(username string) (int, error) {
	GlobalDbLock.Lock()
	defer GlobalDbLock.Unlock()
	var str string
	err := GlobalDb.QueryRow("SELECT permission_level FROM user_permission_level_tb WHERE username = ?", username).Scan(&str)
	if err == nil {
		permissionLevel, err := strconv.Atoi(str)
		if err == nil {
			return permissionLevel, nil
		}
		return 0, err
	}
	return 0, err
}

// 检查并尝试注册，返回注册结果
func doRegister(username string, password string) (bool, error) {
	GlobalDbLock.Lock()
	defer GlobalDbLock.Unlock()

	err := GlobalDb.QueryRow("SELECT username FROM user_tb WHERE username = ?", username).Scan(&username)
	switch {
	case err == sql.ErrNoRows: //不存在结果，即用户名或密码错误
		break
	case err != nil: //查询出现错误
		return false, err
	default:
		return false, nil
	}
	stmt, err := GlobalDb.Prepare("INSERT INTO user_tb VALUES (?,?)")
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(username, password) //执行插入
	if err != nil {
		return false, err
	}

	//初始化权限
	stmt, err = GlobalDb.Prepare("INSERT INTO user_permission_level_tb VALUES (?,?)")
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(username, 1) //执行插入
	if err != nil {
		return false, err
	}

	return true, nil
}

// 检查是否出现错误，若出错，向浏览器返回503，并返回true
func checkServerUnavailableErr(c *Context, err error) bool {
	if err != nil {
		log.Error(err)
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
		log.Panic(err)
	}
}
