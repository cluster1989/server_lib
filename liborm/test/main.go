package main

import (
	"github.com/wuqifei/server_lib/libmysql"
	"github.com/wuqifei/server_lib/liborm"
)

type FF118 struct {
	BBID uint32 `orm:"-"` //忽略
	// 类型|size|名称|是否为空,剩下的就是mysql原生的字段，
	// 如果所有字段为空，则会自动指定一个相应的类型，字符串等等
	BCID uint32 `orm:"size:10|name:bbq|null:true|UNIQUE"`
}

// 继承于FF118
type FF117 struct {
	FF118
	C       string `orm:"size:150|name:txt_b"`
	IBD     uint64 `orm:"name:id|null:false|AUTO_INCREMENT|UNIQUE|PRIMARY KEY"`
	B       string `orm:"null:true"`
	Tatable string
}

var (
	option = &libmysql.Options{
		User:         "root",
		Pwd:          "12345678",
		Host:         "127.0.0.1:3306",
		DB:           "datest",
		MaxOpenConns: 16,
		MaxIdleConns: 4,
	}
)

func main() {
	orm := liborm.NewOrm()
	orm.RegisterModelWithTableName("bkb", &FF117{}, []string{"ENGINE=InnoDB", "CHARSET=utf8"})
	mysql := libmysql.NewMysql(option)
	orm.RegisterDB(mysql)
	orm.BootInDB()
}