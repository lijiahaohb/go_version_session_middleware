package main

import (
	"fmt"
	"net/http"
	"session/gin_session"

	"github.com/gin-gonic/gin"
)

// 用户信息
type UserInfo struct {
	UserName string `form:"username"`
	Password string `form:"password"`
}

// 校验用户是否登录的中间件
// 思路: 从上下文中取出session data，从session data中取出isLogin
func AuthMiddleware(c *gin.Context) {
	// 1. 先从上下文中取出session data
	fmt.Println("in Auth")
	tmpSD, _ := c.Get(gin_session.SessionContextName)
	sd := tmpSD.(gin_session.SessionData)

	// 2. 从session data 中取出isLogin
	fmt.Printf("%#v\n", sd)
	value, err := sd.Get("isLogin")
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	fmt.Println(value)
	isLogin, ok := value.(bool)
	if !ok {
		fmt.Println("!ok")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	fmt.Println(isLogin)
	if !isLogin {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.Next()
}

func loginHandler(c *gin.Context) {
	// 判断请求方法
	if c.Request.Method == "POST" {
		toPath := c.DefaultQuery("next", "/index")
		var u UserInfo
		err := c.ShouldBind(&u)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"err": "用户名或密码不能为空",
			})
			return
		}

		// 解析成功 验证账号密码是否正确
		if u.UserName == "lijiahao" && u.Password == "cv123" {
			// 验证成功之后，在当前session data中设置islogin为true
			// 1. 首先从上下文中获取session data
			tmpSD, ok := c.Get(gin_session.SessionContextName)
			if !ok {
				panic("session middleware")
			}
			sd := tmpSD.(gin_session.SessionData)
			// 给session data设置isLogin = true
			sd.Set("isLogin", true)
			// 调用Save，存储到数据库
			sd.Save()
			// 跳转到index界面
			c.Redirect(http.StatusMovedPermanently, toPath)
		} else {
			// 验证错误重新返回登录界面
			c.HTML(http.StatusOK, "login.html", gin.H{
				"err": "用户名或密码错误",
			})
			return
		}
	} else {
		c.HTML(http.StatusOK, "login.html", nil)
	}
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func homeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

func vipHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "vip.html", nil)
}