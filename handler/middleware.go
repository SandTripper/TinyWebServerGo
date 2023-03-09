package handler

import log "github.com/sirupsen/logrus"

// 中间件，记录访问日志
func RecordAccessLog(c *Context) {
	log.Infof("(ip:%s) %s URI:(%s)", c.Req.RemoteAddr, c.Req.Method, c.Req.RequestURI)
}

// 中间件，查询权限等级,用户信息
func GetUserData(c *Context) {
	if permissionLevel, err := GlobalSessions.Get(c.Req, "permission_level"); err == nil { //已经登陆，设置权限等级并查询用户名字
		c.PermissionLevel = permissionLevel.(int)
		username, err := GlobalSessions.Get(c.Req, "username")
		if checkServerUnavailableErr(c, err) { //存在permission_level但不存在username，逻辑错误，直接panic
			log.Panic("has permission_level but not username")
		}
		c.Username = username.(string)
	} else { //未登录
		c.PermissionLevel = 0
	}
}
