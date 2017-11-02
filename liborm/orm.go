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

// 返回id
func (orm *Orm) Insert(md interface{}) (int64, error) {
	mi, ind := orm.getModelInfoAndIndtype(md)
	val := insertKeyValues(mi, ind)
	orm.db.InsertValue(mi.Table, val)
	return 0, nil
}

// 更新表
// vals[0]表示修改的字段
// vals[1]表示修改的条件
// 如果不传的话，默认用model的主键进行更新，如果vals没有传递的话，默认全部更新
func (orm *Orm) Update(md interface{}, vals ...[]*ModelTableFieldConditionInfo) error {
	mi, ind := orm.getModelInfoAndIndtype(md)

	val := updateKeyValues(mi, ind, vals...)

	return orm.db.UpdateValue(mi.Table, val)
}

// 删除表字段,如果没有传入val 则默认删除主键
func (orm *Orm) Delete(md interface{}, val ...[]*ModelTableFieldConditionInfo) (int64, error) {
	mi, ind := orm.getModelInfoAndIndtype(md)
	v := deleteKeyValues(mi, ind, val...)
	return orm.db.DeleteValue(mi.Table, v)
}
