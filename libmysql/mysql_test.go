package libmysql

import (
	"fmt"
	"testing"
)

var (
	createSQL = "CREATE TABLE `userinfo` ( `id` INT(10) NOT NULL AUTO_INCREMENT,  `name` VARCHAR(64) NULL DEFAULT NULL, `created` DATE NULL DEFAULT NULL, PRIMARY KEY (`id`));"
	option    = &Options{
		User:         "root",
		Pwd:          "mysql",
		Host:         "127.0.0.1:3306",
		DB:           "datest",
		MaxOpenConns: 16,
		MaxIdleConns: 4,
	}
)

// func TestCreate(t *testing.T) {

// 	if err := Init(option); err != nil {
// 		t.Fatal(err)
// 	}
// 	ret, err := execute(createSQL)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("ret:", ret)
// }

func TestTrans(t *testing.T) {
	if err := Init(option); err != nil {
		t.Fatal(err)
	}

	// id, err := Insert("insert userinfo set name = 'xx'")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log("id:", id)

	// Insert("insert userinfo set name = 'xxx'")
	// Insert("insert userinfo set name = 'xxxx'")
	// Insert("insert userinfo set name = 'xxxxx'")
	// Insert("insert userinfo set name = 'xxxxxx'")

	trans, e := Begin()
	if e != nil {
		t.Fatal(e)
	}
	ret, e := trans.Query("select * from userinfo where name = 'xx'")
	if e != nil {
		t.Fatal(e)
	}
	for k := range ret {
		fmt.Println("第", k, "行")
		for v := range ret[k] {
			fmt.Println(v, ret[k][v])
		}
	}

	trans.Insert("insert userinfo set name = 'xxxxxxxxxxxx'")
	trans.Commit()
}
