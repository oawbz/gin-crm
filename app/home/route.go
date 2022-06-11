package home

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApiPost(c *gin.Context) {

	switch c.Param("name") {
	case "project": //项目管理
		project(c)
	case "automaticallocation": //自动分配
		automaticallocation(c)
	case "platform": //平台管理
		platform(c)
	case "user": //用户管理
		user(c)
	case "doctor": //医生管理
		doctor(c)
	case "pas": //密码修改
		pas(c)
	case "DataEntry": //数据录入
		dataEntry(c)
	case "distribution": //分配
		distribution(c)
	case "returnVisit": //回访
		returnVisit(c)
	case "phonelist": //电话列表
		phonelist(c)
	case "reserve": //预约
		reserve(c)
	default:
		c.JSON(http.StatusOK, gin.H{"status": 1, "msg": "404"})
	}

}

func ApiGet(c *gin.Context) {
	switch c.Param("name") {
	case "menu": //菜单
		menu(c)
	case "navigation": //导航
		navigation(c)
	case "automaticallocation": //自动分配
		automaticallocation(c)
	case "project": //项目管理
		project(c)
	case "platform": //平台管理
		platform(c)
	case "user": //用户管理
		user(c)
	case "doctor": //用户管理
		doctor(c)
	case "projectConfigList": //字典
		projectConfigList(c)
	case "DataEntry": //数据录入
		dataEntry(c)
	case "dictionary": //客户字典
		dictionary(c)
	case "phonelist": //电话列表
		phonelist(c)
	case "returnVisit": //回访
		returnVisit(c)
	case "reserve": //预约
		reserve(c)
	case "statistics": //统计
		statistics(c)

	default:
		c.JSON(http.StatusOK, gin.H{"status": 1, "msg": "404"})
	}

}
