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
	return orm.db.RegistNewTable(tables)
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

	val, err := updateKeyValues(mi, ind, vals...)
	if err != nil {
		return err
	}

	return orm.db.UpdateValue(mi.Table, val)
}

func (orm *Orm) UpdateByCondition(md interface{}, val ...[]*ModelTableFieldConditionInfo) error {
	mi, ind := orm.getModelInfoAndIndtype(md)
	v, e := updateConditionKeyValues(mi, ind, val...)
	if e != nil {
		return e
	}
	return orm.db.UpdateValue(mi.Table, v)
}

// 删除表字段,如果没有传入val 则默认删除主键
func (orm *Orm) Delete(md interface{}, val ...[]*ModelTableFieldConditionInfo) (int64, error) {
	mi, ind := orm.getModelInfoAndIndtype(md)
	v, e := deleteKeyValues(mi, ind, val...)
	if e != nil {
		return 0, e
	}
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
	var err error
	if len(vals) == 3 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
		whereCondition, err = selectKeyValues(mi, ind, vals[1])
		if err != nil {
			return nil, err
		}
		sqlCondition = vals[2]
	} else if len(vals) == 2 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
		whereCondition, err = selectKeyValues(mi, ind, vals[1])
		if err != nil {
			return nil, err
		}
	} else if len(vals) == 1 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
	}

	v, e := orm.db.SelectValue(mi.Table, searchCondition, whereCondition, sqlCondition)
	if e != nil {
		return nil, e
	}
	return combineModelWithKeyValues(mi, ind, v)
}

// 读取表数据
// 第1个val代表查询的where的条件
// 第2个val中的数据，key可以是（ORDER BY, LIMIT , ） value(a desc,...)
// 不会默认主键查询，默认全表查询
func (orm *Orm) SelectByCondition(md interface{}, vals ...[]*ModelTableFieldConditionInfo) ([]interface{}, error) {

	mi, ind := orm.getModelInfoAndIndtype(md)
	var (
		whereCondition []*ModelTableFieldConditionInfo
		sqlCondition   []*ModelTableFieldConditionInfo
	)
	var err error
	if len(vals) == 2 {
		whereCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
		sqlCondition = vals[1]
	} else if len(vals) == 1 {
		whereCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
	}

	v, e := orm.db.SelectValue(mi.Table, nil, whereCondition, sqlCondition)
	if e != nil {
		return nil, e
	}
	return combineModelWithKeyValues(mi, ind, v)
}
