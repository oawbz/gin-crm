package home

import (
	"fmt"
	common "gin-crm/app/common"
	db "gin-crm/app/database"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var Db = db.GetDB()

var cstSh, _ = time.LoadLocation("Asia/Shanghai")

func Authorize() gin.HandlerFunc {
	//鉴权中间件
	return func(c *gin.Context) {

		session := sessions.Default(c)
		var user struct {
			Name  string `column:"name"`
			Key   int64  `column:"key" `
			State string `column:"state" `
		}
		Db.Table("crm_user").Where("name = ?", session.Get("UserName")).Where("`key` = ?", session.Get("UserKey")).Take(&user)
		if user.State == "1" && user.Key >= time.Now().Unix()-60*60*24 && session.Get("UserName") != nil {
			c.Next()
		} else {
			// 验证不通过，不再调用后续的函数处理
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": 99, "msg": "重新登录"})
			return
		}
	}
}
func doctortime(c *gin.Context) { //医生日期查询
	s, _ := strconv.ParseInt(c.Query("s"), 10, 64) //预约时间
	s3 := time.Date(time.Unix(s, 0).In(cstSh).Year(), time.Unix(s, 0).In(cstSh).Month(), time.Unix(s, 0).In(cstSh).Day(), 0, 0, 0, 0, time.Unix(s, 0).In(cstSh).Location()).Unix()
	zhou := time.Unix(s3, 0).In(cstSh).Weekday()

	//fmt.Println(s3)

	var kcsl int64
	var sysl int64
	Db.Table("crm_doctor").Select(common.Txt(zhou)).Where("id = ?", c.Query("d")).Where("state = 1").Take(&kcsl)
	Db.Table("crm_customer_reserve").Select("COUNT(`id`)").Where("doctor = ?", c.Query("d")).Where("visit_time >= ?", s3).Where("visit_time <= ?", s3+60*60*24).Take(&sysl)
	data := map[string]interface{}{"syh": kcsl - sysl}
	if c.Query("s") == "" {
		data["syh"] = 0
	}
	c.AsciiJSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "医生排期",
		"data":   data,
	})

}
func automaticallocation(c *gin.Context) { //自动分配
	session := sessions.Default(c)
	usergroups := usergroup(common.Txt(session.Get("UserName")))
	uall := userallow(common.Txt(session.Get("UserName")))

	if uall > 2 {
		c.JSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
		return
	}
	if c.Query("c") == "add" {
		var zdfpcc map[string]int64
		var zs int64
		c.ShouldBindJSON(&zdfpcc)
		for k, v := range zdfpcc {
			if k != "countlist" && k != "doctor" && v != 0 {
				zs = zs + v
			}
		}
		if zdfpcc["countlist"] < zs {
			c.JSON(http.StatusOK, gin.H{"status": 1, "msg": "超过可分配数量"})
			return
		}
		for k, v := range zdfpcc {
			//fmt.Println(zdfpcc["doctor"]) //医生id
			if k != "countlist" && k != "doctor" && v != 0 {
				y1 := strings.Replace(k, "user", "", 1)
				var fenpeilist []int64
				if zdfpcc["doctor"] == 0 {
					Db.Table("crm_customer").Select("id").Where("consultuser = 0").Where("project = ?", usergroups).Limit(int(common.ToInt(v))).Find(&fenpeilist) //列表
				} else {
					Db.Table("crm_customer").Select("id").Where("consultuser = 0").Where("doctor = ?", zdfpcc["doctor"]).Where("project = ?", usergroups).Limit(int(common.ToInt(v))).Find(&fenpeilist) //列表
				}
				Db.Table("crm_customer").Where("consultuser = 0").Where("id IN ?", fenpeilist).Update("consultuser", y1)
				///
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "分配成功"})
		return
	}

	var userlist []map[string]interface{}
	Db.Table("crm_user").Select("id,nick").Where("`state` = 1").Where("`allow`= 4").Where("`group` = ?", usergroups).Find(&userlist) //列表
	var counts int64
	if c.Query("d") != "" {
		Db.Table("crm_customer").Select("id").Where("doctor = ?", c.Query("d")).Where("consultuser = 0").Where("project = ?", usergroups).Count(&counts) //列表
	} else {
		Db.Table("crm_customer").Select("id").Where("consultuser = 0").Where("project = ?", usergroups).Count(&counts) //列表
	}

	if c.Query("s") == "1" {
		datakc := map[string]interface{}{"countlist": counts}
		c.AsciiJSON(http.StatusOK, gin.H{
			"status": 0,
			"msg":    "可分配数量",
			"data":   datakc,
		})
		return
	}

	var body []Body

	for _, v := range userlist {
		body = append(body, Body{
			Type:     "input-number",
			Name:     "user" + common.Txt(v["id"]),
			Label:    common.Txt(v["nick"]),
			Lequired: true,
			Disabled: false,
		})

	}
	body = append(body, Body{
		Type:     "input-number",
		Name:     "countlist",
		Label:    "可分配数量",
		Lequired: true,
		Disabled: true,
		Value:    counts,
	})

	////

	c.AsciiJSON(http.StatusOK, body)

}
func navigation(c *gin.Context) { //系统菜单
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))

	var navigation []navigationMode
	var nav struct {
		Pages struct {
			Children []navigationMode `column:"children" json:"children" type:"string"`
		} `column:"pages" json:"pages" type:"string"`
	}

	navigation = append(navigation, navigationMode{
		Label:     "首页",
		Icon:      "fa fa-home",
		Url:       "/",
		Redirect:  "/",
		SchemaApi: "get:./api/menu?id=home",
	})
	if uall < 4 {
		navigation = append(navigation, navigationMode{
			Label:     "数据录入",
			Icon:      "fa fa-address-card-o",
			Url:       "/DataEntry",
			Redirect:  "/",
			SchemaApi: "get:./api/menu?id=DataEntry",
		})
	}
	if uall < 5 {
		navigation = append(navigation, navigationMode{
			Label:     "电话列表",
			Icon:      "fa fa-address-book",
			Url:       "/PhoneList",
			Redirect:  "/",
			SchemaApi: "get:./api/menu?id=PhoneList",
		})
	}

	navigation = append(navigation, navigationMode{
		Label:     "预约列表",
		Icon:      "fa fa-users",
		Url:       "/AppointmentList",
		Redirect:  "/",
		SchemaApi: "get:./api/menu?id=AppointmentList",
	})

	if uall == 1 || uall == 2 {
		navigation = append(navigation, navigationMode{
			Label:     "数据统计",
			Icon:      "fa fa-bar-chart-o",
			Url:       "/Statistics",
			Redirect:  "/",
			SchemaApi: "get:./api/menu?id=Statistics",
		}, navigationMode{
			Label:     "项目设置",
			Icon:      "fa fa-wrench",
			Url:       "/project",
			Redirect:  "/",
			SchemaApi: "get:./api/menu?id=project",
		})
	}
	navigation = append(navigation, navigationMode{
		Label:     "修改密码",
		Icon:      "fa fa-user-circle-o",
		Url:       "/pas",
		Redirect:  "/",
		SchemaApi: "get:./api/menu?id=pas",
	})
	/*
		, navigationMode{
			Label:     "系统配置",
			Icon:      "fa fa-cog",
			Url:       "/Config",
			Redirect:  "/",
			SchemaApi: "get:./api/menu?id=Config",
		}
	*/
	nav.Pages.Children = navigation

	c.AsciiJSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "ok",
		"data":   nav,
	})
}

func menu(c *gin.Context) { //菜单页面
	name := c.Query("id")
	var menuMode menuMode
	Db.Table("crm_menu").Where("name = ?", name).Where("state = 1").Take(&menuMode)
	if menuMode.Data != "" {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.String(http.StatusOK, common.Txt(menuMode.Data))
	} else {
		c.JSON(http.StatusOK, gin.H{"status": 1, "msg": "系统错误"})
	}
}

func project(c *gin.Context) { //项目管理
	//if c.Request.Method == "POST" { //添加项目
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))

	if c.Query("c") == "add" { //添加项目

		if uall != 1 { //超级管理员可以添加
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "添加失败,没有权限"})
			return
		}

		var projectAdd projectAdd
		c.ShouldBindJSON(&projectAdd)
		projectAdd.Addtime = time.Now().Unix()
		var projectId int64
		Db.Table("crm_project").Select("id").Where("name = ?", projectAdd.Name).Where("state = 1").Find(&projectId) //列表
		if projectId != 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "项目名称重复"})
		} else {
			Db.Table("crm_project").Create(&projectAdd)
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功"})
		}
		return
	}

	if c.Query("c") == "del" { //删除项目
		if uall != 1 { //超级管理员可以删除
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "删除失败,没有权限"})
			return
		}
		c.AsciiJSON(http.StatusOK, del("crm_project", c.Query("id")))
		return
	}
	if c.Query("c") == "renew" { //修改
		if uall > 2 { //超级管理员可以修改
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "修改失败,没有权限"})
			return
		}
		var projectAdd projectAdd
		c.ShouldBindJSON(&projectAdd)
		projectAdd.Addtime = time.Now().Unix()
		Db.Table("crm_project").Select("name").Where("id = ?", c.Query("id")).Where("state = 1").Find(&projectAdd.Name)
		if uall == 1 {
			Db.Table("crm_project").Where("id = ?", c.Query("id")).Updates(projectAdd) //修改全部
		}
		if uall == 2 {
			Db.Table("crm_project").Where("id = ?", c.Query("id")).Updates(projectAdd) //修改医生和平台
		}

		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "修改成功"})
		return
	}
	if c.Query("c") == "list" { //项目列表
		if uall > 2 { //超级管理员可以修改
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		data := xmlist("crm_project", c.Query("page"), c.Query("perPage"), common.Txt(session.Get("UserName")))
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "项目列表"})
		}
		return
	}

}
func platform(c *gin.Context) { //平台管理
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	if uall > 2 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
		return
	}

	if c.Query("c") == "add" { //添加平台
		var platformAdd platformAdd
		c.ShouldBindJSON(&platformAdd)
		platformAdd.Addtime = time.Now().Unix()
		platformAdd.Uid = userid(common.Txt(session.Get("UserName")))
		var projectId int64
		Db.Table("crm_platform").Select("id").Where("name = ?", platformAdd.Name).Find(&projectId)
		if projectId != 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "项目名称重复"})
		} else {
			Db.Table("crm_platform").Create(&platformAdd)
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功"})
		}
		return
	}
	if c.Query("c") == "list" { //平台列表
		data := xmlist("crm_platform", c.Query("page"), c.Query("perPage"), common.Txt(session.Get("UserName")))
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "列表"})
		}
		return
	}
	if c.Query("c") == "del" { //删除平台
		if uall != 1 { //超级管理员可以修改
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		c.AsciiJSON(http.StatusOK, del("crm_platform", c.Query("id")))
		return
	}
}
func user(c *gin.Context) { //用户管理
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	if uall > 2 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
		return
	}
	if c.Query("c") == "add" { //添加用户

		var userAdd userAdd
		c.ShouldBindJSON(&userAdd)

		var projectsId int64
		Db.Table("crm_project").Select("id").Where("id = ?", userAdd.Group).Where("director = ?", userid(common.Txt(session.Get("UserName")))).Where("state = 1").Find(&projectsId)
		if uall == 2 && (userAdd.Allow == 1 || userAdd.Allow == 2) && projectsId == 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}

		userAdd.Addtime = time.Now().Unix()
		userAdd.Pas = common.GetMD5Hash(userAdd.Pas)
		var projectId int64
		Db.Table("crm_user").Select("id").Where("name = ?", userAdd.Name).Where("state = 1").Find(&projectId)
		if projectId != 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "用户名重复"})
		} else {
			Db.Table("crm_user").Create(&userAdd)
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功"})
		}
		return
	}
	if c.Query("c") == "renew" { //修改

		var userrenew userAdd
		c.ShouldBindJSON(&userrenew)

		var projectId int64
		Db.Table("crm_project").Select("id").Where("director = ?", userid(common.Txt(session.Get("UserName")))).Where("id = ?", userrenew.Group).Where("state = 1").Find(&projectId)
		if uall != 1 && (userrenew.Allow == 1 || userrenew.Allow == 2) && projectId == 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "系统错误"})
			return
		}

		if userrenew.Pas == "" {
			Db.Table("crm_user").Where("id = ?", c.Query("id")).Updates(map[string]interface{}{"nick": userrenew.Nick, "allow": userrenew.Allow, "group": userrenew.Group})
		} else {
			Db.Table("crm_user").Where("id = ?", c.Query("id")).Updates(map[string]interface{}{"nick": userrenew.Nick, "pas": common.GetMD5Hash(userrenew.Pas), "allow": userrenew.Allow, "group": userrenew.Group})
		}
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "用户修改成功"})
		return
	}
	if c.Query("c") == "list" { //用户列表
		data := xmlist("crm_user", c.Query("page"), c.Query("perPage"), common.Txt(session.Get("UserName")))
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "列表"})
		}
		return
	}
	if c.Query("c") == "del" { //删除用户
		if uall != 1 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		c.AsciiJSON(http.StatusOK, del("crm_user", c.Query("id")))
		return
	}
}

func doctor(c *gin.Context) { //医生管理
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	if uall > 2 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
		return
	}
	if c.Query("c") == "add" { //添加医生
		var doctorNew doctorAdd
		c.ShouldBindJSON(&doctorNew)
		doctorNew.Addtime = time.Now().Unix()

		var projectsId int64
		Db.Table("crm_project").Select("id").Where("id = ?", doctorNew.Group).Where("director = ?", userid(common.Txt(session.Get("UserName")))).Where("state = 1").Find(&projectsId)
		if uall == 2 && projectsId == 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		var doctorId int64
		Db.Table("crm_doctor").Select("id").Where("name = ?", doctorNew.Name).Where("state = 1").Find(&doctorId)
		//fmt.Println(doctorNew)
		if doctorId != 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "医生名重复"})
		} else {
			Db.Table("crm_doctor").Create(&doctorNew)
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功"})
		}
		return

	}
	if c.Query("c") == "renew" { //修改医生
		var doctorNew doctorAdd
		c.ShouldBindJSON(&doctorNew)
		doctorNew.Addtime = time.Now().Unix()

		var projectsId int64
		Db.Table("crm_project").Select("id").Where("id = ?", doctorNew.Group).Where("director = ?", userid(common.Txt(session.Get("UserName")))).Where("state = 1").Find(&projectsId)
		if uall == 2 && projectsId == 0 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		ss := map[string]interface{}{
			"group":     doctorNew.Group,
			"monday":    doctorNew.Monday,
			"tuesday":   doctorNew.Tuesday,
			"wednesday": doctorNew.Wednesday,
			"thursday":  doctorNew.Thursday,
			"friday":    doctorNew.Friday,
			"saturday":  doctorNew.Saturday,
			"sunday":    doctorNew.Sunday,
		}
		Db.Table("crm_doctor").Where("id = ?", c.Query("id")).Updates(ss)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": "", "msg": "修改成功"})
		return
	}
	if c.Query("c") == "list" { //医生列表
		data := xmlist("crm_doctor", c.Query("page"), c.Query("perPage"), common.Txt(session.Get("UserName")))
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "列表"})
		}
		return
	}
	if c.Query("c") == "del" { //删除医生

		if uall != 1 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		c.AsciiJSON(http.StatusOK, del("crm_doctor", c.Query("id")))
		return
	}
}

func projectConfigList(c *gin.Context) { //项目设置列表

	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))

	if uall > 2 { //
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
		return
	}
	if c.Query("c") == "manager" { //项目主管
		var managerlist []manageroptions
		Db.Table("crm_user").Select("id,nick").Order("id desc").Where("allow = 2").Find(&managerlist)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": managerlist, "msg": "项目主管列表"})
		return
	}
	if c.Query("c") == "platformlist" { //平台列表
		var platformoptions []platformoptions
		Db.Table("crm_platform").Select("id,name").Order("id desc").Where("state = 1").Find(&platformoptions)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": platformoptions, "msg": "平台列表"})
		return
	}
	if c.Query("c") == "doctorlist" { //医生列表
		var platformoptions []platformoptions
		Db.Table("crm_doctor").Select("id,name").Order("id desc").Where("`group` = ?", c.Query("id")).Where("state = 1").Find(&platformoptions)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": platformoptions, "msg": "医生列表"})
		return
	}
	if c.Query("c") == "group" { //项目组
		var platformoptions []platformoptions
		xmlist := Db.Table("crm_project").Select("id,name").Order("id desc").Where("state = 1")
		if uall == 2 {
			xmlist.Where("director = ?", userid(common.Txt(session.Get("UserName"))))
		}
		xmlist.Find(&platformoptions)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": platformoptions, "msg": "项目列表"})
		return
	}

	if c.Query("c") == "allow" { //权限组
		var allow []platformoptions

		if uall == 1 {
			allow = append(allow, platformoptions{
				Name: "超级管理员",
				Id:   1,
			}, platformoptions{
				Name: "项目管理员",
				Id:   2,
			})
		}
		allow = append(allow, platformoptions{
			Name: "套电",
			Id:   3,
		}, platformoptions{
			Name: "咨询",
			Id:   4,
		}, platformoptions{
			Name: "医助",
			Id:   5,
		})

		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": allow, "msg": "权限表"})
		return
	}

	//c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": "", "msg": "列表"})
}
func pas(c *gin.Context) { //密码修改
	var pas password
	c.ShouldBindJSON(&pas)
	if pas.Pas2 != pas.Pas3 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "两次密码输入不一致！"})
		return
	}
	session := sessions.Default(c)
	var userid int64
	Db.Table("crm_user").Select("id").Where("name = ?", session.Get("UserName")).Where("pas = ?", common.GetMD5Hash(pas.Pas1)).Take(&userid)
	if userid == 0 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "原始密码错误"})
		return
	}
	Db.Table("crm_user").Where("name = ?", session.Get("UserName")).Update("pas", common.GetMD5Hash(pas.Pas2))
	c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "密码修改成功"})

}

func del(table, id string) delstate { //统一删除

	state := delstate{Status: 0, Msg: "删除成功"}
	Db.Table(table).Where("id = ?", id).Update("state", 0)
	return state

}
func xmlist(table, pages, pageSizes, uname string) list { //统一列表
	var project []map[string]interface{}
	var count int64
	var data list
	page, _ := strconv.Atoi(pages)
	pageSize, _ := strconv.Atoi(pageSizes)
	da := Db.Table(table).Order("id desc").Where("state = 1").Limit(pageSize).Offset((page - 1) * pageSize)
	db := Db.Table(table).Where("state = 1")

	uall := userallow(uname) //用户组
	uid := userid(uname)

	if uall == 2 && (table == "crm_user" || table == "crm_doctor") { //项目管理只显示自己的
		var projectid []string
		Db.Table("crm_project").Select("`id`").Where("director = ?", uid).Where("state = 1").Take(&projectid) //列表
		//fmt.Println(uid)
		//fmt.Println(projectid)

		da.Where("`group` IN ?", projectid)
		db.Where("`group` IN ?", projectid)
	}

	if uall == 2 && table == "crm_project" { //管理项目
		var uid int64
		Db.Table("crm_user").Select("`id`").Where("name = ?", uname).Where("state = 1").Take(&uid)
		da.Where("director = ?", uid)
		db.Where("director = ?", uid)

	}

	da.Find(&project)
	db.Count(&count)
	if table == "crm_user" {
		for _, v := range project {
			v["pas"] = ""
		}
	}
	if table == "crm_doctor" {
		for _, v := range project {
			var name string
			Db.Table("crm_project").Select("`name`").Where("id = ?", v["group"]).Where("state = 1").Take(&name)
			v["groups"] = name
		}

	}

	data.Count = count
	data.Rows = project
	return data
}

func userallow(userName string) int64 { //用户组

	var userState int64
	Db.Table("crm_user").Select("`allow`").Where("name = ?", userName).Where("state = 1").Take(&userState) //列表
	return userState
}
func usergroup(userName string) int64 { //用户组

	var userState int64
	Db.Table("crm_user").Select("`group`").Where("name = ?", userName).Where("state = 1").Take(&userState) //列表
	return userState
}
func userid(userName string) int64 { //用户id
	var userState int64
	Db.Table("crm_user").Select("`id`").Where("name = ?", userName).Where("state = 1").Take(&userState) //列表
	return userState
}

/////////////

func dataEntry(c *gin.Context) { //数据录入
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	uid := userid(common.Txt(session.Get("UserName")))

	if c.Query("c") == "add" { //增加
		if uall != 3 { //
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "没有权限"})
			return
		}
		var entry dataEntryAdd
		c.ShouldBindJSON(&entry)

		var entryPhoneId entryPhone
		Db.Table("crm_customer").Where("phone = ?", entry.Phone).Take(&entryPhoneId)

		if entryPhoneId.Id != 0 {
			tjsj := time.Unix(entryPhoneId.Addtime, 0).In(cstSh).Format("2006-01-02 15:04:05")
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "电话已存在:" + entryPhoneId.Name + "-" + tjsj})
			return
		}
		entry.Addtime = time.Now().Unix()
		entry.Project = usergroup(common.Txt(session.Get("UserName")))
		entry.Phoneuser = uid
		Db.Table("crm_customer").Create(&entry)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "添加成功"})
		return
	}
	if c.Query("c") == "list" { //项目列表

		if uall > 3 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "无权限"})
			return
		}
		var project []map[string]interface{}
		var count int64
		var data list
		ugroup := usergroup(common.Txt(session.Get("UserName")))
		page, _ := strconv.Atoi(c.Query("page"))
		pageSize, _ := strconv.Atoi(c.Query("perPage"))
		da := Db.Table("crm_customer")

		if uall == 2 {
			da.Where("project = ?", ugroup)
		}
		if uall == 3 {
			da.Where("phoneuser = ?", uid).Where("project = ?", ugroup)
		}
		if c.Query("keywords") != "" {
			da.Where("`phone` LIKE ? or `name` LIKE ?", "%"+c.Query("keywords")+"%", "%"+c.Query("keywords")+"%")

		}
		if c.Query("type") == "1" {
			da.Where("consultuser = 0")
		}
		if c.Query("type") == "2" {
			da.Where("returnvisit = 0")
		}
		if c.Query("type") == "3" {
			da.Where("state = 0")
		}
		da.Count(&count)
		da.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&project)

		for _, v := range project {
			var projectName string
			var doctorName string
			var consultuserName string
			Db.Table("crm_platform").Select("`name`").Where("id = ?", v["platform"]).Where("state = 1").Take(&projectName)
			Db.Table("crm_doctor").Select("`name`").Where("id = ?", v["doctor"]).Where("state = 1").Take(&doctorName)
			Db.Table("crm_user").Select("`nick`").Where("id = ?", v["consultuser"]).Where("state = 1").Take(&consultuserName)
			v["platforms"] = projectName
			v["doctors"] = doctorName
			illness := common.Txt(v["illness"])
			v["illnesss"] = string([]rune(illness)[:15])
			v["consultusers"] = consultuserName
		}

		data.Count = count
		data.Rows = project
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "列表"})
		}
		return
	}
	if c.Query("c") == "renew" { //修改

		if uall > 3 || uall == 1 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "无权限"})
			return
		}
		var entry dataEntryRenew
		c.ShouldBindJSON(&entry)
		da := Db.Table("crm_customer").Where("id = ?", c.Query("id"))

		if uall == 2 { //检查是和否该组
			var customerid int64
			Db.Table("crm_customer").Select("id").Where("id = ?", c.Query("id")).Where("project = ?", usergroup(common.Txt(session.Get("UserName")))).Take(&customerid)
			if customerid == 0 {
				c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "无权限"})
				return
			}

		}

		if uall == 3 { //检测是否 本人
			da.Where("phoneuser = ?", uid)
		}
		da.Updates(entry)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "修改成功"})
		return
	}

}
func phonelist(c *gin.Context) { //咨询电话列表
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	uid := userid(common.Txt(session.Get("UserName")))
	if c.Query("c") == "list" { //项目列表
		if uall == 5 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "系统错误"})
			return
		}
		var project []map[string]interface{}
		var count int64
		var data list
		ugroup := usergroup(common.Txt(session.Get("UserName")))
		page, _ := strconv.Atoi(c.Query("page"))
		pageSize, _ := strconv.Atoi(c.Query("perPage"))
		da := Db.Table("crm_customer")

		if uall == 2 { //项目管理
			da.Where("project = ?", ugroup)
		}
		if uall == 3 { //套电
			da.Where("phoneuser = ?", uid).Where("project = ?", ugroup)
		}
		if uall == 4 { //咨询
			da.Where("consultuser = ?", uid).Where("project = ?", ugroup)
		}
		if c.Query("keywords") != "" {
			da.Where("`phone` LIKE ? or `name` LIKE ?", "%"+c.Query("keywords")+"%", "%"+c.Query("keywords")+"%")
		}
		if c.Query("type") == "1" {
			da.Where("returnvisit = 0")
		}
		if c.Query("type") == "2" {
			da.Where("returnvisit > 0")
		}
		if c.Query("type") == "3" {
			da.Where("state = 0")
		}

		if c.Query("select") == "1" {
			da.Where("state = 1")
		}
		if c.Query("select") == "2" {
			da.Where("state = 2")
		}
		if c.Query("select") == "3" {
			da.Where("state = 3")
		}
		if c.Query("select") == "4" {
			da.Where("state = 4")
		}
		da.Count(&count)
		da.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&project)

		for _, v := range project {
			var projectName string
			var doctorName string
			var consultuserName string
			Db.Table("crm_platform").Select("`name`").Where("id = ?", v["platform"]).Where("state = 1").Take(&projectName)
			Db.Table("crm_doctor").Select("`name`").Where("id = ?", v["doctor"]).Where("state = 1").Take(&doctorName)
			Db.Table("crm_user").Select("`nick`").Where("id = ?", v["consultuser"]).Where("state = 1").Take(&consultuserName)
			v["platforms"] = projectName
			v["doctors"] = doctorName
			illness := common.Txt(v["illness"])
			v["illnesss"] = string([]rune(illness)[:15])
			v["consultusers"] = consultuserName
		}

		data.Count = count
		data.Rows = project
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "列表"})
		}
		return
	}
	if c.Query("c") == "renew" { //修改

		if uall == 3 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "无权限"})
			return
		}
		var entry dataEntryRenew
		c.ShouldBindJSON(&entry)
		da := Db.Table("crm_customer").Where("id = ?", c.Query("id"))

		if uall == 2 || uall == 5 { //检查是和否该组
			var customerid int64
			Db.Table("crm_customer").Select("id").Where("id = ?", c.Query("id")).Where("project = ?", usergroup(common.Txt(session.Get("UserName")))).Take(&customerid)
			if customerid == 0 {
				c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "无权限"})
				return
			}
		}

		if uall == 4 { //检测是否 本人
			da.Where("consultuser = ?", uid)
		}
		da.Omit("platform").Updates(entry)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "修改成功"})
		return
	}

}

func reserve(c *gin.Context) { //预约
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))
	uid := userid(common.Txt(session.Get("UserName")))
	usergroups := usergroup(common.Txt(session.Get("UserName")))

	if c.Query("c") == "arrive" {

		if uall != 5 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "只有医助才能操作！"})
			return
		}
		var dzdata dzReserve
		c.ShouldBindJSON(&dzdata)
		Db.Table("crm_customer_reserve").Where("id = ?", c.Query("id")).Where("projectid = ?", usergroups).Updates(dzdata)

		return
	}
	if c.Query("c") == "add" {
		if uall != 4 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "只有咨询可以预约"})
			return
		}
		var reservedata zxReserve
		c.ShouldBindJSON(&reservedata)
		reservedata.Addtime = time.Now().Unix()
		reservedata.User = uid
		reservedata.Projectid = usergroups
		reservedata.Customerid = c.Query("id")

		Db.Table("crm_customer_reserve").Create(&reservedata)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "预约成功"})
		return

	}
	if c.Query("c") == "list" {

		var project []map[string]interface{}
		var count int64
		var data list
		page, _ := strconv.Atoi(c.Query("page"))
		pageSize, _ := strconv.Atoi(c.Query("perPage"))

		da := Db.Table("crm_customer_reserve")
		da.Select("crm_customer.phoneuser as phoneuser,crm_customer_reserve.remarks as remarks,crm_customer_reserve.reagent_type as reagent_type,crm_customer_reserve.process_cost as process_cost,crm_customer_reserve.frequency as frequency,crm_customer_reserve.drug_cost as drug_cost,crm_customer_reserve.patient_rating as patient_rating,crm_customer_reserve.hsid as hsid,crm_customer_reserve.treatment_method as treatment_method,crm_customer.illness as illness,crm_customer.card as card,crm_customer.age as age,crm_customer_reserve.repeats as repeats,crm_customer_reserve.pay_time as pay_time,crm_customer_reserve.customerid as customerid,crm_customer_reserve.projectid as projectid,crm_customer_reserve.id as id,crm_customer.name as name,crm_customer.gender as gender,crm_customer.phone as phone,crm_customer.platform as platform,crm_customer_reserve.visit_time as visit_time,crm_customer_reserve.registration_fee as registration_fee,crm_customer_reserve.state as state,crm_customer_reserve.doctor as doctor,crm_customer_reserve.user as user")
		da.Joins("left join crm_customer on crm_customer_reserve.customerid =crm_customer.id")
		da.Order("crm_customer_reserve.id desc").Limit(pageSize).Offset((page - 1) * pageSize)
		if uall == 3 { //套电
			da.Where("crm_customer.phoneuser = ?", uid).Where("crm_customer_reserve.projectid = ?", usergroups)
		}
		if uall == 4 { //咨询
			da.Where("crm_customer_reserve.user = ?", uid).Where("crm_customer_reserve.projectid = ?", usergroups)
		}
		if uall == 5 || uall == 2 { //项目负责人/医助
			da.Where("crm_customer_reserve.projectid = ?", usergroups)
		}
		if c.Query("id") != "" {
			da.Where("customerid = ?", c.Query("id"))
		}
		if c.Query("keywords") != "" {
			da.Where("`phone` LIKE ? or `name` LIKE ?", "%"+c.Query("keywords")+"%", "%"+c.Query("keywords")+"%")
		}
		if c.Query("select") != "" {
			//医生搜索
			da.Where("crm_customer_reserve.doctor = ?", c.Query("select"))
		}
		if c.Query("date") != "" {
			date, _ := strconv.Atoi(c.Query("date"))

			da.Where("crm_customer_reserve.visit_time > ?", date).Where("crm_customer_reserve.visit_time < ?", date+60*60*24)
		}
		da.Count(&count)
		da.Limit(pageSize).Offset((page - 1) * pageSize).Find(&project)

		for _, v := range project {
			var doctors string
			Db.Table("crm_doctor").Select("name").Where("id = ?", v["doctor"]).Take(&doctors)
			v["doctors"] = doctors
			var users string
			Db.Table("crm_user").Select("nick").Where("id = ?", v["user"]).Take(&users)
			v["users"] = users
			var platforms string
			Db.Table("crm_platform").Select("name").Where("id = ?", v["platform"]).Take(&platforms)
			v["platforms"] = platforms

		}

		data.Count = count
		data.Rows = project
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "预约记录"})
		}
		return
	}
}
func returnVisit(c *gin.Context) { //回访
	session := sessions.Default(c)
	uall := userallow(common.Txt(session.Get("UserName")))

	uid := userid(common.Txt(session.Get("UserName")))
	if c.Query("c") == "add" { //项目列表
		if uall != 4 {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "只有咨询可以回访"})
			return
		}
		var visitdata returnVisitadd
		c.ShouldBindJSON(&visitdata)
		visitdata.Revisitdays = time.Now().Unix()
		visitdata.User = uid
		visitdata.Customerid = c.Query("id")
		Db.Table("crm_customer_return_visit").Create(&visitdata)
		Db.Table("crm_customer").Where("id = ?", c.Query("id")).Updates(map[string]interface{}{"state": visitdata.Result, "returnvisit": visitdata.Revisitdays})
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "回访成功"})
		return

	}

	if c.Query("c") == "list" { //回访记录
		var project []map[string]interface{}
		var count int64
		var data list
		page, _ := strconv.Atoi(c.Query("page"))
		pageSize, _ := strconv.Atoi(c.Query("perPage"))
		Db.Table("crm_customer_return_visit").Where("customerid = ?", c.Query("id")).Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&project)
		Db.Table("crm_customer_return_visit").Where("customerid = ?", c.Query("id")).Count(&count)

		data.Count = count
		data.Rows = project
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "回访记录"})
		}
		return

	}
}
func distribution(c *gin.Context) { //数据分配

	var data distributionup
	c.ShouldBindJSON(&data)
	session := sessions.Default(c)
	userallow := userallow(common.Txt(session.Get("UserName")))
	if userallow != 2 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "只有项目负责人才可以分配"})
		return
	}
	ugroup := usergroup(common.Txt(session.Get("UserName")))
	Db.Table("crm_customer").Where("consultuser = 0").Where("project = ?", ugroup).Where("id IN ?", strings.Split(data.Ids, ",")).Update("consultuser", data.Consultfp)
	c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "分配成功"})

}
func dictionary(c *gin.Context) { //客户字典

	session := sessions.Default(c)
	//userallow := userallow(common.Txt(session.Get("UserName")))
	ugroup := usergroup(common.Txt(session.Get("UserName")))
	//uid := userid(common.Txt(session.Get("UserName")))

	if c.Query("c") == "platformlist" { //平台列表
		var userPlatformlist string
		//var userState int64
		Db.Table("crm_project").Select("`platformlist`").Where("id = ?", ugroup).Where("state = 1").Take(&userPlatformlist)

		var platformoptions []platformoptions
		Db.Table("crm_platform").Select("id,name").Order("id desc").Where("`id` IN ?", strings.Split(userPlatformlist, ",")).Where("state = 1").Find(&platformoptions)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": platformoptions, "msg": "平台列表"})
		return
	}
	if c.Query("c") == "doctorlist" { //医生列表
		//var userdoctorlist string
		//var userState int64
		//Db.Table("crm_project").Select("`doctorlist`").Where("id = ?", ugroup).Where("state = 1").Take(&userdoctorlist)
		//fmt.Println(userdoctorlist)
		var platformoptions []platformoptions
		Db.Table("crm_doctor").Select("id,name").Order("id desc").Where("`group` = ?", ugroup).Where("state = 1").Find(&platformoptions)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": platformoptions, "msg": "医生列表"})
		return
	}
	if c.Query("c") == "userlist" { //当前组员列表
		//fmt.Println(ugroup)
		var uList []dictionaryUserList
		Db.Table("crm_user").Select("id,nick").Order("id desc").Where("`group` = ?", ugroup).Where("state = 1").Where("allow = ?", c.Query("allow")).Find(&uList)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": uList, "msg": "用户列表"})
		return
	}

}

///////////

func newbaobiao(s1, s2, startTime, endTime int64) newyuedatas { //新报表数据
	var newyuedatas1 newyuedatas
	//fmt.Println("开始时间", startTime, "结束时间", endTime)
	//fmt.Println("医生ID", s1, "平台id", s2)
	//newyuedatas1.Phone = 20   //套电数量
	Db.Table("crm_customer").Select("COUNT(`id`)").Where("`doctor` = ?", s1).Where("platform = ?", s2).Where("addtime >= ?", startTime).Where("addtime <= ?", endTime).Find(&newyuedatas1.Phone)
	//newyuedatas1.Reserve = 30 //预约数量
	/*
		var list []map[string]interface{}
		ss := Db.Table("crm_customer_reserve")

		ss = ss.Joins("left join crm_customer on crm_customer_reserve.customerid =crm_customer.id")
		ss = ss.Select("crm_customer_reserve.id as id,crm_customer_reserve.doctor as doctor ,crm_customer_reserve.addtime as addtime")
		ss = ss.Where("`doctor` = ?", s1).Where("addtime >= ?", startTime).Where("crm_customer_reserve.addtime <= ?", endTime)
		ss = ss.Find(&list)
	*/
	var project []map[string]interface{}
	da := Db.Table("crm_customer_reserve")
	da.Joins("left join crm_customer on crm_customer_reserve.customerid =crm_customer.id")
	da.Select("crm_customer_reserve.id as id,crm_customer_reserve.doctor as doctor,crm_customer_reserve.addtime as addtime,crm_customer.platform as platform")
	da.Where("crm_customer_reserve.doctor = ?", s1).Where("crm_customer.platform = ?", s2).Where("crm_customer_reserve.addtime >= ?", startTime).Where("crm_customer_reserve.addtime <= ?", endTime)
	da.Find(&project)
	newyuedatas1.Reserve = common.ToInt(len(project))
	//fmt.Println(len(project))
	//Db.Table("crm_customer_reserve").Select("COUNT(`id`)").Joins("left join crm_customer on crm_customer_reserve.customerid =crm_customer.id").Where("`doctor` = ?", s1).Where("addtime >= ?", startTime).Where("addtime <= ?", endTime).Find(&newyuedatas1.Reserve)
	//newyuedatas1.Arrive = 40 //到诊数量
	var project2 []map[string]interface{}
	da2 := Db.Table("crm_customer_reserve")
	da2.Joins("left join crm_customer on crm_customer_reserve.customerid =crm_customer.id")
	da2.Select("crm_customer_reserve.id as id,crm_customer_reserve.state as state,crm_customer_reserve.doctor as doctor,crm_customer_reserve.addtime as addtime,crm_customer.platform as platform")
	da2.Where("crm_customer_reserve.doctor = ?", s1).Where("crm_customer.platform = ?", s2).Where("crm_customer_reserve.addtime >= ?", startTime).Where("crm_customer_reserve.addtime <= ?", endTime).Where("crm_customer_reserve.state = 2")
	da2.Find(&project2)
	newyuedatas1.Arrive = common.ToInt(len(project2))
	//Db.Table("crm_customer_reserve").Select("COUNT(`id`)").Where("`state` = 2").Where("`doctor` = ?", s1).Where("addtime >= ?", startTime).Where("addtime <= ?", endTime).Find(&newyuedatas1.Arrive)

	return newyuedatas1
}

//////////
func statistics(c *gin.Context) { //统计
	session := sessions.Default(c)
	userallow := userallow(common.Txt(session.Get("UserName")))
	ugroup := usergroup(common.Txt(session.Get("UserName")))

	if userallow > 3 {
		c.AsciiJSON(http.StatusOK, gin.H{"status": 1, "msg": "系统错误"})
		return
	}
	if c.Query("c") == "newlist" { //新统计
		var zu string
		if userallow == 1 && c.Query("selectgroup") != "" { //判定小组
			zu = c.Query("selectgroup")
		} else {
			zu = common.Txt(ugroup)
		}
		var startTime int64
		var endTime int64
		if c.Query("startTime") != "" && c.Query("endTime") != "" {
			startTime = common.ToInt(c.Query("startTime"))
			endTime = common.ToInt(c.Query("endTime"))
		} else {
			startTime = time.Now().Unix() - 60*60*24
			endTime = time.Now().Unix()
		}

		var zuysid string
		var zuptid string
		//var yslb []map[string]interface{}

		Db.Table("crm_project").Select("doctorlist").Where("id = ?", zu).Take(&zuysid)
		Db.Table("crm_project").Select("platformlist").Where("id = ?", zu).Take(&zuptid)
		var yuedatalist []newyuedata
		for _, v := range strings.Split(zuysid, ",") {
			var doctorname string
			Db.Table("crm_doctor").Select("name").Where("id = ?", v).Take(&doctorname)
			if c.Query("selectplatform") == "" {
				for _, vzu := range strings.Split(zuptid, ",") {
					var zuname string
					Db.Table("crm_platform").Select("name").Where("id = ?", vzu).Take(&zuname)

					bb1 := newbaobiao(common.ToInt(v), common.ToInt(vzu), startTime, endTime)
					yuedatalist = append(yuedatalist, newyuedata{
						Name:     doctorname,
						Platform: zuname,
						Phone:    bb1.Phone,
						Reserve:  bb1.Reserve,
						Arrive:   bb1.Arrive,
					})
				}
			} else {
				var zuname string
				Db.Table("crm_platform").Select("name").Where("id = ?", c.Query("selectplatform")).Take(&zuname)

				//fmt.Println(v, doctorname, common.ToInt(c.Query("selectplatform")), zuname)
				bb1 := newbaobiao(common.ToInt(v), common.ToInt(c.Query("selectplatform")), startTime, endTime)
				yuedatalist = append(yuedatalist, newyuedata{
					Name:     doctorname,
					Platform: zuname,
					Phone:    bb1.Phone,
					Reserve:  bb1.Reserve,
					Arrive:   bb1.Arrive,
				})

			}

		}

		//var startTime int64
		//var endTime int64
		var data list
		data.Count = common.ToInt(len(yuedatalist))
		data.Rows = yuedatalist
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "新统计"})
		}
		return

		//c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "msg": "新统计"})
		//return
	}
	//uid := userid(common.Txt(session.Get("UserName")))
	if c.Query("c") == "moonlist" {
		var zu string
		var startTime int64
		var endTime int64
		if c.Query("date") == "" {
			startTime = time.Date(time.Now().In(cstSh).Year(), time.Now().In(cstSh).Month(), 01, 0, 0, 0, 0, time.Now().In(cstSh).Location()).Unix()
			endTime = time.Unix(startTime, 0).In(cstSh).AddDate(0, 1, 0).Unix()
		} else {
			startTime = time.Unix(common.ToInt(c.Query("date")), 0).In(cstSh).Unix()
			endTime = time.Unix(common.ToInt(c.Query("date")), 0).In(cstSh).AddDate(0, 1, 0).Unix()
		}

		if userallow == 1 && c.Query("selectgroup") != "" {
			zu = c.Query("selectgroup")
		} else {
			zu = common.Txt(ugroup)
		}
		var yuedatalist []yuedata
		if c.Query("selectplatform") == "" {
			///////
			var ptlbList string
			Db.Table("crm_project").Select("platformlist").Where("id = ?", zu).Where("state = 1").Find(&ptlbList) //列表
			for _, v := range strings.Split(ptlbList, ",") {
				var zxcustomerlist []int64
				var ptmcName string
				Db.Table("crm_platform").Select("name").Where("id = ?", v).Take(&ptmcName)
				Db.Table("crm_customer").Select("id").Where("project = ?", zu).Where("platform = ?", v).Where("addtime > ?", startTime).Where("addtime < ?", endTime).Find(&zxcustomerlist)
				yysl, dzsl, _ := khlist2(zxcustomerlist)
				yuedatalist = append(yuedatalist, yuedata{
					Name:    ptmcName,
					Phone:   len(zxcustomerlist),
					Reserve: yysl,
					Arrive:  dzsl,
					Amount:  0,
				})
				//fmt.Println("平台名称：", ptmcName, "电话数量:", len(zxcustomerlist), "预约数量：", yysl, "到诊数量", dzsl)
			}
			/////
		} else {

			var yszulist []map[string]interface{}
			var ptmcName string
			var zxcustomerlist []int64
			Db.Table("crm_platform").Select("name").Where("id = ?", c.Query("selectplatform")).Take(&ptmcName)

			Db.Table("crm_doctor").Select("name,id").Where("`group` = ?", zu).Find(&yszulist)
			for _, v := range yszulist {

				Db.Table("crm_customer").Select("id").Where("project = ?", zu).Where("doctor = ?", v["id"]).Where("platform = ?", c.Query("selectplatform")).Where("addtime > ?", startTime).Where("addtime < ?", endTime).Find(&zxcustomerlist)
				yysl, dzsl, _ := khlist2(zxcustomerlist)
				yuedatalist = append(yuedatalist, yuedata{
					Name:    ptmcName + "【" + common.Txt(v["name"]) + "】",
					Phone:   len(zxcustomerlist),
					Reserve: yysl,
					Arrive:  dzsl,
					Amount:  0,
				})
				//fmt.Println("平台名称：", ptmcName+"【"+common.Txt(v["name"])+"】", "电话数量:", len(zxcustomerlist), "预约数量：", yysl, "到诊数量", dzsl)
			}

		}

		//fmt.Println(endTime)

		var data list
		data.Count = common.ToInt(len(yuedatalist))
		data.Rows = yuedatalist
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "回访记录"})
		}
		return

	}
	if c.Query("c") == "consultlist" { //咨询统计

		var zu string
		var startTime int64
		var endTime int64

		if c.Query("date") == "" {
			startTime = time.Date(time.Now().In(cstSh).Year(), time.Now().In(cstSh).Month(), 01, 0, 0, 0, 0, time.Now().In(cstSh).Location()).Unix()
			endTime = time.Unix(startTime, 0).In(cstSh).AddDate(0, 1, 0).Unix()
		} else {
			startTime = time.Unix(common.ToInt(c.Query("date")), 0).In(cstSh).Unix()
			endTime = time.Unix(common.ToInt(c.Query("date")), 0).In(cstSh).AddDate(0, 1, 0).Unix()
		}

		if userallow == 1 && c.Query("selectgroup") != "" {
			zu = c.Query("selectgroup")
		} else {
			zu = common.Txt(ugroup)
		}
		var userlist []dictionaryUserList
		Db.Table("crm_user").Select("id,nick").Where("`group` = ?", zu).Where("`allow` = 4").Where("`state` = 1").Find(&userlist)
		var yuedatalist []yuedata
		for _, v := range userlist {
			var zxcustomerlist []int64
			Db.Table("crm_customer").Select("id").Where("consultuser = ?", v.Id).Where("addtime > ?", startTime).Where("addtime < ?", endTime).Find(&zxcustomerlist)
			yysl, dzsl, xfje := khlist2(zxcustomerlist)
			yuedatalist = append(yuedatalist, yuedata{
				Name:    v.Nick,
				Phone:   len(zxcustomerlist),
				Reserve: yysl,
				Arrive:  dzsl,
				Amount:  xfje,
			})

		}
		var data list
		data.Count = common.ToInt(len(yuedatalist))
		data.Rows = yuedatalist
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "回访记录"})
		}
		return

	}
	if c.Query("c") == "daylist" {
		var project []statisticsdata
		//var count int64
		var projects []int64
		var name string
		var data list

		var zu string

		da := Db.Table("crm_customer").Select("id")
		db := Db.Table("crm_project").Select("name")
		if userallow == 1 && c.Query("selectgroup") != "" {
			zu = c.Query("selectgroup")
		} else {
			zu = common.Txt(ugroup)
		}
		db.Where("id = ?", zu).Take(&name)
		da.Where("project = ?", zu)
		var startTime string
		var endTime string
		if c.Query("date") != "" {
			datess, _ := strconv.Atoi(c.Query("date"))
			startTime = common.Txt(datess)
			endTime = common.Txt(datess + 60*60*24)
		} else {
			datess := time.Date(time.Now().In(cstSh).Year(), time.Now().In(cstSh).Month(), time.Now().In(cstSh).Day(), 0, 0, 0, 0, time.Now().In(cstSh).Location()).Unix()
			startTime = common.Txt(datess)
			endTime = common.Txt(datess + 60*60*24)
		}
		da.Where("addtime > ?", startTime).Where("addtime < ?", endTime)
		da.Find(&projects)
		var yylst []int64
		for _, v := range projects {

			var yy int64
			Db.Table("crm_customer_reserve").Select("id").Where("customerid = ?", v).Take(&yy)
			if yy != 0 {
				yylst = append(yylst, yy)
			}
		}

		project = append(project, statisticsdata{
			Name:    name,
			Phone:   len(projects),
			Reserve: len(yylst),
			Rate:    fmt.Sprintf("%.2f", float64(len(yylst))/float64(len(projects))),
		})

		if c.Query("type") == "1" { //平台

			var ptzulist string
			Db.Table("crm_project").Select("platformlist").Where("id = ?", zu).Take(&ptzulist)
			for _, v := range strings.Split(ptzulist, ",") {
				/*
					for _, vv := range khlist(v, startTime, endTime) {
						project = append(project, vv)
					}
				*/
				project = append(project, khlist(v, startTime, endTime)...)
			}

		}
		if c.Query("type") == "2" { //医生
			var ptyslist string
			Db.Table("crm_project").Select("doctorlist").Where("id = ?", zu).Take(&ptyslist)
			for _, v := range strings.Split(ptyslist, ",") {
				//fmt.Println(v)

				/*
					for _, vv := range khlistdoctor(v, startTime, endTime) {
						project = append(project, vv)
					}
				*/
				project = append(project, khlistdoctor(v, startTime, endTime)...)

				//khlistdoctor
			}
		}

		data.Count = 1
		data.Rows = project
		if data.Count == 0 {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.String(http.StatusOK, `{"data":{"count":0,"rows":[]},"msg":"ok","status":0}`)
		} else {
			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": data, "msg": "回访记录"})
		}
		return
	}

	if c.Query("c") == "selectgroup" { //可以查看的小组
		var projectlist []platformoptions
		da := Db.Table("crm_project").Where("state = 1")
		if userallow != 1 {
			da.Where("id = ?", ugroup) //列表
		}
		da.Find(&projectlist)
		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": projectlist, "msg": "平台列表"})
	}

	if c.Query("c") == "selectplatform" { //可以查看的平台
		var projectlist []platformoptions
		var pt string
		Db.Table("crm_project").Select("platformlist").Where("id = ?", c.Query("selectgroup")).Take(&pt)
		Db.Table("crm_platform").Where("state = 1").Where("`id` IN ?", strings.Split(pt, ",")).Find(&projectlist)

		c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": projectlist, "msg": "平台列表"})
	}
	/*
		if c.Query("c") == "selectuser" { //可以查看的咨询员
			var projectlist []manageroptions

			Db.Table("crm_user").Where("`state` = 1").Where("`allow` = 4").Where("`group` = ?", c.Query("selectgroup")).Find(&projectlist)

			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": projectlist, "msg": "平台列表"})
		}
	*/
	/*
		if c.Query("c") == "selectdoctor" { //可以查看的平台
			var projectlist []platformoptions
			var ys string
			Db.Table("crm_project").Select("doctorlist").Where("id = ?", c.Query("selectgroup")).Take(&ys)
			Db.Table("crm_doctor").Where("state = 1").Where("`id` IN ?", strings.Split(ys, ",")).Find(&projectlist)

			c.AsciiJSON(http.StatusOK, gin.H{"status": 0, "data": projectlist, "msg": "平台列表"})
		}
	*/
}
func khlist2(s1 []int64) (s2, s3, s4 int64) { //日报细查

	//for _, v := range s1 {
	var yysl int64
	var count int64
	var sum int64
	Db.Table("crm_customer_reserve").Select("id").Where("customerid IN ?", s1).Count(&yysl)
	Db.Table("crm_customer_reserve").Select("id").Where("customerid IN ?", s1).Where("state = 2").Count(&count)
	Db.Table("crm_customer_reserve").Select("SUM(`drug_cost`)").Where("customerid IN ?", s1).Where("state = 2").Find(&sum)
	return yysl, count, sum

}
func khlist(s1, s2, s3 string) []statisticsdata { //日报细查

	var projects []int64
	var ptname string
	Db.Table("crm_platform").Select("name").Where("id = ?", s1).Find(&ptname)
	var project []statisticsdata
	var yylst []int64
	Db.Table("crm_customer").Select("id").Where("platform = ?", s1).Where("addtime > ?", s2).Where("addtime < ?", s3).Find(&projects)
	for _, v := range projects {
		var yy int64
		Db.Table("crm_customer_reserve").Select("id").Where("customerid = ?", v).Take(&yy)
		if yy != 0 {
			yylst = append(yylst, yy)
		}

	}
	project = append(project, statisticsdata{
		Name:    ptname,
		Phone:   len(projects),
		Reserve: len(yylst),
		Rate:    fmt.Sprintf("%.2f", float64(len(yylst))/float64(len(projects))),
	})
	return project

}
func khlistdoctor(s1, s2, s3 string) []statisticsdata { //日报医生细查

	var projects []int64
	var ptname string
	Db.Table("crm_doctor").Select("name").Where("id = ?", s1).Find(&ptname)
	var project []statisticsdata
	var yylst []int64
	Db.Table("crm_customer").Select("id").Where("doctor = ?", s1).Where("addtime > ?", s2).Where("addtime < ?", s3).Find(&projects)
	for _, v := range projects {
		var yy int64
		Db.Table("crm_customer_reserve").Select("id").Where("customerid = ?", v).Take(&yy)
		if yy != 0 {
			yylst = append(yylst, yy)
		}
	}
	project = append(project, statisticsdata{
		Name:    ptname,
		Phone:   len(projects),
		Reserve: len(yylst),
		Rate:    fmt.Sprintf("%.2f", float64(len(yylst))/float64(len(projects))),
	})
	return project

}

/*
func khlist(table, pages, pageSizes, uname string) list { //客户列表
	var project []map[string]interface{}
	var count int64
	var data list
	page, _ := strconv.Atoi(pages)
	pageSize, _ := strconv.Atoi(pageSizes)
	da := Db.Table(table).Order("id desc").Where("state = 1").Limit(pageSize).Offset((page - 1) * pageSize)
	db := Db.Table(table).Where("state = 1")

	//uall := userallow(uname) //用户组
	//uid := userid(uname)

	da.Find(&project)
	db.Count(&count)
	if table == "crm_customer" {
		for _, v := range project {
			var projectName string
			var doctorName string
			var consultuserName string
			Db.Table("crm_platform").Select("`name`").Where("id = ?", v["platform"]).Where("state = 1").Take(&projectName)
			Db.Table("crm_doctor").Select("`name`").Where("id = ?", v["doctor"]).Where("state = 1").Take(&doctorName)
			Db.Table("crm_user").Select("`name`").Where("id = ?", v["consultuser"]).Where("state = 1").Take(&consultuserName)
			v["platforms"] = projectName
			v["doctors"] = doctorName
			illness := common.Txt(v["illness"])
			v["illness"] = string([]rune(illness)[:15])
			v["consultusers"] = consultuserName
		}

	}

	data.Count = count
	data.Rows = project
	return data
}
*/
