package libmysql

import (
	"fmt"
	"testing"
)

var (
	createSQL = "CREATE TABLE `user` ( `id` INT(10) NOT NULL AUTO_INCREMENT,  `name` VARCHAR(64) NULL DEFAULT NULL, `created` DATE NULL DEFAULT NULL, PRIMARY KEY (`id`));"
	option    = &Options{
		User:         "root",
		Pwd:          "12345678",
		Host:         "127.0.0.1:3306",
		DB:           "datest",
		MaxOpenConns: 16,
		MaxIdleConns: 4,
	}
)

func TestTrans(t *testing.T) {

	msql := NewMysql(option)

	// id, err := msql.Insert("insert user set tel = 'xx',pwd = '123456'")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log("id:", id)

	// msql.Insert("insert user set tel = 'xxx'，pwd = '123456'")
	// msql.Insert("insert user set tel = 'xxxx',pwd = '123456'")
	// msql.Insert("insert user set tel = 'xxxxx',pwd = '123456'")
	// msql.Insert("insert user set tel = 'xxxxxx',pwd = '123456'")

	trans, e := msql.NewLocalransaction()
	if e != nil {
		t.Fatal(e)
	}
	ret, e := trans.Query("select * from user where tel = 'xx'")
	if e != nil {
		t.Fatal(e)
	}
	for k := range ret {
		fmt.Println("第", k, "行")
		for v := range ret[k] {
			fmt.Println(v, ret[k][v])
		}
	}

	trans.Insert("insert user set tel = 'xxxxxxxxxxxx'")
	trans.Commit()
}
