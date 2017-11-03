package main

import (
	"encoding/json"

	"github.com/wuqifei/server_lib/libmysql"
	"github.com/wuqifei/server_lib/liborm"
	"github.com/wuqifei/server_lib/logs"
)

type User struct {
	UUIB uint32 `orm:"-"` //忽略
	// 类型|size|名称|是否为空,剩下的就是mysql原生的字段，
	// 如果所有字段为空，则会自动指定一个相应的类型，字符串等等
	Tibick uint32 `orm:"size:10|name:bbq|null:true|UNIQUE"`
}

// 继承于FF118
type USER8999 struct {
	IBD     uint64 `orm:"name:id|null:false|AUTO_INCREMENT|UNIQUE|PRIMARY KEY"`
	Pads    string `orm:"size:150|name:txt_b"`
	KLaos   string `orm:"null:true"`
	AKNIO   string
	Bkaskl  string
	Bkaskl2 string
	Bkaskl3 string

	User
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
	orm.RegisterModelWithTableName("", &USER8999{}, []string{"ENGINE=InnoDB", "CHARSET=utf8"})
	mysql := libmysql.NewMysql(option)
	orm.RegisterDB(mysql)
	orm.BootInDB()

	user := &USER8999{}
	user.Pads = "wwwww"
	user.IBD = 123414123
	user.KLaos = "wwwww"
	user.AKNIO = "swww"
	user.UUIB = 123411
	user.Tibick = 889
	// id, _ := orm.Insert(user)
	// logs.Info("id:[%d]", id)
	// orm.Update(user)
	// orm.Delete(user)

	// user2 := &UserInfoBBQ8{}

	// vals, e := orm.Select(user2)
	// if e != nil {
	// 	logs.Debug("err------[%v]", e)
	// }
	// b, _ := json.Marshal(vals)
	// logs.Debug("------%s", string(b))

	t := orm.NewTransaction()
	t.Begin()
	id, _ := t.Insert(user)
	user.IBD = uint64(id)
	user.KLaos = "qqqqqq"
	user.Tibick = 575757
	t.Update(user)
	vals, _ := t.Select(user)

	b, _ := json.Marshal(vals)
	logs.Debug("---1---%s", string(b))
	t.Commit()

	vals, _ = orm.Select(user)

	b, _ = json.Marshal(vals)
	logs.Debug("---2---%s", string(b))
}
