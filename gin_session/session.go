package gin_session

import "github.com/gin-gonic/gin"

const (
	SessionCookieName  = "session_id" // cookie中session_id对应的key
	SessionContextName = "session"    // session data 在gin上下文中对应的key
)

var (
	// MgrObj 全局的session管理对象
	MgrObj Mgr
)

// 初始化Mgr
func InitMgr(name string, addr string, option ...string) {
	switch name {
	case "memory":
		MgrObj = NewMemory()
	case "redis":
		MgrObj = NewRedisMgr()
	}
	// 初始化mgr
	MgrObj.Init(addr, option...)
}

type SessionData interface {
	GetID() string  // 返回自己的ID
	Get(key string) (value interface{}, err error)
	Set(key string, value interface{}) 
	Del(key string)
	Save()
}

// 不同版本的管理者需要实现的接口
type Mgr interface {
	Init(addr string, option... string)
	GetSessionData(sessionId string) (sd SessionData, err error)
	CreateSession() (sd SessionData)
}

// gin框架中间件
func SessionMiddleware(mgrObj Mgr) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求的cookie中获取SessionId
		SessionID, err := c.Cookie(SessionCookieName)
		var sd SessionData
		if err != nil {
			// 如果没有SessionID， 创建一个Session，并返回对应的SessionData
			sd = mgrObj.CreateSession()
			SessionID = sd.GetID()
		} else {
			// 获取到SessionId之后取出SessionData
			sd, err = mgrObj.GetSessionData(SessionID)
			if err != nil {
				// sessionId有误，获取不到SessionData
				sd = mgrObj.CreateSession()
				SessionID = sd.GetID()
			}
		}
		// 实现让后续所有请求的方法都拿到SessionData
		c.Set(SessionContextName, sd)
		// 回写cookie
		c.SetCookie(SessionCookieName, SessionID, 3600, "/", "127.0.0.1", false, true)
	}
}
