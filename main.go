package main

import (
	common "gin-crm/app/common"
	db "gin-crm/app/database"
	home "gin-crm/app/home"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {

	r := setupRouter()
	r.Run(":8081")
}

func setupRouter() *gin.Engine {

	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(Recover)                                                                                    //500错误
	r.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"status": 1, "msg": "404"}) })                 //404错误
	r.LoadHTMLGlob(exPath + "/templates/*")                                                           //模版目录
	r.Static("/assets", exPath+"/assets")                                                             //静态资源
	r.Use(sessions.Sessions("userData", cookie.NewStore([]byte("pkTq2pjcaOe568I7qpPPPkbOytw8YIcc")))) //session密钥

	{ //登录
		r.GET("/", index)
		r.GET("/login", login)
		r.POST("/login", login)
	}

	{
		api := r.Group("/", home.Authorize())
		api.GET("/api/:name", home.ApiGet)
		api.POST("/api/:name", home.ApiPost)
	}

	return r
}
func index(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("UserName") == nil {
		c.Redirect(302, "/login")
	} else {
		var db = db.GetDB()
		var nick string
		db.Table("crm_user").Select("nick").Where("name = ?", session.Get("UserName")).Take(&nick)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "客户管理",
			"nick":  nick,
		})
	}
}

func login(c *gin.Context) {

	//用户登录
	session := sessions.Default(c)
	session.Delete("UserName")
	session.Delete("UserKey")
	session.Save()

	if c.Request.Method == "POST" {
		var postJson struct {
			User     string `form:"user" json:"user" binding:"required" `
			Password string `form:"password" json:"password" binding:"required"`
		}
		var db = db.GetDB()
		var userstate int

		if err := c.ShouldBindJSON(&postJson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 1, "msg": "用户名密码错误！"})
			return
		}

		db.Table("crm_user").Select("state").Where("name = ?", postJson.User).Where("pas = ?", common.GetMD5Hash(postJson.Password)).Take(&userstate)

		if userstate > 0 {
			userKey := time.Now().Unix()
			session.Set("UserName", postJson.User)
			session.Set("UserKey", userKey)
			session.Save()
			db.Table("crm_user").Where("name = ?", postJson.User).Update("key", userKey)
			db.Table("crm_login_log").Create(&map[string]interface{}{"login_time": userKey, "login_ip": c.ClientIP(), "user": postJson.User})
			c.JSON(200, gin.H{"status": 0, "msg": "登录成功！"})
		} else {
			db.Table("crm_login_log").Create(&map[string]interface{}{"login_time": time.Now().Unix(), "login_ip": c.ClientIP(), "user": postJson.User, "pas": postJson.Password, "state": 1})
			c.JSON(200, gin.H{"status": 1, "msg": "用户名密码错误！"})
		}
		return

	}

	c.HTML(http.StatusOK, "login.html", gin.H{"title": "登录"})
}
func Recover(c *gin.Context) {
	//全局错误中间件
	defer func() {
		if recover() != nil {
			//debug.PrintStack()
			c.JSON(500, gin.H{"status": 1, "msg": "500"})
		}
	}()
	c.Next()
}
