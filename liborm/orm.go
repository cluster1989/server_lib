package liborm

import (
	"github.com/wuqifei/server_lib/concurrent"
)

type Orm struct {
	// 对象模型的储存
	modelCache *concurrent.ConcurrentMap

	// 注册的orm储存
	db SQL
}

func NewOrm() *Orm {
	orm := &Orm{}
	orm.modelCache = concurrent.NewCocurrentMap()
	return orm
}

// 根据对象名创建表
func (orm *Orm) RegisterModel(models ...interface{}) {
	for _, m := range models {
		orm.registerModel("", m, nil)
	}
}

// 根据表名创建表
func (orm *Orm) RegisterModelWithTableName(tableName string, model interface{}, tags []string) {
	orm.registerModel(tableName, model, tags)
}

func (orm *Orm) RegisterDB(sql SQL) {
	orm.db = sql
}

func (orm *Orm) BootInDB() error {
	if orm.db == nil {
		panic("orm: should register db first")
	}
	tables := make([]*ModelTableInfo, 0)
	for _, v := range orm.modelCache.Items {
		tables = append(tables, v.(*ModelTableInfo))
	}
	orm.db.RegistNewTable(tables)
	return nil
}
