package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func init() {

	var err error

	dsn := "root:7758521@tcp(127.0.0.1:3306)/xmtcrm?charset=utf8mb4&parseTime=True&loc=Local"
	//dsn := "xmtcrm:jxFoygPv3WGSPvrycFAOIE7ZtBdUN93s@tcp(rm-2ze12988qb9ehpcf490130.mysql.rds.aliyuncs.com:3306)/xmtcrm?charset=utf8mb4&parseTime=True&loc=Local"

	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //日志级别
		//Logger: logger.Default.LogMode(logger.Info), //日志级别
	})

	//_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
	sqlDB, _ := _db.DB()
	//设置数据库连接池参数
	sqlDB.SetMaxOpenConns(4) //设置数据库连接池最大连接数
	sqlDB.SetMaxIdleConns(2) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
}

func GetDB() *gorm.DB {
	return _db
}
