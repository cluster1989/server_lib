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
	return orm.db.InsertValue(mi.Table, val)
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

// 读取表数据
// 如果传入数据，第一个val是代表想要查询的数据
// 第二个val代表查询的where的条件
// 第三个val中的数据，key可以是（ORDER BY, LIMIT , ） value(a desc,...)
// 不会默认主键查询，默认全表查询
func (orm *Orm) Select(md interface{}, vals ...[]*ModelTableFieldConditionInfo) ([]interface{}, error) {

	mi, ind := orm.getModelInfoAndIndtype(md)
	var (
		searchCondition []*ModelTableFieldConditionInfo
		whereCondition  []*ModelTableFieldConditionInfo
		sqlCondition    []*ModelTableFieldConditionInfo
	)
	if len(vals) == 3 {
		searchCondition = selectKeyValues(mi, ind, vals[0])
		whereCondition = selectKeyValues(mi, ind, vals[1])
		sqlCondition = vals[2]
	} else if len(vals) == 2 {
		searchCondition = selectKeyValues(mi, ind, vals[0])
		whereCondition = selectKeyValues(mi, ind, vals[1])
	} else if len(vals) == 1 {
		searchCondition = selectKeyValues(mi, ind, vals[0])
	}

	v, e := orm.db.SelectValue(mi.Table, searchCondition, whereCondition, sqlCondition)
	if e != nil {
		return nil, e
	}
	return combineModelWithKeyValues(mi, ind, v)
}
